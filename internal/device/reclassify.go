package device

import "time"

// Reclassify switches the device between ICU and general classification.
// Trend evaluation ranges must follow the new classification.
func (d *Device) Reclassify(c Classification, at time.Time) error {
	if c != ClassificationICU && c != ClassificationGeneral {
		return ErrUnknownClassification
	}
	if c == d.Classification {
		return nil
	}
	d.Classification = c
	d.ReclassifiedAt = at
	return nil
}
