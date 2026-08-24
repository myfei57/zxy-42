package trend

import "medops/internal/device"

const MaxHistory = 24

// TrimHistory keeps at most the newest MaxHistory samples.
func TrimHistory(samples []Sample) []Sample {
	if len(samples) <= MaxHistory {
		return samples
	}
	return append([]Sample(nil), samples[len(samples)-MaxHistory:]...)
}

// History returns the persisted sample history of the device.
func (s *Service) History(dev *device.Device) []Sample {
	var baseline Baseline
	if err := s.fs.ReadJSON(s.baselinePath(dev.ID), &baseline); err != nil {
		return nil
	}
	return append([]Sample(nil), baseline.Samples...)
}
