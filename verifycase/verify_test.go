package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/plan"
	"medops/internal/store"
)

func TestMoLockoutClearedOnMaintenanceComplete(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := plan.NewService(fs)
	dev, err := device.NewDevice("SN-LCK-1", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Activate(at); err != nil {
		t.Fatal(err)
	}
	p, err := service.Create(dev, device.RiskMedium, "scheduled service", at.Add(time.Hour), at)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Begin(p, dev, p.Reason, at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := service.Complete(p, dev, at.Add(2*time.Hour)); err != nil {
		t.Fatalf("maintenance completion failed while lockout set: %v", err)
	}
	if dev.Lockout.IsActive() {
		t.Fatal("lockout still active after maintenance completed")
	}
	if dev.Status != device.StatusInUse {
		t.Fatalf("device did not return to in-use: %s", dev.Status)
	}
}
