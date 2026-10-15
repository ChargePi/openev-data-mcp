package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// searchVehiclesOutput is the structured content returned by search_vehicles.
// MCP requires structured content to be an object, so the list is wrapped.
type searchVehiclesOutput struct {
	Vehicles []vehicle.Vehicle `json:"vehicles"`
}

// searchVehiclesOutputSchema derives the output schema from searchVehiclesOutput,
// overriding types whose default translation does not match their JSON encoding.
func searchVehiclesOutputSchema() (json.RawMessage, error) {
	schema, err := jsonschema.For[searchVehiclesOutput](&jsonschema.ForOptions{
		TypeSchemas: map[reflect.Type]*jsonschema.Schema{
			reflect.TypeFor[time.Time](): {Type: "string", Format: "date-time"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("inferring search_vehicles output schema: %w", err)
	}

	// The MCP spec requires the output schema root to be an object.
	schema.Type, schema.Types = "object", nil
	return json.Marshal(schema)
}

func (s *Server) registerTools() error {
	outputSchema, err := searchVehiclesOutputSchema()
	if err != nil {
		return err
	}

	s.mcp.AddTool(
		mcp.NewTool("search_vehicles",
			mcp.WithDescription("Semantic search over electric vehicles. Describe the vehicle you are looking for in natural language "+
				"(e.g. 'AWD SUV with 800V fast charging and over 500 km WLTP range') and get the best matching vehicles, most relevant first. "+
				"Returns an empty list when nothing is relevant enough."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Natural language description of the electric vehicle you are looking for"),
			),
			mcp.WithRawOutputSchema(outputSchema),
			mcp.WithReadOnlyHintAnnotation(true),
		),
		s.handleSearchVehicles,
	)
	return nil
}

func (s *Server) handleSearchVehicles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return mcp.NewToolResultError("query must not be empty"), nil
	}

	s.logger.Info("tool call", zap.String("tool", "search_vehicles"), zap.String("query", query))

	vehicles, err := s.searcher.SearchVehicles(ctx, query)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("searching vehicles failed", err), nil
	}

	output := searchVehiclesOutput{Vehicles: vehicles}
	text, err := json.Marshal(output)
	if err != nil {
		s.logger.Error("failed to marshal search results", zap.Error(err))
		return mcp.NewToolResultErrorFromErr("marshaling search results failed", err), nil
	}

	// The text block mirrors the structured content for clients that don't support it.
	return mcp.NewToolResultStructured(output, string(text)), nil
}
