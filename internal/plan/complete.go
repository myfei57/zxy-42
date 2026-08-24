package plan

import (
	"time"

	"medops/internal/device"
)

// Complete finishes a maintenance intervention: the device lockout is cleared
// and the device is reactivated as one step, then the plan is marked completed.
func (s *Service) Complete(plan *Plan, dev *device.Device, at time.Time) error {
	if plan.Status != StatusInProgress {
		return ErrPlanNotInProgress
	}
	if err := dev.EndMaintenance(at); err != nil {
		return err
	}
	plan.Status = StatusCompleted
	plan.CompletedAt = at
	if err := s.Save(plan); err != nil {
		return err
	}
	record := device.MaintenanceRecord{
		DeviceID:  dev.ID,
		Reason:    plan.Reason,
		StartedAt: plan.CreatedAt,
		EndedAt:   at,
	}
	return s.fs.WriteJSON("maintenance/records/"+dev.ID+".json", record)
}

// Begin moves a scheduled plan into progress and applies the maintenance
// lockout on the device.
func (s *Service) Begin(plan *Plan, dev *device.Device, reason string, at time.Time) error {
	if plan.Status != StatusScheduled {
		return ErrPlanNotScheduled
	}
	if err := dev.BeginMaintenance(reason, at); err != nil {
		return err
	}
	plan.Status = StatusInProgress
	return s.Save(plan)
}
