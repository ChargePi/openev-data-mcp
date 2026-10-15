package vehicle

import "time"

// The types in this file mirror the OpenEV Data vehicle record schema
// (https://github.com/open-ev-data/open-ev-data-dataset/blob/main/schema.json),
// so a marshaled Vehicle is a valid record. The schema's "$schema" path is
// intentionally not modelled.

// NamedEntity is the schema's slug_name definition.
type NamedEntity struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Vehicle is an OpenEV Data vehicle record.
type Vehicle struct {
	// ID is the database key. It is not part of the record schema.
	ID int `json:"-"`

	SchemaVersion string        `json:"schema_version"`
	UniqueCode    string        `json:"unique_code,omitempty"`
	Make          NamedEntity   `json:"make"`
	Model         NamedEntity   `json:"model"`
	Year          int           `json:"year"`
	Trim          NamedEntity   `json:"trim"`
	Variant       *Variant      `json:"variant,omitempty"`
	Markets       []string      `json:"markets,omitempty"`
	Availability  *Availability `json:"availability,omitempty"`
	VehicleType   string        `json:"vehicle_type"`
	Body          *Body         `json:"body,omitempty"`
	Dimensions    *Dimensions   `json:"dimensions,omitempty"`
	Weights       *Weights      `json:"weights,omitempty"`
	Capacity      *Capacity     `json:"capacity,omitempty"`
	Powertrain    Powertrain    `json:"powertrain"`
	Battery       Battery       `json:"battery"`
	ChargePorts   []ChargePort  `json:"charge_ports"`
	Charging      Charging      `json:"charging"`
	V2X           *V2X          `json:"v2x,omitempty"`
	Range         Range         `json:"range"`
	Efficiency    *Efficiency   `json:"efficiency,omitempty"`
	Performance   *Performance  `json:"performance,omitempty"`
	WheelsTires   *WheelsTires  `json:"wheels_tires,omitempty"`
	Pricing       *Pricing      `json:"pricing,omitempty"`
	Software      *Software     `json:"software,omitempty"`
	Links         *Links        `json:"links,omitempty"`
	Images        *Images       `json:"images,omitempty"`
	Sources       []Source      `json:"sources"`
	Metadata      *Metadata     `json:"metadata,omitempty"`
}

type Variant struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Kind  string `json:"kind,omitempty"`
	Notes string `json:"notes,omitempty"`
}

