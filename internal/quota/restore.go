package quota

import "time"

// Restore refunds one unit of usage, used when a heartbeat or result was
// rejected after the quota had already been consumed.
func (s *Service) Restore(scope string, kind Kind, at time.Time) {
	q := s.snapshot(scope, kind)
	if q.Used > 0 {
		q.Used--
	}
	s.quotas[s.key(scope, kind)] = &q
	_ = s.ledger.Append(s.key(scope, kind), -1, at, "restore")
}
