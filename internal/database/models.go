package database

import (
	"encoding/json"
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

	BatteryCapacityGrossKWh *float64
	BatteryCapacityNetKWh   *float64
	BatteryChemistry        string

	DCMaxPowerKW *float64
	ACMaxPowerKW *float64

	RangeWLTPKm *float64
	RangeEPAKm  *float64

	Acceleration0To100S *float64
	TopSpeedKmh         *float64

	JsonData  json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time

	ChargePorts  []ChargePortModel  `gorm:"foreignKey:VehicleID;constraint:OnDelete:CASCADE"`
	RangeRatings []RangeRatingModel `gorm:"foreignKey:VehicleID;constraint:OnDelete:CASCADE"`
	Sources      []SourceModel      `gorm:"foreignKey:VehicleID;constraint:OnDelete:CASCADE"`
}

func (VehicleModel) TableName() string { return "vehicles" }

func (m VehicleModel) ToDomain() vehicle.Vehicle {
	v := vehicle.Vehicle{
		ID:          m.ID,
		UniqueCode:  m.UniqueCode,
		Make:        vehicle.NamedEntity{Slug: m.MakeSlug, Name: m.MakeName},
		Model:       vehicle.NamedEntity{Slug: m.ModelSlug, Name: m.ModelName},
		Year:        m.Year,
		Trim:        vehicle.NamedEntity{Slug: m.TrimSlug, Name: m.TrimName},
		VehicleType: m.VehicleType,
		Drivetrain:  m.Drivetrain,

		SystemPowerKW:  m.SystemPowerKW,
		SystemTorqueNm: m.SystemTorqueNm,

		Acceleration0To100S: m.Acceleration0To100S,
		TopSpeedKmh:         m.TopSpeedKmh,

		BatteryCapacityGrossKWh: m.BatteryCapacityGrossKWh,
		BatteryCapacityNetKWh:   m.BatteryCapacityNetKWh,
		BatteryChemistry:        m.BatteryChemistry,

		DCMaxPowerKW: m.DCMaxPowerKW,
		ACMaxPowerKW: m.ACMaxPowerKW,

		RangeWLTPKm: m.RangeWLTPKm,
		RangeEPAKm:  m.RangeEPAKm,

		JsonData: m.JsonData,
	}

	if m.VariantSlug != "" {
		v.Variant = &vehicle.NamedEntity{Slug: m.VariantSlug, Name: m.VariantName}
	}

	for _, p := range m.ChargePorts {
		v.ChargePorts = append(v.ChargePorts, p.ToDomain())
	}
	for _, rr := range m.RangeRatings {
		v.RangeRatings = append(v.RangeRatings, rr.ToDomain())
	}
	for _, s := range m.Sources {
		v.Sources = append(v.Sources, s.ToDomain())
	}

	return v
}

type ChargePortModel struct {
	ID               int    `gorm:"primaryKey;autoIncrement"`
	VehicleID        int    `gorm:"not null;index"`
	Kind             string `gorm:"not null"`
	Connector        string `gorm:"not null"`
	LocationSide     string
	LocationPosition string
}

func (ChargePortModel) TableName() string { return "charge_ports" }

func (m ChargePortModel) ToDomain() vehicle.ChargePort {
	return vehicle.ChargePort{
		Kind:             m.Kind,
		Connector:        m.Connector,
		LocationSide:     m.LocationSide,
		LocationPosition: m.LocationPosition,
	}
}

type RangeRatingModel struct {
	ID        int     `gorm:"primaryKey;autoIncrement"`
	VehicleID int     `gorm:"not null;index"`
	Cycle     string  `gorm:"not null"`
	RangeKm   float64 `gorm:"not null"`
	Notes     string
}

func (RangeRatingModel) TableName() string { return "range_ratings" }

func (m RangeRatingModel) ToDomain() vehicle.RangeRating {
	return vehicle.RangeRating{
		Cycle:   m.Cycle,
		RangeKm: m.RangeKm,
		Notes:   m.Notes,
	}
}

type SourceModel struct {
	ID         int       `gorm:"primaryKey;autoIncrement"`
	VehicleID  int       `gorm:"not null;index"`
	SourceType string    `gorm:"not null"`
	Title      string    `gorm:"not null"`
	URL        string    `gorm:"not null"`
	AccessedAt time.Time `gorm:"not null"`
	Publisher  string
}

func (SourceModel) TableName() string { return "sources" }

func (m SourceModel) ToDomain() vehicle.Source {
	return vehicle.Source{
		Type:       m.SourceType,
		Title:      m.Title,
		Publisher:  m.Publisher,
		URL:        m.URL,
		AccessedAt: m.AccessedAt,
	}
}
