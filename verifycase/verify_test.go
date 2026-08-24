package verifycase

import (
	"testing"
	"time"

	"medops/internal/device"
	"medops/internal/heart"
)

func TestMoHeartbeatWindowCurrentClock(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-WIN-1", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	service := heart.NewService(heart.NewWindow(30*time.Minute, 2*time.Minute), nil)
	first := heart.NewBeat(dev.DeviceID(), 1, at.Add(time.Minute), at.Add(time.Minute))
	if _, err := service.Record(dev, first); err != nil {
		t.Fatalf("first beat rejected: %v", err)
	}
	dev.Clock.Advance(2 * time.Hour)
	now := dev.Clock.Current()
	beat := heart.NewBeat(dev.DeviceID(), 2, now.Add(-time.Minute), now)
	if _, err := service.Record(dev, beat); err != nil {
		t.Fatalf("active device marked offline after clock shift: %v", err)
	}
}
