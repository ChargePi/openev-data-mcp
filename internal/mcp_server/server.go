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
	GetVehicle(ctx context.Context, id int) (*vehicle.Vehicle, error)
	ListMakes(ctx context.Context) ([]vehicle.NamedEntity, error)
	GetVehiclesByMake(ctx context.Context, makeSlug string) ([]vehicle.Vehicle, error)
}

type Server struct {
	mcp    *mcpserver.MCPServer
	svc    VehicleService
	logger *zap.Logger
}

func New(svc VehicleService, cacheTTL time.Duration, logger *zap.Logger) *Server {
	cache := newResourceCache(cacheTTL)

	hooks := &mcpserver.Hooks{}
	mcp.LoggingHooks(logger, hooks)
	mcp.TraceHooks(hooks)

	s := mcpserver.NewMCPServer(mcpName, mcpVersion,
		mcpserver.WithRecovery(),
		mcpserver.WithHooks(hooks),
		mcpserver.WithResourceHandlerMiddleware(cache.Middleware),
	)
	srv := &Server{mcp: s, svc: svc, logger: logger}
	srv.registerResources()
	return srv
}

func (s *Server) Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)
	handler := otelhttp.NewHandler(mcpserver.NewStreamableHTTPServer(s.mcp), "mcp")
	s.logger.Info("MCP server listening", zap.Int("port", port))
	return http.ListenAndServe(addr, handler)
}
