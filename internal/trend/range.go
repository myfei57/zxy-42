package trend

import "medops/internal/device"

// Range is the acceptable value band of one device classification.
type Range struct {
	Classification device.Classification `json:"classification"`
	Min            float64               `json:"min"`
	Max            float64               `json:"max"`
}

// RangeFor resolves the acceptable range using the device's current
// classification. A reclassified device must be judged against the range of
// its new class, never the class snapshot at baseline creation.
func (s *Service) RangeFor(dev *device.Device) Range {
	classification := dev.CurrentClassification()
	switch classification {
	case device.ClassificationICU:
		return s.icu
	case device.ClassificationGeneral:
		return s.general
	}
	return s.general
}
