package qc

import (
	"time"

	"github.com/google/uuid"
)

// Verdict is the threshold comparison result of one measured parameter.
type Verdict struct {
	Passed    bool    `json:"passed"`
	Parameter string  `json:"parameter"`
	Value     float64 `json:"value"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Reason    string  `json:"reason"`
}

// Result is one executed QC measurement for a device.
type Result struct {
	ID         string    `json:"id"`
	PlanID     string    `json:"plan_id"`
	DeviceID   string    `json:"device_id"`
	Parameter  string    `json:"parameter"`
	Value      float64   `json:"value"`
	Verdict    Verdict   `json:"verdict"`
	MethodName string    `json:"method_name"`
	ReagentLot string    `json:"reagent_lot"`
	ExecutedAt time.Time `json:"executed_at"`
	ReportPath string    `json:"report_path"`
}

func NewResult(planID, deviceID, parameter string, value float64, verdict Verdict, methodName, reagentLot string, executedAt time.Time) *Result {
	return &Result{
		ID:         uuid.NewString(),
		PlanID:     planID,
		DeviceID:   deviceID,
		Parameter:  parameter,
		Value:      value,
		Verdict:    verdict,
		MethodName: methodName,
		ReagentLot: reagentLot,
		ExecutedAt: executedAt,
	}
}
