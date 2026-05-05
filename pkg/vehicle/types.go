package vehicle

import (
	"encoding/json"
	"time"
)

type NamedEntity struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type ChargePort struct {
	Kind             string `json:"kind"`
	Connector        string `json:"connector"`
	LocationSide     string `json:"location_side,omitempty"`
	LocationPosition string `json:"location_position,omitempty"`
}

type RangeRating struct {
	Cycle   string  `json:"cycle"`
	RangeKm float64 `json:"range_km"`
	Notes   string  `json:"notes,omitempty"`
}

type Source struct {
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Publisher  string    `json:"publisher,omitempty"`
	URL        string    `json:"url,omitempty"`
	AccessedAt time.Time `json:"accessed_at,omitempty"`
}

type Vehicle struct {
	ID          int          `json:"id"`
	UniqueCode  string       `json:"unique_code,omitempty"`
	Make        NamedEntity  `json:"make"`
	Model       NamedEntity  `json:"model"`
	Year        int          `json:"year"`
	Trim        NamedEntity  `json:"trim"`
	Variant     *NamedEntity `json:"variant,omitempty"`
	VehicleType string       `json:"vehicle_type"`
	Drivetrain  string       `json:"drivetrain,omitempty"`

	SystemPowerKW  *float64 `json:"system_power_kw,omitempty"`
	SystemTorqueNm *float64 `json:"system_torque_nm,omitempty"`

	Acceleration0To100S *float64 `json:"acceleration_0_100_s,omitempty"`
	TopSpeedKmh         *float64 `json:"top_speed_kmh,omitempty"`

	BatteryCapacityGrossKWh *float64 `json:"battery_capacity_gross_kwh,omitempty"`
	BatteryCapacityNetKWh   *float64 `json:"battery_capacity_net_kwh,omitempty"`
	BatteryChemistry        string   `json:"battery_chemistry,omitempty"`

	DCMaxPowerKW *float64 `json:"dc_max_power_kw,omitempty"`
	ACMaxPowerKW *float64 `json:"ac_max_power_kw,omitempty"`

	RangeWLTPKm *float64 `json:"range_wltp_km,omitempty"`
	RangeEPAKm  *float64 `json:"range_epa_km,omitempty"`

	// JsonData holds the full canonical vehicle record from the dataset.
	JsonData json.RawMessage `json:"data,omitempty"`

	ChargePorts  []ChargePort  `json:"charge_ports,omitempty"`
	RangeRatings []RangeRating `json:"range_ratings,omitempty"`
	Sources      []Source      `json:"sources,omitempty"`
}

type Stats struct {
	TotalVehicles int            `json:"total_vehicles"`
	MakeCount     int            `json:"make_count"`
	YearRange     [2]int         `json:"year_range"`
	ByType        map[string]int `json:"by_type"`
	ByMake        map[string]int `json:"by_make"`
}
