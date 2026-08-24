package test

import (
	"fmt"
	"time"

	"medops/internal/device"
	"medops/internal/qc"
)

// Report is a generated QC report summary.
type Report struct {
	PlanID      string    `json:"plan_id"`
	DeviceID    string    `json:"device_id"`
	Summary     string    `json:"summary"`
	ResultCount int       `json:"result_count"`
	GeneratedAt time.Time `json:"generated_at"`
}

// GenerateReport builds a report summary from committed results and writes it
// to the report store.
func (s *Service) GenerateReport(plan *qc.Plan, dev *device.Device, results []*qc.Result, at time.Time) (*Report, error) {
	passed := 0
	for _, result := range results {
		if result != nil && result.Verdict.Passed {
			passed++
		}
	}
	report := &Report{
		PlanID:      plan.ID,
		DeviceID:    dev.ID,
		Summary:     fmt.Sprintf("%d/%d parameters passed", passed, len(results)),
		ResultCount: len(results),
		GeneratedAt: at,
	}
	if _, err := s.reports.Write(plan.ID, dev.ID+".report", report); err != nil {
		return nil, err
	}
	return report, nil
}
