package quota

import (
	"errors"
	"time"

	"medops/internal/store"
)

var ErrQuotaExceeded = errors.New("quota: limit exceeded")

// Consume records one unit of usage for the scope/kind pair. The quota window
// slides when the usage falls outside the current window.
func (s *Service) Consume(scope string, kind Kind, at time.Time) error {
	q := s.snapshot(scope, kind)
	if !store.WithinWindow(q.WindowStart, s.policy.Window, at) {
		q.WindowStart = at
		q.Used = 0
	}
	if q.Used >= q.Limit {
		return ErrQuotaExceeded
	}
	q.Used++
	s.quotas[s.key(scope, kind)] = &q
	return s.ledger.Append(s.key(scope, kind), +1, at, "consume")
}
