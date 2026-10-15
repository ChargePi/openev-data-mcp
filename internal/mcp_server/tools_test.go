package mcp_server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
	"github.com/google/jsonschema-go/jsonschema"
)

func ptr(f float64) *float64 { return &f }

func boolPtr(b bool) *bool { return &b }

func intPtr(i int) *int { return &i }

// TestSearchVehiclesOutputSchema checks that real tool output validates against
// the advertised output schema, so strict MCP clients accept it.
func TestSearchVehiclesOutputSchema(t *testing.T) {
	raw, err := searchVehiclesOutputSchema()
	if err != nil {
		t.Fatalf("searchVehiclesOutputSchema: %v", err)
	}

	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshaling schema: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("resolving schema: %v", err)
	}

	outputs := map[string]searchVehiclesOutput{
		"empty": {Vehicles: []vehicle.Vehicle{}},
		"full": {Vehicles: []vehicle.Vehicle{{
			ID:            1,
			SchemaVersion: "1.0.0",
			UniqueCode:    "audi:a6_e_tron:2024:a6_e_tron",
			Make:          vehicle.NamedEntity{Slug: "audi", Name: "Audi"},
			Model:         vehicle.NamedEntity{Slug: "a6_e_tron", Name: "A6 e-tron"},
			Year:          2024,
			Trim:          vehicle.NamedEntity{Slug: "base", Name: "Base"},
			Variant:       &vehicle.Variant{Slug: "sportback", Name: "Sportback", Kind: "body_style"},
			Markets:       []string{"DE"},
			VehicleType:   "passenger_car",
			Body:          &vehicle.Body{Style: "sedan", Doors: intPtr(4), Seats: intPtr(5)},
			Powertrain: vehicle.Powertrain{
				Drivetrain:    "rwd",
				SystemPowerKW: ptr(350),
				Motors:        []vehicle.Motor{{Position: "rear", Type: "pmsm", PowerKW: ptr(350)}},
			},
			Battery: vehicle.Battery{PackCapacityKWhNet: ptr(94.9), HeatPump: boolPtr(true)},
			ChargePorts: []vehicle.ChargePort{{
				Kind:      "combo",
				Connector: "ccs2",
				Location:  &vehicle.ChargePortLocation{Side: "right", Position: "rear"},
				Motorized: boolPtr(false),
			}},
			Charging: vehicle.Charging{DC: &vehicle.DCCharging{MaxPowerKW: 270, ArchitectureVoltageClass: "800v"}},
			V2X: &vehicle.V2X{V2L: &vehicle.V2L{
				Supported: true,
				Outlets:   []vehicle.V2LOutlet{{Kind: "socket_type_f", Location: "cabin", Count: intPtr(1)}},
			}},
			Range: vehicle.Range{Rated: []vehicle.RangeRating{{Cycle: "wltp", RangeKm: 700}}},
			Pricing: &vehicle.Pricing{MSRP: []vehicle.Price{{
				Currency: "EUR", Amount: 75000, Country: "DE", Year: intPtr(2024),
			}}},
			Sources: []vehicle.Source{{
				Type:       "oem",
				Title:      "Spec sheet",
				URL:        "https://www.audi.com",
				AccessedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			}},
		}}},
		"minimal": {Vehicles: []vehicle.Vehicle{{
			ID:          2,
			Make:        vehicle.NamedEntity{Slug: "x", Name: "X"},
			Model:       vehicle.NamedEntity{Slug: "y", Name: "Y"},
			Year:        2020,
			Trim:        vehicle.NamedEntity{Slug: "z", Name: "Z"},
			VehicleType: "passenger_car",
		}}},
	}

	for name, output := range outputs {
		t.Run(name, func(t *testing.T) {
			b, err := json.Marshal(output)
			if err != nil {
				t.Fatalf("marshaling output: %v", err)
			}
			var instance map[string]any
			if err := json.Unmarshal(b, &instance); err != nil {
				t.Fatalf("unmarshaling output: %v", err)
			}
			if err := resolved.Validate(instance); err != nil {
				t.Fatalf("output does not match schema: %v", err)
			}
		})
	}
}
