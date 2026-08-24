package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
)

func TestMoQcMethodFollowsFirmware(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-MTH-1", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	methods := qc.NewMethodService([]qc.Method{
		{FirmwareVersion: "v1.0.0", Name: "legacy-spo2", Parameters: []string{"spo2"}},
		{FirmwareVersion: "v1.1.0", Name: "continuous-spo2", Parameters: []string{"spo2", "pr"}},
	})
	plan, err := qc.NewPlan(dev, 7, "v1.0.0", dev.Firmware.Current(), "LOT-A", at)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Firmware.Upgrade("v1.1.0", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	method, err := methods.Resolve(plan, dev)
	if err != nil {
		t.Fatal(err)
	}
	if method.Name != "continuous-spo2" {
		t.Fatalf("QC method bound to stale firmware: %s", method.Name)
	}
}
