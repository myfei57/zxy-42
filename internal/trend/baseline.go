package trend

import (
	"errors"
	"time"

	"medops/internal/device"
)

var ErrNoTrendData = errors.New("trend: no trend data for device")

// Sample is one measured value in the performance history.
type Sample struct {
	At        time.Time `json:"at"`
	Value     float64   `json:"value"`
	Parameter string    `json:"parameter"`
}

// Baseline is the persisted performance history of one device.
type Baseline struct {
	DeviceID       string                `json:"device_id"`
	Classification device.Classification `json:"classification"`
	CreatedAt      time.Time             `json:"created_at"`
	Samples        []Sample              `json:"samples"`
}

func (s *Service) baselinePath(deviceID string) string {
	return "trend/" + deviceID + ".json"
}

// EnsureBaseline loads the device baseline or creates one from the device's
// current classification.
func (s *Service) EnsureBaseline(dev *device.Device, at time.Time) (*Baseline, error) {
	var baseline Baseline
	if err := s.fs.ReadJSON(s.baselinePath(dev.ID), &baseline); err == nil {
		return &baseline, nil
	}
	baseline = Baseline{
		DeviceID:       dev.ID,
		Classification: dev.CurrentClassification(),
		CreatedAt:      at,
	}
	s.ranges[dev.ID] = dev.CurrentClassification()
	if err := s.fs.WriteJSON(s.baselinePath(dev.ID), baseline); err != nil {
		return nil, err
	}
	return &baseline, nil
}

// Append records one measured value in the device history.
func (s *Service) Append(dev *device.Device, value float64, parameter string, at time.Time) error {
	baseline, err := s.EnsureBaseline(dev, at)
	if err != nil {
		return err
	}
	baseline.Samples = append(baseline.Samples, Sample{At: at, Value: value, Parameter: parameter})
	baseline.Samples = TrimHistory(baseline.Samples)
	return s.fs.WriteJSON(s.baselinePath(dev.ID), baseline)
}
