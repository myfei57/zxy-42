package trend_test

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/store"
	"medops/internal/trend"
)

func TestBaselineAppendAndHistoryCap(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := trend.NewService(
		fs,
		trend.Range{Classification: device.ClassificationICU, Min: 94, Max: 100},
		trend.Range{Classification: device.ClassificationGeneral, Min: 90, Max: 100},
		nil,
	)
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-003", "M-2", "呼吸机", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 30; i++ {
		if err := service.Append(dev, 95+float64(i%5), "spo2", at.Add(time.Duration(i)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	history := service.History(dev)
	if len(history) != trend.MaxHistory {
		t.Fatalf("history cap mismatch: got %d want %d", len(history), trend.MaxHistory)
	}
	latest, ok := service.Latest(dev)
	if !ok || latest.Value != 95+float64(29%5) {
		t.Fatalf("latest sample mismatch: %+v ok=%v", latest, ok)
	}
}
