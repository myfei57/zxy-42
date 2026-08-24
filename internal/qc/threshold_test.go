package qc_test

import (
	"testing"
	"time"

	medops_device "medops/internal/device"
	"medops/internal/qc"
)

// newDeviceFixture builds a device whose QC plan was created under the old
// firmware and then upgraded, mirroring the batch that mass-upgraded firmware
// overnight and ran QC the next morning.
func newDeviceFixture(t *testing.T, planAt, upgradeAt time.Time) (*medops_device.Device, *qc.Plan) {
	t.Helper()
	dev, err := medops_device.NewDevice("SN-100", "M-9", "监护仪", "H1", "W1", medops_device.ClassificationGeneral, planAt)
	if err != nil {
		t.Fatal(err)
	}
	// Plan bound while the device still ran the pre-upgrade firmware.
	plan, err := qc.NewPlan(dev, 7, dev.Firmware.VersionAt(planAt), dev.Firmware.Current(), "LOT-A", planAt)
	if err != nil {
		t.Fatal(err)
	}
	// Thursday-night firmware upgrade: the device moves to v1.1.0 after the
	// plan was created. The plan snapshot must stay on v1.0.0 for the audit
	// trail, but live evaluation must follow the upgrade.
	if err := dev.Firmware.Upgrade("v1.1.0", upgradeAt); err != nil {
		t.Fatal(err)
	}
	return dev, plan
}

func TestThresholdEvaluateFollowsCurrentFirmware(t *testing.T) {
	// v1.0.0: spo2 90..100 ; v1.1.0: spo2 92..100 (tighter lower bound).
	thresholds := qc.NewThresholdService([]qc.ThresholdTable{
		{FirmwareVersion: "v1.0.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 90, Max: 100}}},
		{FirmwareVersion: "v1.1.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 92, Max: 100}}},
	})
	planAt := time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	upgradeAt := time.Date(2026, 8, 20, 22, 0, 0, 0, time.UTC) // Thursday night
	dev, plan := newDeviceFixture(t, planAt, upgradeAt)

	if plan.ThresholdVersion != "v1.0.0" {
		t.Fatalf("audit snapshot should stay on pre-upgrade firmware: got %s", plan.ThresholdVersion)
	}
	if dev.Firmware.Current() != "v1.1.0" {
		t.Fatalf("device should report upgraded firmware: got %s", dev.Firmware.Current())
	}

	// spo2 = 91: passes v1.0.0 (>=90) but fails v1.1.0 (<92). The Friday
	// morning QC run must use the upgraded firmware's band, so 91 must be a
	// fail, not the false "pass" the pre-upgrade snapshot would produce.
	verdict, err := thresholds.Evaluate(plan, dev, "spo2", 91)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Min != 92 || verdict.Max != 100 {
		t.Fatalf("evaluated against wrong threshold band: min=%v max=%v (expected v1.1.0 92..100)", verdict.Min, verdict.Max)
	}
	if verdict.Passed {
		t.Fatalf("value 91 must fail under upgraded firmware v1.1.0 (92..100); got passed=%v", verdict.Passed)
	}

	// A value in-spec for the upgraded firmware must pass, confirming we did
	// not simply invert the result.
	ok, err := thresholds.Evaluate(plan, dev, "spo2", 96)
	if err != nil {
		t.Fatal(err)
	}
	if !ok.Passed {
		t.Fatalf("value 96 must pass under upgraded firmware v1.1.0 (92..100); got passed=%v", ok.Passed)
	}
}

func TestThresholdEvaluatePreUpgradeUsesSnapshot(t *testing.T) {
	// When no upgrade has happened, plan snapshot and current firmware agree,
	// so evaluation must still resolve the device's current firmware and the
	// looser v1.0.0 band applies.
	thresholds := qc.NewThresholdService([]qc.ThresholdTable{
		{FirmwareVersion: "v1.0.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 90, Max: 100}}},
		{FirmwareVersion: "v1.1.0", Thresholds: []qc.Threshold{{Parameter: "spo2", Min: 92, Max: 100}}},
	})
	planAt := time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)
	dev, err := medops_device.NewDevice("SN-101", "M-9", "监护仪", "H1", "W1", medops_device.ClassificationGeneral, planAt)
	if err != nil {
		t.Fatal(err)
	}
	// No firmware upgrade: plan snapshot and current firmware are both v1.0.0.
	plan, err := qc.NewPlan(dev, 7, dev.Firmware.VersionAt(planAt), dev.Firmware.Current(), "LOT-A", planAt)
	if err != nil {
		t.Fatal(err)
	}

	verdict, err := thresholds.Evaluate(plan, dev, "spo2", 91)
	if err != nil {
		t.Fatal(err)
	}
	if !verdict.Passed {
		t.Fatalf("value 91 must pass under v1.0.0 (90..100); got passed=%v reason=%s", verdict.Passed, verdict.Reason)
	}
}
