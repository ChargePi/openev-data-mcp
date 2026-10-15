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
		Order("make_slug, model_slug, year, trim_slug").
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("querying vehicles: %w", result.Error)
	}

	return toDomain(models)
}

func (r *Repository) GetVehicle(ctx context.Context, uniqueCode string) (*vehicle.Vehicle, error) {
	var m VehicleModel
	result := r.db.WithContext(ctx).
		Where("unique_code = ?", uniqueCode).
		First(&m)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("querying vehicle %q: %w", uniqueCode, result.Error)
	}

	v, err := m.ToDomain()
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Repository) ListMakes(ctx context.Context) ([]vehicle.NamedEntity, error) {
	var rows []makeRow
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
		Where("make_slug = ?", makeSlug).
		Order("model_slug, year, trim_slug").
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("querying vehicles by make %q: %w", makeSlug, result.Error)
	}

	return toDomain(models)
}

func (r *Repository) GetVehiclesByUniqueCodes(ctx context.Context, codes []string) ([]vehicle.Vehicle, error) {
	var models []VehicleModel
	result := r.db.WithContext(ctx).
		Where("unique_code IN ?", codes).
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("querying vehicles by unique codes: %w", result.Error)
	}

	return toDomain(models)
}

func toDomain(models []VehicleModel) ([]vehicle.Vehicle, error) {
	vehicles := make([]vehicle.Vehicle, len(models))
	for i, m := range models {
		v, err := m.ToDomain()
		if err != nil {
			return nil, err
		}
		vehicles[i] = v
	}
	return vehicles, nil
}
