package database

import (
	"context"
	"fmt"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListVehicles(ctx context.Context) ([]vehicle.Vehicle, error) {
	var models []VehicleModel
	result := r.db.WithContext(ctx).
		Preload("ChargePorts").
		Preload("RangeRatings").
		Preload("Sources").
		Order("make_slug, model_slug, year, trim_slug").
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("querying vehicles: %w", result.Error)
	}

	vehicles := make([]vehicle.Vehicle, len(models))
	for i, m := range models {
		vehicles[i] = m.ToDomain()
	}
	return vehicles, nil
}

func (r *Repository) GetVehicle(ctx context.Context, id int) (*vehicle.Vehicle, error) {
	var m VehicleModel
	result := r.db.WithContext(ctx).
		Preload("ChargePorts").
		Preload("RangeRatings").
		Preload("Sources").
		First(&m, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("querying vehicle %d: %w", id, result.Error)
	}

	v := m.ToDomain()
	return &v, nil
}

func (r *Repository) ListMakes(ctx context.Context) ([]vehicle.NamedEntity, error) {
	var rows []struct {
		MakeSlug string
		MakeName string
	}
	result := r.db.WithContext(ctx).
		Model(&VehicleModel{}).
		Select("DISTINCT make_slug, make_name").
		Order("make_slug").
		Scan(&rows)
	if result.Error != nil {
		return nil, fmt.Errorf("querying makes: %w", result.Error)
	}

	makes := make([]vehicle.NamedEntity, len(rows))
	for i, row := range rows {
		makes[i] = vehicle.NamedEntity{Slug: row.MakeSlug, Name: row.MakeName}
	}
	return makes, nil
}

func (r *Repository) GetVehiclesByMake(ctx context.Context, makeSlug string) ([]vehicle.Vehicle, error) {
	var models []VehicleModel
	result := r.db.WithContext(ctx).
		Preload("ChargePorts").
		Preload("RangeRatings").
		Preload("Sources").
		Where("make_slug = ?", makeSlug).
		Order("model_slug, year, trim_slug").
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("querying vehicles by make %q: %w", makeSlug, result.Error)
	}

	vehicles := make([]vehicle.Vehicle, len(models))
	for i, m := range models {
		vehicles[i] = m.ToDomain()
	}
	return vehicles, nil
}
