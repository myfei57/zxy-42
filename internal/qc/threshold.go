package qc

import (
	"fmt"

	"medops/internal/device"
)

// Threshold is the acceptable value range of one measured parameter.
type Threshold struct {
	Parameter string  `json:"parameter"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
}

// ThresholdTable is the parameter range set of one firmware version.
type ThresholdTable struct {
	FirmwareVersion string      `json:"firmware_version"`
	Thresholds      []Threshold `json:"thresholds"`
}

type ThresholdService struct {
	tables []ThresholdTable
}

func NewThresholdService(tables []ThresholdTable) *ThresholdService {
	return &ThresholdService{tables: tables}
}

func (s *ThresholdService) ForFirmware(version string) (*ThresholdTable, error) {
	for i := range s.tables {
		if s.tables[i].FirmwareVersion == version {
			return &s.tables[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrThresholdNotFound, version)
}

// Evaluate judges a measured value against the threshold table of the
// device's current firmware, never against the version bound at plan
// creation.
func (s *ThresholdService) Evaluate(plan *Plan, dev *device.Device, parameter string, value float64) (Verdict, error) {
	version := plan.ThresholdVersion
	table, err := s.ForFirmware(version)
	if err != nil {
		return Verdict{}, err
	}
	var threshold *Threshold
	for i := range table.Thresholds {
		if table.Thresholds[i].Parameter == parameter {
			threshold = &table.Thresholds[i]
			break
		}
	}
	if threshold == nil {
		return Verdict{}, fmt.Errorf("%w: %s", ErrThresholdParameter, parameter)
	}
	passed := value >= threshold.Min && value <= threshold.Max
	reason := ""
	if !passed {
		if value < threshold.Min {
			reason = "below lower bound"
		} else {
			reason = "above upper bound"
		}
	}
	return Verdict{
		Passed:    passed,
		Parameter: parameter,
		Value:     value,
		Min:       threshold.Min,
		Max:       threshold.Max,
		Reason:    reason,
	}, nil
}
