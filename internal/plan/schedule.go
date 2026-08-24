package plan

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"medops/internal/device"
	"medops/internal/store"
)

var (
	ErrInvalidRisk       = errors.New("plan: invalid risk level")
	ErrPlanNotScheduled  = errors.New("plan: plan is not scheduled")
	ErrPlanNotInProgress = errors.New("plan: plan is not in progress")
)

type Status string

const (
	StatusScheduled  Status = "scheduled"
	StatusInProgress Status = "in-progress"
	StatusCompleted  Status = "completed"
)

// Plan is one maintenance work item.
type Plan struct {
	ID          string      `json:"id"`
	DeviceID    string      `json:"device_id"`
	Risk        device.Risk `json:"risk"`
	Status      Status      `json:"status"`
	Reason      string      `json:"reason"`
	DueAt       time.Time   `json:"due_at"`
	CreatedAt   time.Time   `json:"created_at"`
	CompletedAt time.Time   `json:"completed_at"`
}

type Service struct {
	fs    *store.FileStore
	plans map[string]*Plan
	order []string
}

func NewService(fs *store.FileStore) *Service {
	return &Service{fs: fs, plans: map[string]*Plan{}}
}

func (s *Service) Create(dev *device.Device, risk device.Risk, reason string, due, at time.Time) (*Plan, error) {
	if !risk.Valid() {
		return nil, ErrInvalidRisk
	}
	plan := &Plan{
		ID:        uuid.NewString(),
		DeviceID:  dev.ID,
		Risk:      risk,
		Status:    StatusScheduled,
		Reason:    reason,
		DueAt:     due,
		CreatedAt: at,
	}
	s.plans[plan.ID] = plan
	s.order = append(s.order, plan.ID)
	if err := s.Save(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) Get(id string) (*Plan, bool) {
	plan, ok := s.plans[id]
	return plan, ok
}

func (s *Service) All() []*Plan {
	plans := make([]*Plan, 0, len(s.order))
	for _, id := range s.order {
		if plan, ok := s.plans[id]; ok {
			plans = append(plans, plan)
		}
	}
	return plans
}

// Due returns the scheduled plans whose due moment is already reached.
func (s *Service) Due(at time.Time) []*Plan {
	var due []*Plan
	for _, plan := range s.All() {
		if plan.Status == StatusScheduled && !at.Before(plan.DueAt) {
			due = append(due, plan)
		}
	}
	return due
}

func (s *Service) Save(plan *Plan) error {
	return s.fs.WriteJSON("maintenance/plans/"+plan.ID+".json", plan)
}

func (s *Service) LoadAll() error {
	names, err := s.fs.List("maintenance/plans")
	if err != nil {
		return err
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		var plan Plan
		if err := s.fs.ReadJSON("maintenance/plans/"+name, &plan); err != nil {
			return err
		}
		if _, exists := s.plans[id]; !exists {
			s.plans[id] = &plan
			s.order = append(s.order, id)
		}
	}
	return nil
}
