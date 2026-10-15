package database

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// loadRecordSchema compiles the upstream OpenEV Data record schema.
func loadRecordSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	schema, err := jsonschema.NewCompiler().Compile("testdata/schema.json")
	if err != nil {
		t.Fatalf("compiling record schema: %v", err)
	}
	return schema
}

// decodeJSON decodes JSON into generic values so documents can be compared and validated.
func decodeJSON(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decoding JSON: %v", err)
	}
	return out
}

// TestVehicleModelToDomain_MatchesSchema checks that a record using every schema field
// comes back unchanged (minus the "$schema" path) and validates against the schema.
func TestVehicleModelToDomain_MatchesSchema(t *testing.T) {
	t.Parallel()
	record, err := os.ReadFile("testdata/full_record.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	m := VehicleModel{ID: 7, UniqueCode: "acme:volt_one:2025:long_range", JsonData: record}
	got, err := m.ToDomain()
	if err != nil {
		t.Fatalf("ToDomain: %v", err)
	}
	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}

	out, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshaling vehicle: %v", err)
	}
	output := decodeJSON(t, out)

	if err := loadRecordSchema(t).Validate(output); err != nil {
		t.Errorf("output does not match the record schema: %v", err)
	}

	want := decodeJSON(t, record)
	delete(want, "$schema")
	if !reflect.DeepEqual(output, want) {
		t.Errorf("output differs from the record:\n got: %s\nwant: %s", out, record)
	}
}

func TestVehicleModelToDomain_InvalidData(t *testing.T) {
	t.Parallel()
	m := VehicleModel{UniqueCode: "x:y:2020:z", JsonData: json.RawMessage(`{"body": "not an object"}`)}
	if _, err := m.ToDomain(); err == nil {
		t.Fatal("expected an error for malformed data")
	}
}
