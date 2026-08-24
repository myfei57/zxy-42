package trend

import "medops/internal/device"

// Latest returns the most recent sample of the device, if any.
func (s *Service) Latest(dev *device.Device) (Sample, bool) {
	var baseline Baseline
	if err := s.fs.ReadJSON(s.baselinePath(dev.ID), &baseline); err != nil {
		return Sample{}, false
	}
	if len(baseline.Samples) == 0 {
		return Sample{}, false
	}
	return baseline.Samples[len(baseline.Samples)-1], true
}
