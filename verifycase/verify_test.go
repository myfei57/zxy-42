package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/plan"
)

func TestMoPlanSortsByRisk(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	items := []plan.Item{
		{PlanID: "p-low", DeviceID: "d1", Risk: device.RiskLow, Due: at.Add(time.Hour)},
		{PlanID: "p-high", DeviceID: "d2", Risk: device.RiskHigh, Due: at.Add(3 * time.Hour)},
	}
	sorted := plan.SortItems(items)
	if sorted[0].PlanID != "p-high" {
		t.Fatalf("high-risk device pushed behind low-risk one: %s first", sorted[0].PlanID)
	}
}
