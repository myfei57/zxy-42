package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/store"
	"medops/internal/trend"
)

func TestMoTrendUsesReclassifiedRange(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
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
	dev, err := device.NewDevice("SN-RCL-1", "M-1", "监护仪", "H1", "W1", device.ClassificationICU, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Append(dev, 91, "spo2", at); err != nil {
		t.Fatal(err)
	}
	if err := dev.Reclassify(device.ClassificationGeneral, at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	evaluation, err := service.Evaluate(dev)
	if err != nil {
		t.Fatal(err)
	}
	if !evaluation.WithinRange {
		t.Fatalf("reclassified device judged against a stale range: %+v", evaluation.Range)
	}
}
