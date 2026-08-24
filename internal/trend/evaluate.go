package trend

import "medops/internal/device"

// Evaluation is the outcome of comparing the latest sample with the range of
// the device's current classification.
type Evaluation struct {
	DeviceID    string  `json:"device_id"`
	Range       Range   `json:"range"`
	Value       float64 `json:"value"`
	Parameter   string  `json:"parameter"`
	WithinRange bool    `json:"within_range"`
	AlertRaised bool    `json:"alert_raised"`
}

func (s *Service) Evaluate(dev *device.Device) (Evaluation, error) {
	sample, ok := s.Latest(dev)
	if !ok {
		return Evaluation{}, ErrNoTrendData
	}
	rng := s.RangeFor(dev)
	within := sample.Value >= rng.Min && sample.Value <= rng.Max
	evaluation := Evaluation{
		DeviceID:    dev.ID,
		Range:       rng,
		Value:       sample.Value,
		Parameter:   sample.Parameter,
		WithinRange: within,
	}
	if !within && s.alerts != nil {
		severity := "medium"
		if dev.CurrentClassification() == device.ClassificationICU {
			severity = "high"
		}
		if _, err := s.alerts.Raise(dev, "trend-out-of-range", severity, "trend value outside classification range", sample.At); err == nil {
			evaluation.AlertRaised = true
		}
	}
	return evaluation, nil
}
