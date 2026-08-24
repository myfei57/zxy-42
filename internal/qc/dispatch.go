package qc

import (
	"time"

	"medops/internal/device"
)

// Dispatch is one QC plan dispatch to a ward.
type Dispatch struct {
	PlanID       string    `json:"plan_id"`
	DeviceID     string    `json:"device_id"`
	TargetWard   string    `json:"target_ward"`
	DispatchedAt time.Time `json:"dispatched_at"`
}

// Dispatch sends the plan to the device's current ward. The dispatch target
// must follow ward moves, never a snapshot taken at plan creation.
func (s *ScheduleService) Dispatch(plan *Plan, dev *device.Device, at time.Time) (Dispatch, error) {
	if plan.Status == PlanStatusSuspended {
		return Dispatch{}, ErrPlanSuspended
	}
	return Dispatch{
		PlanID:       plan.ID,
		DeviceID:     dev.ID,
		TargetWard:   plan.WardSnapshot,
		DispatchedAt: at,
	}, nil
}
