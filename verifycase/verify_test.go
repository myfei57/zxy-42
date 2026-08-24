package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
)

func TestMoQcDispatchFollowsWardMove(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-WRD-1", "M-1", "监护仪", "H1", "W-A", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := qc.NewPlan(dev, 7, "v1.0.0", "v1.0.0", "LOT-A", at)
	if err != nil {
		t.Fatal(err)
	}
	plan.Status = qc.PlanStatusActive
	if err := dev.MoveWard("W-B", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	schedule := &qc.ScheduleService{}
	dispatch, err := schedule.Dispatch(plan, dev, at.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if dispatch.TargetWard != "W-B" {
		t.Fatalf("QC dispatch still targets the old ward: %s", dispatch.TargetWard)
	}
}
