package device

import (
	"time"

	"medops/internal/ns"
)

// RegisterParams carries the validated input of a device registration.
type RegisterParams struct {
	Serial         string
	Model          string
	Name           string
	HospitalID     string
	WardID         string
	Classification Classification
}

// Register validates the ward placement, checks serial uniqueness and stores
// the new device in the registry.
func Register(reg *Registry, wards *ns.WardRegistry, params RegisterParams, at time.Time) (*Device, error) {
	if _, ok := wards.Ward(params.WardID); !ok {
		return nil, ns.ErrWardNotFound
	}
	for _, existing := range reg.All() {
		if existing.Serial == params.Serial {
			return nil, ErrSerialExists
		}
	}
	dev, err := NewDevice(params.Serial, params.Model, params.Name, params.HospitalID, params.WardID, params.Classification, at)
	if err != nil {
		return nil, err
	}
	if err := reg.Register(dev); err != nil {
		return nil, err
	}
	return dev, nil
}
