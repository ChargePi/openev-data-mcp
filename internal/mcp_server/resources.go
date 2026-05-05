package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
		mcp.NewResourceTemplate("evdata://vehicles/{id}",
			"EV Vehicle Details",
			mcp.WithTemplateDescription("Full details for a specific electric vehicle by numeric ID"),
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
			s.logger.Error("failed to marshal vehicle", zap.Int("id", v.ID), zap.Error(err))
			return nil, fmt.Errorf("marshaling vehicle %d: %w", v.ID, err)
		}
		contents = append(contents, mcp.TextResourceContents{
			URI:      fmt.Sprintf("evdata://vehicles/%d", v.ID),
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
	id, err := strconv.Atoi(raw)
	if err != nil {
		s.logger.Warn("invalid vehicle ID", zap.String("raw", raw), zap.Error(err))
		return nil, fmt.Errorf("invalid vehicle ID %q: %w", raw, err)
	}

	v, err := s.svc.GetVehicle(ctx, id)
	if err != nil {
		s.logger.Error("failed to get vehicle", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("getting vehicle %d: %w", id, err)
	}
	if v == nil {
		s.logger.Warn("vehicle not found", zap.Int("id", id))
		return nil, fmt.Errorf("vehicle %d not found", id)
	}

	data, err := json.Marshal(v)
	if err != nil {
		s.logger.Error("failed to marshal vehicle", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("marshaling vehicle: %w", err)
	}

	s.logger.Debug("vehicle detail served", zap.Int("id", id))
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
			s.logger.Error("failed to marshal vehicle", zap.Int("id", v.ID), zap.Error(err))
			return nil, fmt.Errorf("marshaling vehicle %d: %w", v.ID, err)
		}
		contents = append(contents, mcp.TextResourceContents{
			URI:      fmt.Sprintf("evdata://vehicles/%d", v.ID),
			MIMEType: "application/json",
			Text:     string(data),
		})
	}

	s.logger.Debug("vehicles by make served", zap.String("make", slug), zap.Int("count", len(vehicles)))
	return contents, nil
}
