package device_test

import (
	"testing"
	"time"

	"medops/internal/device"
)

func TestDeviceRegistrationAndActivation(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-001", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	if dev.Status != device.StatusRegistered {
		t.Fatalf("expected registered, got %s", dev.Status)
	}
	if err := dev.Activate(at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if dev.Status != device.StatusInUse {
		t.Fatalf("expected in-use, got %s", dev.Status)
	}
}

func TestFirmwareUpgradeHistory(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	firmware := device.NewFirmware("v1.0.0", at)
	if err := firmware.Upgrade("v1.1.0", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if firmware.Current() != "v1.1.0" {
		t.Fatalf("current version mismatch: %s", firmware.Current())
	}
	if got := firmware.VersionAt(at); got != "v1.0.0" {
		t.Fatalf("version at creation mismatch: %s", got)
	}
	if got := firmware.VersionAt(at.Add(2 * time.Hour)); got != "v1.1.0" {
		t.Fatalf("version after upgrade mismatch: %s", got)
	}
}

func TestLockoutBlocksActivation(t *testing.T) {
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	dev, err := device.NewDevice("SN-002", "M-1", "监护仪", "H1", "W1", device.ClassificationGeneral, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Activate(at); err != nil {
		t.Fatal(err)
	}
	if err := dev.BeginMaintenance("scheduled service", at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := dev.Activate(at.Add(2 * time.Minute)); err == nil {
		t.Fatal("activation with active lockout accepted")
	}
	dev.Lockout.Clear()
	if err := dev.Activate(at.Add(3 * time.Minute)); err != nil {
		t.Fatalf("activation after lockout clear failed: %v", err)
	}
	if dev.Status != device.StatusInUse {
		t.Fatalf("expected in-use after activation, got %s", dev.Status)
	}
}