type Availability struct {
	Status    string `json:"status"`
	StartYear *int   `json:"start_year,omitempty"`
	EndYear   *int   `json:"end_year,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

type Body struct {
	Style             string   `json:"style,omitempty"`
	Doors             *int     `json:"doors,omitempty"`
	Seats             *int     `json:"seats,omitempty"`
	Rows              *int     `json:"rows,omitempty"`
	Platform          string   `json:"platform,omitempty"`
	DragCoefficientCd *float64 `json:"drag_coefficient_cd,omitempty"`
}

type Dimensions struct {
	LengthMm           *float64 `json:"length_mm,omitempty"`
	WidthMm            *float64 `json:"width_mm,omitempty"`
	WidthWithMirrorsMm *float64 `json:"width_with_mirrors_mm,omitempty"`
	HeightMm           *float64 `json:"height_mm,omitempty"`
	WheelbaseMm        *float64 `json:"wheelbase_mm,omitempty"`
	GroundClearanceMm  *float64 `json:"ground_clearance_mm,omitempty"`
	TurningCircleM     *float64 `json:"turning_circle_m,omitempty"`
}

type Weights struct {
	CurbWeightKg         *float64 `json:"curb_weight_kg,omitempty"`
	GrossVehicleWeightKg *float64 `json:"gross_vehicle_weight_kg,omitempty"`
	MaxPayloadKg         *float64 `json:"max_payload_kg,omitempty"`
	RoofLoadKg           *float64 `json:"roof_load_kg,omitempty"`
}

type Capacity struct {
	CargoL               *float64 `json:"cargo_l,omitempty"`
	CargoMaxL            *float64 `json:"cargo_max_l,omitempty"`
	FrunkL               *float64 `json:"frunk_l,omitempty"`
	TowingBrakedKg       *float64 `json:"towing_braked_kg,omitempty"`
	TowingUnbrakedKg     *float64 `json:"towing_unbraked_kg,omitempty"`
	TowingVerticalLoadKg *float64 `json:"towing_vertical_load_kg,omitempty"`
}

type Powertrain struct {
	Drivetrain     string        `json:"drivetrain"`
	SystemPowerKW  *float64      `json:"system_power_kw,omitempty"`
	SystemTorqueNm *float64      `json:"system_torque_nm,omitempty"`
	Motors         []Motor       `json:"motors,omitempty"`
	Transmission   *Transmission `json:"transmission,omitempty"`
}

type Motor struct {
	Position string   `json:"position"`
	Type     string   `json:"type,omitempty"`
	PowerKW  *float64 `json:"power_kw,omitempty"`
	TorqueNm *float64 `json:"torque_nm,omitempty"`
	Cooling  string   `json:"cooling,omitempty"`
}

type Transmission struct {
	Gears *int   `json:"gears,omitempty"`
	Type  string `json:"type,omitempty"`
}

type Battery struct {
	Manufacturer           string                  `json:"manufacturer,omitempty"`
	Chemistry              string                  `json:"chemistry,omitempty"`
	CathodeMaterial        string                  `json:"cathode_material,omitempty"`
	PackCapacityKWhGross   *float64                `json:"pack_capacity_kwh_gross,omitempty"`
	PackCapacityKWhNet     *float64                `json:"pack_capacity_kwh_net,omitempty"`
	PackVoltageNominalV    *float64                `json:"pack_voltage_nominal_v,omitempty"`
	PackVoltageMaxV        *float64                `json:"pack_voltage_max_v,omitempty"`
	PackVoltageMinV        *float64                `json:"pack_voltage_min_v,omitempty"`
	CellCount              *int                    `json:"cell_count,omitempty"`
	ModuleCount            *int                    `json:"module_count,omitempty"`
	ThermalManagement      string                  `json:"thermal_management,omitempty"`
	HeatPump               *bool                   `json:"heat_pump,omitempty"`
	Preconditioning        *BatteryPreconditioning `json:"preconditioning,omitempty"`
	Warranty               *BatteryWarranty        `json:"warranty,omitempty"`
	UsableSOCWindowPercent *SOCWindow              `json:"usable_soc_window_percent,omitempty"`
}

type BatteryPreconditioning struct {
	Supported bool     `json:"supported"`
	Modes     []string `json:"modes,omitempty"`
	Notes     string   `json:"notes,omitempty"`
}

type BatteryWarranty struct {
	Years                    *int     `json:"years,omitempty"`
	DistanceKm               *int     `json:"distance_km,omitempty"`
	CapacityRetentionPercent *float64 `json:"capacity_retention_percent,omitempty"`
}

type SOCWindow struct {
	MinPercent *float64 `json:"min_percent,omitempty"`
	MaxPercent *float64 `json:"max_percent,omitempty"`
	Notes      string   `json:"notes,omitempty"`
}

type ChargePort struct {
	Kind      string              `json:"kind"`
	Connector string              `json:"connector"`
	Location  *ChargePortLocation `json:"location,omitempty"`
	Covers    string              `json:"covers,omitempty"`
	Light     *bool               `json:"light,omitempty"`
	Motorized *bool               `json:"motorized,omitempty"`
	Notes     string              `json:"notes,omitempty"`
}

type ChargePortLocation struct {
	Side     string `json:"side,omitempty"`
	Position string `json:"position,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type Charging struct {
	AC            *ACCharging        `json:"ac,omitempty"`
	DC            *DCCharging        `json:"dc,omitempty"`
	Protocols     *ChargingProtocols `json:"protocols,omitempty"`
	DCChargeCurve *ChargeCurve       `json:"dc_charge_curve,omitempty"`
	ChargingTime  *ChargingTimes     `json:"charging_time,omitempty"`
}

