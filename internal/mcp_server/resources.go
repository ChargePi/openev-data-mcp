package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

func (s *Server) registerResources() {
	s.mcp.AddResource(
		mcp.NewResource("evdata://vehicles",
			"EV Vehicles",
			mcp.WithResourceDescription("List of all electric vehicles in the dataset"),
			mcp.WithMIMEType("application/json"),
		),
		s.handleVehiclesList,
	)

	s.mcp.AddResource(
		mcp.NewResource("evdata://makes",
			"EV Makes",
			mcp.WithResourceDescription("List of all electric vehicle manufacturers"),
			mcp.WithMIMEType("application/json"),
		),
		s.handleMakesList,
	)

	s.mcp.AddResourceTemplate(
		// Reserved expansion ({+...}) so the colons in unique codes match unencoded.
		mcp.NewResourceTemplate("evdata://vehicles/{+unique_code}",
			"EV Vehicle Details",
			mcp.WithTemplateDescription("Full details for a specific electric vehicle by its unique code "+
				"('make:model:year:trim' slugs, e.g. 'audi:q6_e_tron:2025:q6_e_tron')"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		s.handleVehicleDetail,
	)

	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate("evdata://makes/{make}/vehicles",
			"Vehicles by Make",
			mcp.WithTemplateDescription("All vehicles from a specific manufacturer (use make slug, e.g. 'tesla')"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		s.handleVehiclesByMake,
	)
}

func (s *Server) handleVehiclesList(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Info("resource read", zap.String("uri", req.Params.URI))

	vehicles, err := s.svc.ListVehicles(ctx)
	if err != nil {
		s.logger.Error("failed to list vehicles", zap.Error(err))
		return nil, fmt.Errorf("listing vehicles: %w", err)
	}

	contents := make([]mcp.ResourceContents, 0, len(vehicles))
	for _, v := range vehicles {
		data, err := json.Marshal(v)
		if err != nil {
			s.logger.Error("failed to marshal vehicle", zap.String("unique_code", v.UniqueCode), zap.Error(err))
			return nil, fmt.Errorf("marshaling vehicle %q: %w", v.UniqueCode, err)
		}
		contents = append(contents, mcp.TextResourceContents{
			URI:      "evdata://vehicles/" + v.UniqueCode,
			MIMEType: "application/json",
			Text:     string(data),
		})
	}

	s.logger.Debug("vehicles list served", zap.Int("count", len(vehicles)))
	return contents, nil
}

func (s *Server) handleMakesList(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Info("resource read", zap.String("uri", req.Params.URI))

	makes, err := s.svc.ListMakes(ctx)
	if err != nil {
		s.logger.Error("failed to list makes", zap.Error(err))
		return nil, fmt.Errorf("listing makes: %w", err)
	}

	contents := make([]mcp.ResourceContents, 0, len(makes))
	for _, m := range makes {
		data, err := json.Marshal(m)
		if err != nil {
			s.logger.Error("failed to marshal make", zap.String("slug", m.Slug), zap.Error(err))
			return nil, fmt.Errorf("marshaling make %q: %w", m.Slug, err)
		}
		contents = append(contents, mcp.TextResourceContents{
			URI:      fmt.Sprintf("evdata://makes/%s", m.Slug),
			MIMEType: "application/json",
			Text:     string(data),
		})
	}

	s.logger.Debug("makes list served", zap.Int("count", len(makes)))
	return contents, nil
}

func (s *Server) handleVehicleDetail(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Info("resource read", zap.String("uri", req.Params.URI))

	raw := strings.TrimPrefix(req.Params.URI, "evdata://vehicles/")
	// Clients may percent-encode the colons in the unique code.
	code, err := url.PathUnescape(raw)
	if err != nil {
		s.logger.Warn("invalid vehicle unique code", zap.String("raw", raw), zap.Error(err))
		return nil, fmt.Errorf("invalid vehicle unique code %q: %w", raw, err)
	}

	v, err := s.svc.GetVehicle(ctx, code)
	if err != nil {
		s.logger.Error("failed to get vehicle", zap.String("unique_code", code), zap.Error(err))
		return nil, fmt.Errorf("getting vehicle %q: %w", code, err)
	}
	if v == nil {
		s.logger.Warn("vehicle not found", zap.String("unique_code", code))
		return nil, fmt.Errorf("vehicle %q not found", code)
	}

	data, err := json.Marshal(v)
	if err != nil {
		s.logger.Error("failed to marshal vehicle", zap.String("unique_code", code), zap.Error(err))
		return nil, fmt.Errorf("marshaling vehicle: %w", err)
	}

	s.logger.Debug("vehicle detail served", zap.String("unique_code", code))
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func (s *Server) handleVehiclesByMake(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	s.logger.Info("resource read", zap.String("uri", req.Params.URI))

	slug := strings.TrimSuffix(strings.TrimPrefix(req.Params.URI, "evdata://makes/"), "/vehicles")

	vehicles, err := s.svc.GetVehiclesByMake(ctx, slug)
	if err != nil {
		s.logger.Error("failed to get vehicles by make", zap.String("make", slug), zap.Error(err))
		return nil, fmt.Errorf("getting vehicles by make %q: %w", slug, err)
	}

	contents := make([]mcp.ResourceContents, 0, len(vehicles))
	for _, v := range vehicles {
		data, err := json.Marshal(v)
		if err != nil {
			s.logger.Error("failed to marshal vehicle", zap.String("unique_code", v.UniqueCode), zap.Error(err))
			return nil, fmt.Errorf("marshaling vehicle %q: %w", v.UniqueCode, err)
		}
		contents = append(contents, mcp.TextResourceContents{
			URI:      "evdata://vehicles/" + v.UniqueCode,
			MIMEType: "application/json",
			Text:     string(data),
		})
	}

	s.logger.Debug("vehicles by make served", zap.String("make", slug), zap.Int("count", len(vehicles)))
	return contents, nil
}
