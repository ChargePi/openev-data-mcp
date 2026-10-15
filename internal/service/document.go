package service

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"strconv"
	"text/template"

	"github.com/ChargePi/openev-data-mcp/pkg/vehicle"
)

//go:embed templates/vehicle_document.tmpl
var templatesFS embed.FS

var vehicleDocumentTmpl = template.Must(template.ParseFS(templatesFS, "templates/vehicle_document.tmpl"))

type vehicleDocument struct {
	Make        string
	Model       string
	Year        int
	Trim        string
	Variant     string
	VehicleType string
	Drivetrain  string
	PowerKW     string
	Battery     batteryDocument
	ACPowerKW   string
	DCPowerKW   string
	Ranges      []rangeDocument
	ChargePorts []chargePortDocument
}

type batteryDocument struct {
	GrossKWh          string
	NetKWh            string
	Chemistry         string
	ThermalManagement string
	HeatPump          string
}

type rangeDocument struct {
	Cycle string
	Km    string
}

type chargePortDocument struct {
	Kind      string
	Connector string
}

// buildDocument renders the text that gets embedded for a vehicle.
func buildDocument(v vehicle.Vehicle) (string, error) {
	doc := vehicleDocument{
		Make:        v.Make.Name,
		Model:       v.Model.Name,
		Year:        v.Year,
		Trim:        v.Trim.Name,
		VehicleType: v.VehicleType,
		Drivetrain:  v.Powertrain.Drivetrain,
		PowerKW:     formatFloat(v.Powertrain.SystemPowerKW),
		Battery: batteryDocument{
			GrossKWh:          formatFloat(v.Battery.PackCapacityKWhGross),
			NetKWh:            formatFloat(v.Battery.PackCapacityKWhNet),
			Chemistry:         v.Battery.Chemistry,
			ThermalManagement: v.Battery.ThermalManagement,
		},
	}

	if v.Variant != nil {
		doc.Variant = v.Variant.Name
	}
	if v.Battery.HeatPump != nil {
		doc.Battery.HeatPump = "no"
		if *v.Battery.HeatPump {
			doc.Battery.HeatPump = "yes"
		}
	}
	if v.Charging.AC != nil {
		doc.ACPowerKW = formatFloat(&v.Charging.AC.MaxPowerKW)
	}
	if v.Charging.DC != nil {
		doc.DCPowerKW = formatFloat(&v.Charging.DC.MaxPowerKW)
	}

	for _, r := range v.Range.Rated {
		doc.Ranges = append(doc.Ranges, rangeDocument{Cycle: r.Cycle, Km: formatFloat(&r.RangeKm)})
	}
	for _, p := range v.ChargePorts {
		doc.ChargePorts = append(doc.ChargePorts, chargePortDocument{Kind: p.Kind, Connector: p.Connector})
	}

	var buf bytes.Buffer
	if err := vehicleDocumentTmpl.Execute(&buf, doc); err != nil {
		return "", fmt.Errorf("rendering vehicle %q document: %w", v.UniqueCode, err)
	}
	return buf.String(), nil
}

func documentHash(doc string) string {
	sum := sha256.Sum256([]byte(doc))
	return hex.EncodeToString(sum[:])
}

func formatFloat(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}
