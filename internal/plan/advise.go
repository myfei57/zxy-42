package plan

import (
	"time"

	"medops/internal/device"
	"medops/internal/trend"
)

// Advice is one recommended maintenance work item.
type Advice struct {
	PlanID      string      `json:"plan_id"`
	DeviceID    string      `json:"device_id"`
	Risk        device.Risk `json:"risk"`
	Recommended bool        `json:"recommended"`
	Reason      string      `json:"reason"`
	SuggestedAt time.Time   `json:"suggested_at"`
}

// Advise recommends maintenance when the device trend is out of range or the
// QC round is overdue, using the device's clinical risk as the due anchor.
func (s *Service) Advise(dev *device.Device, evaluation *trend.Evaluation, qcOverdue bool, at time.Time) (*Advice, error) {
	reason := ""
	recommended := false
	if evaluation != nil && !evaluation.WithinRange {
		recommended = true
		reason = "trend out of range"
	} else if qcOverdue {
		recommended = true
		reason = "qc round overdue"
	}
	if !recommended {
		return nil, nil
	}
	due := at.Add(time.Duration(dev.Risk.Weight()) * 24 * time.Hour)
	plan, err := s.Create(dev, dev.Risk, reason, due, at)
	if err != nil {
		return nil, err
	}
	return &Advice{
		PlanID:      plan.ID,
		DeviceID:    dev.ID,
		Risk:        dev.Risk,
		Recommended: true,
		Reason:      reason,
		SuggestedAt: at,
	}, nil
}
