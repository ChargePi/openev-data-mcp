package service

import (
	"context"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	"go.uber.org/zap"
)

// Repository is the data-access interface the service depends on.
type Repository interface {
	ListVehicles(ctx context.Context) ([]vehicle.Vehicle, error)
	GetVehicle(ctx context.Context, id int) (*vehicle.Vehicle, error)
	ListMakes(ctx context.Context) ([]vehicle.NamedEntity, error)
	GetVehiclesByMake(ctx context.Context, makeSlug string) ([]vehicle.Vehicle, error)
}

type VehicleService struct {
	repo   Repository
	logger *zap.Logger
}

func NewVehicleService(repo Repository, logger *zap.Logger) *VehicleService {
	return &VehicleService{repo: repo, logger: logger}
}

func (s *VehicleService) ListVehicles(ctx context.Context) ([]vehicle.Vehicle, error) {
	s.logger.Debug("listing vehicles")
	vehicles, err := s.repo.ListVehicles(ctx)
	if err != nil {
		s.logger.Error("failed to list vehicles", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("listed vehicles", zap.Int("count", len(vehicles)))
	return vehicles, nil
}

func (s *VehicleService) GetVehicle(ctx context.Context, id int) (*vehicle.Vehicle, error) {
	s.logger.Debug("getting vehicle", zap.Int("id", id))
	v, err := s.repo.GetVehicle(ctx, id)
	if err != nil {
		s.logger.Error("failed to get vehicle", zap.Int("id", id), zap.Error(err))
		return nil, err
	}
	return v, nil
}

func (s *VehicleService) ListMakes(ctx context.Context) ([]vehicle.NamedEntity, error) {
	s.logger.Debug("listing makes")
	makes, err := s.repo.ListMakes(ctx)
	if err != nil {
		s.logger.Error("failed to list makes", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("listed makes", zap.Int("count", len(makes)))
	return makes, nil
}

func (s *VehicleService) GetVehiclesByMake(ctx context.Context, makeSlug string) ([]vehicle.Vehicle, error) {
	s.logger.Debug("getting vehicles by make", zap.String("make", makeSlug))
	vehicles, err := s.repo.GetVehiclesByMake(ctx, makeSlug)
	if err != nil {
		s.logger.Error("failed to get vehicles by make", zap.String("make", makeSlug), zap.Error(err))
		return nil, err
	}
	s.logger.Debug("got vehicles by make", zap.String("make", makeSlug), zap.Int("count", len(vehicles)))
	return vehicles, nil
}