package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
)

func TestMoThresholdFollowsFirmware(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-THR-1", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	thresholds := qc.NewThresholdService([]qc.ThresholdTable{
		{FirmwareVersion: "v1.0.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 90, Max: 100}}},
		{FirmwareVersion: "v1.1.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 92, Max: 100}}},
	})
	plan, err := qc.NewPlan(dev, 7, dev.Firmware.Current(), dev.Firmware.Current(), "LOT-A", at)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Firmware.Upgrade("v1.1.0", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	verdict, err := thresholds.Evaluate(plan, dev, "spo2", 91)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Passed {
		t.Fatalf("value 91 must fail the v1.1.0 threshold (min 92): %+v", verdict)
	}
}
