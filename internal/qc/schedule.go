package qc

import (
	"strings"
	"time"

	"medops/internal/device"
	"medops/internal/store"
)

// ScheduleService owns QC plans, their calibration sequences and the batch
// commit journal.
type ScheduleService struct {
	fs           *store.FileStore
	plans        map[string]*Plan
	order        []string
	thresholds   *ThresholdService
	methods      *MethodService
	calibrations *CalibrationRegistry
	batch        *store.BatchStore
}

func NewScheduleService(
	fs *store.FileStore,
	thresholds *ThresholdService,
	methods *MethodService,
	calibrations *CalibrationRegistry,
	batch *store.BatchStore,
) *ScheduleService {
	return &ScheduleService{
		fs:           fs,
		plans:        map[string]*Plan{},
		thresholds:   thresholds,
		methods:      methods,
		calibrations: calibrations,
		batch:        batch,
	}
}

// Create builds a draft plan for the device and binds a calibration sequence.
func (s *ScheduleService) Create(dev *device.Device, cycleDays int, reagentLot string, at time.Time) (*Plan, error) {
	thresholdVersion := dev.Firmware.VersionAt(at)
	methodVersion := dev.Firmware.Current()
	plan, err := NewPlan(dev, cycleDays, thresholdVersion, methodVersion, reagentLot, at)
	if err != nil {
		return nil, err
	}
	s.plans[plan.ID] = plan
	s.order = append(s.order, plan.ID)
	if err := s.Save(plan); err != nil {
		return nil, err
	}
	if err := s.calibrations.Put(NewCalibration(plan.ID, []string{"zero", "span", "linearity"})); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *ScheduleService) Activate(plan *Plan, at time.Time) error {
	if plan.Status != PlanStatusDraft {
		return ErrPlanNotActive
	}
	plan.Status = PlanStatusActive
	return s.Save(plan)
}

func (s *ScheduleService) Suspend(plan *Plan) error {
	if plan.Status != PlanStatusActive {
		return ErrPlanNotActive
	}
	plan.Status = PlanStatusSuspended
	return s.Save(plan)
}

func (s *ScheduleService) AdvanceCycle(plan *Plan, at time.Time) error {
	if plan.Status != PlanStatusActive {
		return ErrPlanNotActive
	}
	plan.LastRun = at
	plan.NextDue = NextDue(at, plan.CycleDays)
	return s.Save(plan)
}

func (s *ScheduleService) Due(at time.Time) []*Plan {
	var due []*Plan
	for _, plan := range s.All() {
		if plan.Status == PlanStatusActive && Overdue(plan, at) {
			due = append(due, plan)
		}
	}
	return due
}

func (s *ScheduleService) Get(id string) (*Plan, bool) {
	plan, ok := s.plans[id]
	return plan, ok
}

func (s *ScheduleService) All() []*Plan {
	plans := make([]*Plan, 0, len(s.order))
	for _, id := range s.order {
		if plan, ok := s.plans[id]; ok {
			plans = append(plans, plan)
		}
	}
	return plans
}

func (s *ScheduleService) Save(plan *Plan) error {
	return s.fs.WriteJSON("qc/plans/"+plan.ID+".json", plan)
}

func (s *ScheduleService) LoadAll() error {
	names, err := s.fs.List("qc/plans")
	if err != nil {
		return err
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		var plan Plan
		if err := s.fs.ReadJSON("qc/plans/"+name, &plan); err != nil {
			return err
		}
		if _, exists := s.plans[id]; !exists {
			s.plans[id] = &plan
			s.order = append(s.order, id)
		}
	}
	return nil
}
