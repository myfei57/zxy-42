package quota

import (
	"time"

	"medops/internal/store"
)

// Kind is the resource kind a quota limits.
type Kind string

const (
	KindHeartbeat Kind = "heartbeat"
	KindResult    Kind = "result"
)

// Quota is the current usage snapshot of one scope/kind pair.
type Quota struct {
	Scope       string    `json:"scope"`
	Kind        Kind      `json:"kind"`
	Limit       int       `json:"limit"`
	WindowStart time.Time `json:"window_start"`
	Used        int       `json:"used"`
}

// Service enforces heartbeat and result quotas per scope (device or ward).
type Service struct {
	fs     *store.FileStore
	quotas map[string]*Quota
	order  []string
	policy Policy
	ledger *ledger
}

func NewService(fs *store.FileStore, policy Policy) *Service {
	return &Service{
		fs:     fs,
		quotas: map[string]*Quota{},
		policy: policy,
		ledger: newLedger(fs),
	}
}

func (s *Service) key(scope string, kind Kind) string {
	return scope + "|" + string(kind)
}

func (s *Service) snapshot(scope string, kind Kind) Quota {
	q, ok := s.quotas[s.key(scope, kind)]
	if !ok {
		q = &Quota{
			Scope:       scope,
			Kind:        kind,
			Limit:       s.limit(kind),
			WindowStart: time.Now(),
			Used:        0,
		}
		s.quotas[s.key(scope, kind)] = q
		s.order = append(s.order, s.key(scope, kind))
	}
	return *q
}

func (s *Service) limit(kind Kind) int {
	if kind == KindHeartbeat {
		return s.policy.HeartbeatLimit
	}
	return s.policy.ResultLimit
}

func (s *Service) Snapshot(scope string, kind Kind) Quota {
	return s.snapshot(scope, kind)
}
