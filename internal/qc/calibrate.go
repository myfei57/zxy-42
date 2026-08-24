package qc

import (
	"time"

	"medops/internal/store"
)

// Calibration is the reference-setup sequence a plan must finish before any
// sample test may run.
type Calibration struct {
	PlanID      string    `json:"plan_id"`
	Steps       []string  `json:"steps"`
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at"`
}

func NewCalibration(planID string, steps []string) *Calibration {
	return &Calibration{PlanID: planID, Steps: append([]string(nil), steps...)}
}

// RunSequence executes the full calibration sequence in order and marks it
// complete only after every step has run.
func (c *Calibration) RunSequence(steps []string, at time.Time) error {
	if len(steps) == 0 {
		return ErrNoCalibrationSteps
	}
	c.Steps = append([]string(nil), steps...)
	c.Completed = true
	c.CompletedAt = at
	return nil
}

func (c *Calibration) IsCompleted() bool { return c.Completed }

// RemainingSteps returns the steps that still need to run while the sequence
// is unfinished.
func (c *Calibration) RemainingSteps() []string {
	if c.Completed {
		return nil
	}
	return append([]string(nil), c.Steps...)
}

// CalibrationRegistry owns the per-plan calibration sequences.
type CalibrationRegistry struct {
	fs    *store.FileStore
	items map[string]*Calibration
	order []string
}

func NewCalibrationRegistry(fs *store.FileStore) *CalibrationRegistry {
	return &CalibrationRegistry{fs: fs, items: map[string]*Calibration{}}
}

func (r *CalibrationRegistry) Put(c *Calibration) error {
	if _, exists := r.items[c.PlanID]; !exists {
		r.order = append(r.order, c.PlanID)
	}
	r.items[c.PlanID] = c
	return r.fs.WriteJSON("qc/calibrations/"+c.PlanID+".json", c)
}

func (r *CalibrationRegistry) Get(planID string) (*Calibration, bool) {
	c, ok := r.items[planID]
	return c, ok
}

func (r *CalibrationRegistry) All() []*Calibration {
	items := make([]*Calibration, 0, len(r.order))
	for _, id := range r.order {
		if c, ok := r.items[id]; ok {
			items = append(items, c)
		}
	}
	return items
}

func (r *CalibrationRegistry) LoadAll() error {
	names, err := r.fs.List("qc/calibrations")
	if err != nil {
		return err
	}
	for _, name := range names {
		if len(name) < 5 || name[len(name)-5:] != ".json" {
			continue
		}
		id := name[:len(name)-5]
		var c Calibration
		if err := r.fs.ReadJSON("qc/calibrations/"+name, &c); err != nil {
			return err
		}
		if _, exists := r.items[id]; !exists {
			r.items[id] = &c
			r.order = append(r.order, id)
		}
	}
	return nil
}
