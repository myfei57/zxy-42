package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
	"medops/internal/store"
	"medops/internal/test"
)

func TestMoReagentExpiryAtExecutionTime(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dev, err := device.NewDevice("SN-RGT-1", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	thresholds := qc.NewThresholdService([]qc.ThresholdTable{
		{FirmwareVersion: "v1.0.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 90, Max: 100}}},
	})
	reagents := test.NewReagentRegistry()
	if err := reagents.Add(test.NewReagent("质控液", "LOT-E", at.Add(time.Hour), at)); err != nil {
		t.Fatal(err)
	}
	cals := qc.NewCalibrationRegistry(fs)
	reports := store.NewReportStore(fs)
	service := test.NewService(reagents, cals, thresholds, reports, nil)
	plan, err := qc.NewPlan(dev, 7, "v1.0.0", "v1.0.0", "LOT-E", at)
	if err != nil {
		t.Fatal(err)
	}
	cal := qc.NewCalibration(plan.ID, []string{"zero", "span", "linearity"})
	if err := cal.RunSequence(cal.Steps, at); err != nil {
		t.Fatal(err)
	}
	if err := cals.Put(cal); err != nil {
		t.Fatal(err)
	}
	method := &qc.Method{FirmwareVersion: "v1.0.0", Name: "legacy-spo2", Parameters: []string{"spo2"}}
	if _, err := service.Run(plan, dev, method, "spo2", 96, at.Add(2*time.Hour)); err == nil {
		t.Fatal("expired reagent was still used at test execution time")
	}
}
