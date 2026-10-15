package mcp_server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ChargePi/openev-data-mcp/pkg/observability/mcp"
	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

const (
	mcpName    = "openev-data-mcp"
	mcpVersion = "1.0.0"
)

// VehicleService is the read interface the server depends on.
type VehicleService interface {
	ListVehicles(ctx context.Context) ([]vehicle.Vehicle, error)
	GetVehicle(ctx context.Context, uniqueCode string) (*vehicle.Vehicle, error)
	ListMakes(ctx context.Context) ([]vehicle.NamedEntity, error)
	GetVehiclesByMake(ctx context.Context, makeSlug string) ([]vehicle.Vehicle, error)
}

// VehicleSearcher is the semantic search interface the server depends on.
type VehicleSearcher interface {
	SearchVehicles(ctx context.Context, query string) ([]vehicle.Vehicle, error)
}

type Server struct {
	mcp      *mcpserver.MCPServer
	svc      VehicleService
	searcher VehicleSearcher
	logger   *zap.Logger
}

func New(svc VehicleService, searcher VehicleSearcher, cacheTTL time.Duration, logger *zap.Logger) (*Server, error) {
	cache := newResourceCache(cacheTTL)

	hooks := &mcpserver.Hooks{}
	mcp.LoggingHooks(logger, hooks)
	mcp.TraceHooks(hooks)

	s := mcpserver.NewMCPServer(mcpName, mcpVersion,
		mcpserver.WithRecovery(),
		mcpserver.WithHooks(hooks),
		mcpserver.WithResourceHandlerMiddleware(cache.Middleware),
		mcpserver.WithToolCapabilities(false),
	)
	srv := &Server{mcp: s, svc: svc, searcher: searcher, logger: logger}
	srv.registerResources()
	if err := srv.registerTools(); err != nil {
		return nil, fmt.Errorf("registering tools: %w", err)
	}
	return srv, nil
}

func (s *Server) Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)
	handler := otelhttp.NewHandler(mcpserver.NewStreamableHTTPServer(s.mcp), "mcp")
	s.logger.Info("MCP server listening", zap.Int("port", port))
	return http.ListenAndServe(addr, handler)
}
