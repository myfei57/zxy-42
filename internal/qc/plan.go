package qc

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"medops/internal/device"
)

var (
	ErrInvalidCycleDays   = errors.New("qc: cycle days must be positive")
	ErrPlanNotActive      = errors.New("qc: plan is not active")
	ErrPlanSuspended      = errors.New("qc: plan is suspended")
	ErrThresholdNotFound  = errors.New("qc: no threshold table for firmware")
	ErrThresholdParameter = errors.New("qc: no threshold for parameter")
	ErrMethodNotFound     = errors.New("qc: no method for firmware")
	ErrNoCalibrationSteps = errors.New("qc: calibration sequence has no steps")
	ErrCalibrationPending = errors.New("qc: calibration sequence not completed")
)

type PlanStatus string

const (
	PlanStatusDraft     PlanStatus = "draft"
	PlanStatusActive    PlanStatus = "active"
	PlanStatusSuspended PlanStatus = "suspended"
)

// Plan is a device QC plan. The threshold and method versions are snapshots
// taken at plan creation for the audit trail; live evaluation always uses the
// device's current firmware.
type Plan struct {
	ID               string     `json:"id"`
	DeviceID         string     `json:"device_id"`
	WardID           string     `json:"ward_id"`
	WardSnapshot     string     `json:"ward_snapshot"`
	CycleDays        int        `json:"cycle_days"`
	ThresholdVersion string     `json:"threshold_version"`
	MethodVersion    string     `json:"method_version"`
	ReagentLot       string     `json:"reagent_lot"`
	NextDue          time.Time  `json:"next_due"`
	LastRun          time.Time  `json:"last_run"`
	CreatedAt        time.Time  `json:"created_at"`
	Status           PlanStatus `json:"status"`
	PendingRetry     []string   `json:"pending_retry"`
}

func NewPlan(dev *device.Device, cycleDays int, thresholdVersion, methodVersion, reagentLot string, at time.Time) (*Plan, error) {
	if cycleDays <= 0 {
		return nil, ErrInvalidCycleDays
	}
	return &Plan{
		ID:               uuid.NewString(),
		DeviceID:         dev.ID,
		WardID:           dev.Ward(),
		WardSnapshot:     dev.Ward(),
		CycleDays:        cycleDays,
		ThresholdVersion: thresholdVersion,
		MethodVersion:    methodVersion,
		ReagentLot:       reagentLot,
		NextDue:          NextDue(at, cycleDays),
		CreatedAt:        at,
		Status:           PlanStatusDraft,
	}, nil
}
