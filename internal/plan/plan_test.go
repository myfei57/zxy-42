package plan_test

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/plan"
	"medops/internal/store"
)

func TestPlanCreateAndDue(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := plan.NewService(fs)
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-004", "M-3", "输液泵", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	p, err := service.Create(dev, device.RiskHigh, "trend out of range", at.Add(time.Hour), at)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != plan.StatusScheduled {
		t.Fatalf("expected scheduled, got %s", p.Status)
	}
	if len(service.Due(at.Add(2*time.Hour))) != 1 {
		t.Fatal("due plan not returned")
	}
	if len(service.Due(at)) != 0 {
		t.Fatal("plan returned before due")
	}
}
