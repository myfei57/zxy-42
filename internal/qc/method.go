package qc

import (
	"fmt"

	"medops/internal/device"
)

// Method is the QC measurement procedure required by one firmware version.
type Method struct {
	FirmwareVersion string   `json:"firmware_version"`
	Name            string   `json:"name"`
	Parameters      []string `json:"parameters"`
}

type MethodService struct {
	methods []Method
}

func NewMethodService(methods []Method) *MethodService {
	return &MethodService{methods: methods}
}

func (s *MethodService) ForFirmware(version string) (*Method, error) {
	for i := range s.methods {
		if s.methods[i].FirmwareVersion == version {
			return &s.methods[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrMethodNotFound, version)
}

// Resolve binds the QC method to the device's current firmware version.
func (s *MethodService) Resolve(plan *Plan, dev *device.Device) (*Method, error) {
	return s.ForFirmware(dev.Firmware.Current())
}
