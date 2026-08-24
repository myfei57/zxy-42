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

// TestSortItemsByRiskThenDue mirrors the reported ventilator scheduling
// issue: a high-risk plan must never queue behind a low-risk one, even when
// the low-risk plan is due earlier. Risk is the primary key; due time only
// breaks ties within the same risk level.
func TestSortItemsByRiskThenDue(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	items := []plan.Item{
		{PlanID: "low-early", DeviceID: "infusion", Risk: device.RiskLow, Due: at.Add(1 * time.Hour)},
		{PlanID: "high-late", DeviceID: "ventilator", Risk: device.RiskHigh, Due: at.Add(3 * time.Hour)},
		{PlanID: "high-early", DeviceID: "monitor", Risk: device.RiskHigh, Due: at.Add(2 * time.Hour)},
	}

	got := plan.SortItems(items)

	want := []string{"high-early", "high-late", "low-early"}
	for i, w := range want {
		if got[i].PlanID != w {
			t.Fatalf("position %d: want %s, got %s (full order: %v)",
				i, w, got[i].PlanID, planIDs(got))
		}
	}
}

func planIDs(items []plan.Item) []string {
	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.PlanID)
	}
	return ids
}
