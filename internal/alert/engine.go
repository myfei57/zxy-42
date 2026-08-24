package alert

import (
	"fmt"
	"time"

	"medops/internal/device"
)

// Evaluate raises an alert for every active rule matched by the measured
// value and returns the raised alerts.
func (s *Service) Evaluate(dev *device.Device, kind string, value float64, rules []Rule, at time.Time) ([]*Alert, error) {
	var raised []*Alert
	for _, rule := range rules {
		if !rule.Matches(kind, value) {
			continue
		}
		message := fmt.Sprintf("%s exceeded %.2f", kind, rule.Threshold)
		a, err := s.Raise(dev, kind, rule.Severity, message, at)
		if err != nil {
			return raised, err
		}
		raised = append(raised, a)
	}
	return raised, nil
}
