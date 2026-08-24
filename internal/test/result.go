package test

import "medops/internal/qc"

// SaveResult persists a QC result as a report file.
func (s *Service) SaveResult(result *qc.Result) error {
	rel, err := s.reports.Write(result.PlanID, result.DeviceID, result)
	if err != nil {
		return err
	}
	result.ReportPath = rel
	return nil
}