type ACCharging struct {
	MaxPowerKW            float64       `json:"max_power_kw"`
	SupportedPowerStepsKW []float64     `json:"supported_power_steps_kw,omitempty"`
	Phases                *int          `json:"phases,omitempty"`
	VoltageRangeV         *VoltageRange `json:"voltage_range_v,omitempty"`
	FrequencyHz           *float64      `json:"frequency_hz,omitempty"`
	MaxCurrentA           *float64      `json:"max_current_a,omitempty"`
	OnboardChargerCount   *int          `json:"onboard_charger_count,omitempty"`
	Notes                 string        `json:"notes,omitempty"`
}

type DCCharging struct {
	MaxPowerKW               float64        `json:"max_power_kw"`
	VoltageRangeV            *VoltageRange  `json:"voltage_range_v,omitempty"`
	MaxCurrentA              *float64       `json:"max_current_a,omitempty"`
	ArchitectureVoltageClass string         `json:"architecture_voltage_class,omitempty"`
	PowerLimitsByVoltage     []DCPowerLimit `json:"power_limits_by_voltage,omitempty"`
	Notes                    string         `json:"notes,omitempty"`
}

// VoltageRange is the schema's min_max_v definition.
type VoltageRange struct {
	MinV *float64 `json:"min_v,omitempty"`
	MaxV *float64 `json:"max_v,omitempty"`
}

type DCPowerLimit struct {
	VoltageClass string  `json:"voltage_class"`
	MaxPowerKW   float64 `json:"max_power_kw"`
	Notes        string  `json:"notes,omitempty"`
}

type ChargingProtocols struct {
	AC            []string `json:"ac,omitempty"`
	DC            []string `json:"dc,omitempty"`
	PlugAndCharge *bool    `json:"plug_and_charge,omitempty"`
	Notes         string   `json:"notes,omitempty"`
}

type ChargeCurve struct {
	CurveType  string             `json:"curve_type"`
	Points     []ChargeCurvePoint `json:"points"`
	Conditions *Conditions        `json:"conditions,omitempty"`
	SourceURL  string             `json:"source_url,omitempty"`
	Notes      string             `json:"notes,omitempty"`
}

type ChargeCurvePoint struct {
	SOCPercent float64  `json:"soc_percent"`
	PowerKW    *float64 `json:"power_kw,omitempty"`
	CurrentA   *float64 `json:"current_a,omitempty"`
	VoltageV   *float64 `json:"voltage_v,omitempty"`
}

