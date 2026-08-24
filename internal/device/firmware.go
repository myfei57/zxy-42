package device

import (
	"errors"
	"time"
)

// FirmwareVersion is one applied firmware release.
type FirmwareVersion struct {
	Version   string    `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}

// Firmware tracks the installed version and the full upgrade history.
type Firmware struct {
	CurrentVersion string            `json:"current_version"`
	History        []FirmwareVersion `json:"history"`
}

func NewFirmware(version string, at time.Time) *Firmware {
	return &Firmware{
		CurrentVersion: version,
		History:        []FirmwareVersion{{Version: version, AppliedAt: at}},
	}
}

// Current returns the firmware version currently installed on the device.
func (f *Firmware) Current() string { return f.CurrentVersion }

// Upgrade applies a new firmware version and records the change.
func (f *Firmware) Upgrade(version string, at time.Time) error {
	if version == "" {
		return errors.New("device: firmware version required")
	}
	if version == f.CurrentVersion {
		return nil
	}
	f.CurrentVersion = version
	f.History = append(f.History, FirmwareVersion{Version: version, AppliedAt: at})
	return nil
}

// VersionAt returns the firmware version in effect at the given moment.
func (f *Firmware) VersionAt(at time.Time) string {
	version := f.History[0].Version
	for _, entry := range f.History {
		if !entry.AppliedAt.After(at) {
			version = entry.Version
		}
	}
	return version
}

// Upgrades returns the recorded firmware releases, oldest first.
func (f *Firmware) Upgrades() []FirmwareVersion {
	return append([]FirmwareVersion(nil), f.History...)
}
