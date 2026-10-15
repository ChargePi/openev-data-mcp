package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
)

// VehicleModel maps to the vehicles table from the open-ev-data dataset schema.
type VehicleModel struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	UniqueCode  string `gorm:"uniqueIndex;not null"`
	MakeSlug    string `gorm:"not null;index"`
	MakeName    string `gorm:"not null"`
	ModelSlug   string `gorm:"not null;index"`
	ModelName   string `gorm:"not null"`
	Year        int    `gorm:"not null;index"`
	TrimSlug    string `gorm:"not null"`
	TrimName    string `gorm:"not null"`
	VariantSlug string
	VariantName string
	VehicleType string `gorm:"not null;index"`
	Drivetrain  string `gorm:"not null"`

	SystemPowerKW  *float64
	SystemTorqueNm *float64

	BatteryCapacityGrossKWh *float64 `gorm:"column:battery_capacity_gross_kwh"`
	BatteryCapacityNetKWh   *float64 `gorm:"column:battery_capacity_net_kwh"`
	BatteryChemistry        string

	DCMaxPowerKW *float64
	ACMaxPowerKW *float64

	RangeWLTPKm *float64
	RangeEPAKm  *float64

	Acceleration0To100S *float64 `gorm:"column:acceleration_0_100_s"`
	TopSpeedKmh         *float64

	JsonData  json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time
}

func (VehicleModel) TableName() string { return "vehicles" }

// makeRow is the projection of distinct makes from the vehicles table.
type makeRow struct {
	MakeSlug string
	MakeName string
}

// ToDomain decodes the canonical OpenEV Data record stored in json_data. The
// other columns are denormalized copies used for querying and ordering.
func (m VehicleModel) ToDomain() (vehicle.Vehicle, error) {
	var v vehicle.Vehicle
	if err := json.Unmarshal(m.JsonData, &v); err != nil {
		return vehicle.Vehicle{}, fmt.Errorf("parsing data of vehicle %q: %w", m.UniqueCode, err)
	}
	v.ID = m.ID
	return v, nil
}
