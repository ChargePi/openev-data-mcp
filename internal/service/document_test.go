package service

import (
	"strings"
	"testing"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
)

func ptr(f float64) *float64 { return &f }

func boolPtr(b bool) *bool { return &b }

func fullVehicle() vehicle.Vehicle {
	return vehicle.Vehicle{
		UniqueCode:  "audi:a6_e_tron:2024:a6_e_tron",
		Make:        vehicle.NamedEntity{Slug: "audi", Name: "Audi"},
		Model:       vehicle.NamedEntity{Slug: "a6_e_tron", Name: "A6 e-tron"},
		Year:        2024,
		Trim:        vehicle.NamedEntity{Slug: "base", Name: "Base"},
		Variant:     &vehicle.Variant{Slug: "sportback", Name: "Sportback"},
		VehicleType: "passenger_car",
		Powertrain:  vehicle.Powertrain{Drivetrain: "rwd", SystemPowerKW: ptr(350)},
		Battery: vehicle.Battery{
			PackCapacityKWhGross: ptr(100),
			PackCapacityKWhNet:   ptr(94.9),
			Chemistry:            "li-ion",
			ThermalManagement:    "liquid",
			HeatPump:             boolPtr(true),
		},
		Charging: vehicle.Charging{
			AC: &vehicle.ACCharging{MaxPowerKW: 11},
			DC: &vehicle.DCCharging{MaxPowerKW: 270},
		},
		ChargePorts: []vehicle.ChargePort{
			{Kind: "combo", Connector: "ccs2", Location: &vehicle.ChargePortLocation{Side: "right", Position: "rear"}},
			{Kind: "ac_only", Connector: "type2"},
		},
		Range: vehicle.Range{Rated: []vehicle.RangeRating{
			{Cycle: "wltp", RangeKm: 700, Notes: "secret-range-note"},
			{Cycle: "epa", RangeKm: 610.5},
		}},
	}
}

func TestBuildDocument_AllFields(t *testing.T) {
	doc, err := buildDocument(fullVehicle())
	if err != nil {
		t.Fatalf("buildDocument: %v", err)
	}

	want := `Make: Audi
Model: A6 e-tron
Year: 2024
Trim: Base
Variant: Sportback
Vehicle type: passenger_car
Drivetrain: rwd
Power: 350 kW
Battery gross capacity: 100 kWh
Battery net capacity: 94.9 kWh
Battery chemistry: li-ion
Battery thermal management: liquid
Heat pump: yes
AC charging power: 11 kW
DC charging power: 270 kW
Range: wltp 700 km, epa 610.5 km
Charging ports: combo ccs2, ac_only type2`
	if strings.TrimSpace(doc) != want {
		t.Fatalf("unexpected document:\n%s\nwant:\n%s", doc, want)
	}
}

func TestBuildDocument_ExcludesNotesAndPortLocation(t *testing.T) {
	doc, err := buildDocument(fullVehicle())
	if err != nil {
		t.Fatalf("buildDocument: %v", err)
	}
	for _, s := range []string{"secret-range-note", "right", "rear"} {
		if strings.Contains(doc, s) {
			t.Errorf("document should not contain %q:\n%s", s, doc)
		}
	}
}

func TestBuildDocument_OmitsMissingFields(t *testing.T) {
	v := vehicle.Vehicle{
		UniqueCode:  "x:y:2020:z",
		Make:        vehicle.NamedEntity{Name: "Make"},
		Model:       vehicle.NamedEntity{Name: "Model"},
		Year:        2020,
		Trim:        vehicle.NamedEntity{Name: "Trim"},
		VehicleType: "passenger_car",
		Battery:     vehicle.Battery{HeatPump: boolPtr(false)},
	}
	doc, err := buildDocument(v)
	if err != nil {
		t.Fatalf("buildDocument: %v", err)
	}

	want := `Make: Make
Model: Model
Year: 2020
Trim: Trim
Vehicle type: passenger_car
Heat pump: no`
	if strings.TrimSpace(doc) != want {
		t.Fatalf("unexpected document:\n%s\nwant:\n%s", doc, want)
	}
}
