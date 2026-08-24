package test

import (
	"time"

	"medops/internal/device"
	"medops/internal/qc"
	"medops/internal/store"
	"medops/internal/trend"
)

// Service executes QC tests for devices: reagent validity at execution time,
// calibration completion, method binding and threshold evaluation.
type Service struct {
	reagents     *ReagentRegistry
	calibrations *qc.CalibrationRegistry
	thresholds   *qc.ThresholdService
	reports      *store.ReportStore
	trend        *trend.Service
}

func NewService(
	reagents *ReagentRegistry,
	calibrations *qc.CalibrationRegistry,
	thresholds *qc.ThresholdService,
	reports *store.ReportStore,
	trendService *trend.Service,
) *Service {
	return &Service{
		reagents:     reagents,
		calibrations: calibrations,
		thresholds:   thresholds,
		reports:      reports,
		trend:        trendService,
	}
}

// Run executes a full QC test at the given execution moment.
func (s *Service) Run(plan *qc.Plan, dev *device.Device, method *qc.Method, parameter string, value float64, at time.Time) (*qc.Result, error) {
	reagent, ok := s.reagents.Get(plan.ReagentLot)
	if !ok {
		return nil, ErrReagentUnknown
	}
	if err := s.CheckForRun(plan, reagent, at); err != nil {
		return nil, err
	}
	cal, ok := s.calibrations.Get(plan.ID)
	if !ok || !cal.IsCompleted() {
		return nil, qc.ErrCalibrationPending
	}
	verdict, err := s.thresholds.Evaluate(plan, dev, parameter, value)
	if err != nil {
		return nil, err
	}
	result := qc.NewResult(plan.ID, dev.ID, parameter, value, verdict, method.Name, reagent.Lot, at)
	if err := s.SaveResult(result); err != nil {
		return nil, err
	}
	if s.trend != nil {
		if err := s.trend.Append(dev, value, parameter, at); err != nil {
			return nil, err
		}
	}
	return result, nil
}
