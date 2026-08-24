package test

import (
	"time"

	"medops/internal/device"
	"medops/internal/qc"
)

// RunSample measures a sample against the calibration reference. The sample
// must never run before the calibration sequence completes.
func (s *Service) RunSample(plan *qc.Plan, dev *device.Device, cal *qc.Calibration, method *qc.Method, value float64, at time.Time) (*qc.Result, error) {
	if method == nil || len(method.Parameters) == 0 {
		return nil, ErrMethodRequired
	}
	verdict, err := s.thresholds.Evaluate(plan, dev, method.Parameters[0], value)
	if err != nil {
		return nil, err
	}
	result := qc.NewResult(plan.ID, dev.ID, method.Parameters[0], value, verdict, method.Name, plan.ReagentLot, at)
	if err := s.SaveResult(result); err != nil {
		return nil, err
	}
	if s.trend != nil {
		if err := s.trend.Append(dev, value, method.Parameters[0], at); err != nil {
			return nil, err
		}
	}
	return result, nil
}
