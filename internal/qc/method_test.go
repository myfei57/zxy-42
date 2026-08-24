package qc_test

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
)

// methods mirror the firmware-bound method table wired in production: the
// legacy firmware measures spo2 only, the upgraded firmware also requires pr.
func newTestMethodService() *qc.MethodService {
	return qc.NewMethodService([]qc.Method{
		{FirmwareVersion: "v1.0.0", Name: "legacy-spo2", Parameters: []string{"spo2"}},
		{FirmwareVersion: "v1.1.0", Name: "continuous-spo2", Parameters: []string{"spo2", "pr"}},
	})
}

// TestResolveFollowsCurrentFirmware guards the firmware-upgrade regression:
// a plan created under v1.0.0 must resolve to the v1.1.0 method once the
// device has been upgraded, otherwise QC runs the old procedure and the
// report omits the new parameter required by the current firmware.
func TestResolveFollowsCurrentFirmware(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-001", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}

	// Plan is drafted while the device still runs v1.0.0 firmware; the
	// method_version snapshot is therefore the legacy one.
	plan, err := qc.NewPlan(dev, 7, "v1.0.0", "v1.0.0", "LOT-A", at)
	if err != nil {
		t.Fatal(err)
	}

	// Firmware is upgraded in the field; the live method must follow it.
	if err := dev.Firmware.Upgrade("v1.1.0", at.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	methods := newTestMethodService()
	method, err := methods.Resolve(plan, dev)
	if err != nil {
		t.Fatalf("resolve after upgrade: %v", err)
	}
	if method.Name != "continuous-spo2" {
		t.Fatalf("expected upgraded method continuous-spo2, got %s (still bound to plan snapshot %s)",
			method.Name, plan.MethodVersion)
	}
	if len(method.Parameters) != 2 {
		t.Fatalf("expected upgraded method to require spo2 and pr, got %v", method.Parameters)
	}
}

// TestResolveRejectsUnknownFirmware ensures a method missing for the current
// firmware surfaces a clear error rather than silently falling back to the
// plan snapshot.
func TestResolveRejectsUnknownFirmware(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-002", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Firmware.Upgrade("v9.9.9", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	plan, err := qc.NewPlan(dev, 7, "v1.0.0", "v1.0.0", "LOT-A", at)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newTestMethodService().Resolve(plan, dev); err == nil {
		t.Fatal("expected error when no method exists for current firmware")
	}
}