// Conditions is the schema's conditions definition.
type Conditions struct {
	BatteryTempC    *float64 `json:"battery_temp_c,omitempty"`
	AmbientTempC    *float64 `json:"ambient_temp_c,omitempty"`
	Preconditioning *bool    `json:"preconditioning,omitempty"`
	ChargerPowerKW  *float64 `json:"charger_power_kw,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

type ChargingTimes struct {
	AC []ACChargingTime `json:"ac,omitempty"`
	DC []DCChargingTime `json:"dc,omitempty"`
}

type ACChargingTime struct {
	PowerKW        float64     `json:"power_kw"`
	FromSOCPercent float64     `json:"from_soc_percent"`
	ToSOCPercent   float64     `json:"to_soc_percent"`
	TimeMin        float64     `json:"time_min"`
	Conditions     *Conditions `json:"conditions,omitempty"`
	Notes          string      `json:"notes,omitempty"`
}

type DCChargingTime struct {
	ChargerPowerKW float64     `json:"charger_power_kw"`
	FromSOCPercent float64     `json:"from_soc_percent"`
	ToSOCPercent   float64     `json:"to_soc_percent"`
	TimeMin        float64     `json:"time_min"`
	Conditions     *Conditions `json:"conditions,omitempty"`
	Notes          string      `json:"notes,omitempty"`
}

type V2X struct {
	V2L *V2L                   `json:"v2l,omitempty"`
	V2H *BidirectionalCharging `json:"v2h,omitempty"`
	V2G *BidirectionalCharging `json:"v2g,omitempty"`
}

type V2L struct {
	Supported  bool        `json:"supported"`
	MaxPowerKW *float64    `json:"max_power_kw,omitempty"`
	Outlets    []V2LOutlet `json:"outlets,omitempty"`
	Notes      string      `json:"notes,omitempty"`
}

type V2LOutlet struct {
	Kind     string `json:"kind"`
	Count    *int   `json:"count,omitempty"`
	Location string `json:"location,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// BidirectionalCharging describes vehicle-to-home or vehicle-to-grid support.
type BidirectionalCharging struct {
	Supported  bool     `json:"supported"`
	MaxPowerKW *float64 `json:"max_power_kw,omitempty"`
	Connector  string   `json:"connector,omitempty"`
	Protocols  []string `json:"protocols,omitempty"`
	Notes      string   `json:"notes,omitempty"`
}

type Range struct {
	Rated     []RangeRating    `json:"rated"`
	RealWorld []RealWorldRange `json:"real_world,omitempty"`
}

type RangeRating struct {
	Cycle   string  `json:"cycle"`
	RangeKm float64 `json:"range_km"`
	Notes   string  `json:"notes,omitempty"`
}

type RealWorldRange struct {
	Profile    string               `json:"profile"`
	RangeKm    float64              `json:"range_km"`
	Conditions *RealWorldConditions `json:"conditions,omitempty"`
	Notes      string               `json:"notes,omitempty"`
}

type RealWorldConditions struct {
	Weather  string   `json:"weather,omitempty"`
	SpeedKmh *float64 `json:"speed_kmh,omitempty"`
}

type Efficiency struct {
	EnergyConsumptionWhPerKm *float64 `json:"energy_consumption_wh_per_km,omitempty"`
	MPGe                     *float64 `json:"mpge,omitempty"`
	Notes                    string   `json:"notes,omitempty"`
}

type Performance struct {
	Acceleration0To100KmhS *float64 `json:"acceleration_0_100_kmh_s,omitempty"`
	Acceleration0To60MphS  *float64 `json:"acceleration_0_60_mph_s,omitempty"`
	TopSpeedKmh            *float64 `json:"top_speed_kmh,omitempty"`
	QuarterMileS           *float64 `json:"quarter_mile_s,omitempty"`
}

type WheelsTires struct {
	StandardWheelSizeIn    *float64      `json:"standard_wheel_size_in,omitempty"`
	OptionalWheelSizesIn   []float64     `json:"optional_wheel_sizes_in,omitempty"`
	TireSizes              []string      `json:"tire_sizes,omitempty"`
	RecommendedPressureKPa *TirePressure `json:"recommended_pressure_kpa,omitempty"`
	Notes                  string        `json:"notes,omitempty"`
}

type TirePressure struct {
	FrontKPa *float64 `json:"front_kpa,omitempty"`
	RearKPa  *float64 `json:"rear_kpa,omitempty"`
}

type Pricing struct {
	MSRP []Price `json:"msrp,omitempty"`
}

type Price struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Country  string  `json:"country,omitempty"`
	Year     *int    `json:"year,omitempty"`
	Notes    string  `json:"notes,omitempty"`
}

type Software struct {
	OS           string `json:"os,omitempty"`
	OTASupported *bool  `json:"ota_supported,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

type Links struct {
	ManufacturerURL string `json:"manufacturer_url,omitempty"`
	PressKitURL     string `json:"press_kit_url,omitempty"`
	SpecSheetURL    string `json:"spec_sheet_url,omitempty"`
	ConfiguratorURL string `json:"configurator_url,omitempty"`
}

type Images struct {
	ExteriorURL          string `json:"exterior_url,omitempty"`
	InteriorURL          string `json:"interior_url,omitempty"`
	ChargingCurvePlotURL string `json:"charging_curve_plot_url,omitempty"`
}

type Source struct {
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Publisher  string    `json:"publisher,omitempty"`
	URL        string    `json:"url"`
	AccessedAt time.Time `json:"accessed_at"`
	License    string    `json:"license,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}

type Metadata struct {
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	Contributors  []string   `json:"contributors,omitempty"`
	DataQuality   string     `json:"data_quality,omitempty"`
	InternalNotes string     `json:"internal_notes,omitempty"`
}

type Stats struct {
	TotalVehicles int            `json:"total_vehicles"`
	MakeCount     int            `json:"make_count"`
	YearRange     [2]int         `json:"year_range"`
	ByType        map[string]int `json:"by_type"`
	ByMake        map[string]int `json:"by_make"`
}
