package verifycase

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
	"medops/internal/store"
)

func TestMoPartialCommitKeepsRemaining(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dev, err := device.NewDevice("SN-COM-1", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	schedule := qc.NewScheduleService(fs, nil, nil, nil, store.NewBatchStore(fs))
	plan, err := qc.NewPlan(dev, 7, "v1.0.0", "v1.0.0", "LOT-A", at)
	if err != nil {
		t.Fatal(err)
	}
	plan.Status = qc.PlanStatusActive
	blocker := filepath.Join(fs.Root, "results", "batches", plan.ID, "dev-2.json")
	if err := os.MkdirAll(blocker, 0o755); err != nil {
		t.Fatal(err)
	}
	batch := qc.NewBatch([]qc.BatchItem{
		{DeviceID: "dev-1", Result: &qc.Result{DeviceID: "dev-1"}},
		{DeviceID: "dev-2", Result: &qc.Result{DeviceID: "dev-2"}},
		{DeviceID: "dev-3", Result: &qc.Result{DeviceID: "dev-3"}},
	})
	if _, err := schedule.CommitBatch(plan, batch); err == nil {
		t.Fatal("commit succeeded despite blocked result path")
	}
	if len(plan.PendingRetry) == 0 {
		t.Fatal("partial commit left the remaining devices untracked")
	}
	want := map[string]bool{"dev-2": true, "dev-3": true}
	for _, id := range plan.PendingRetry {
		if !want[id] {
			t.Fatalf("unexpected pending device: %s", id)
		}
	}
}
