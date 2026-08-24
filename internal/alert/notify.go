package alert

import (
	"errors"
	"time"
)

// Notification is one alert delivery to a ward console.
type Notification struct {
	AlertID string    `json:"alert_id"`
	Target  string    `json:"target"`
	SentAt  time.Time `json:"sent_at"`
}

func (s *Service) Notify(a *Alert, target string, at time.Time) (Notification, error) {
	if a == nil {
		return Notification{}, ErrAlertNotFound
	}
	if target == "" {
		return Notification{}, errors.New("alert: target required")
	}
	return Notification{AlertID: a.ID, Target: target, SentAt: at}, nil
}
