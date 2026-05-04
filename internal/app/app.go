package app

import (
	"context"
	"fmt"
	"net"
	"slices"
	"sync"

	"github.com/ChargePi/openev-data-mcp/internal/config"
	"github.com/ChargePi/openev-data-mcp/internal/database"
	"github.com/ChargePi/openev-data-mcp/internal/server"
	"github.com/ChargePi/openev-data-mcp/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	serviceName    = "openev-data-mcp"
	serviceVersion = "1.0.0"
)

type App struct {
	deferWaitGroup sync.WaitGroup
	defers         []func(ctx context.Context)
	logger         *zap.Logger
}

func New(logger *zap.Logger) *App {
	return &App{logger: logger}
}

func (a *App) Run(ctx context.Context, cfg *config.Config) {
	logger := a.logger
	logger.Info("starting service",
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)

	db, err := database.Connect(cfg.Database, logger)
	if err != nil {
		logger.Fatal("connecting to database", zap.Error(err))
	}
	a.addDeferFunc(func(_ context.Context) {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	repo := database.NewRepository(db)
	svc := service.NewVehicleService(repo, logger)

	healthSrv := a.startHealthServer(cfg.HealthPort)
	healthSrv.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	srv := server.New(svc, cfg.RefreshInterval, logger)

	go func() {
		if err := srv.Serve(cfg.Port); err != nil {
			logger.Error("MCP server stopped", zap.Error(err))
		}
	}()

	<-ctx.Done()
	healthSrv.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	logger.Info("context cancelled, shutting down")
}

func (a *App) startHealthServer(port int) *health.Server {
	logger := a.logger

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("failed to listen for health check", zap.Int("port", port), zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSrv)

	go func() {
		logger.Info("gRPC health server listening", zap.Int("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC health server stopped", zap.Error(err))
		}
	}()

	a.addDeferFunc(func(_ context.Context) {
		grpcServer.GracefulStop()
	})

	return healthSrv
}

func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("stopping service", zap.String("service", serviceName))

	slices.Reverse(a.defers)

	a.deferWaitGroup.Add(len(a.defers))
	for _, f := range a.defers {
		f(ctx)
		a.deferWaitGroup.Done()
	}
	a.deferWaitGroup.Wait()
}

func (a *App) addDeferFunc(f func(ctx context.Context)) {
	a.defers = append(a.defers, f)
}
