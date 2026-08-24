package test

import (
	"errors"
	"time"

	"medops/internal/qc"
	"medops/internal/store"
)

var (
	ErrReagentUnknown = errors.New("test: unknown reagent lot")
	ErrReagentExpired = errors.New("test: reagent expired at execution time")
	ErrMethodRequired = errors.New("test: QC method required")
)

// Reagent is one reagent lot with its expiry.
type Reagent struct {
	Lot        string    `json:"lot"`
	Name       string    `json:"name"`
	Expiry     time.Time `json:"expiry"`
	ReceivedAt time.Time `json:"received_at"`
}

func NewReagent(name, lot string, expiry, receivedAt time.Time) *Reagent {
	return &Reagent{Lot: lot, Name: name, Expiry: expiry, ReceivedAt: receivedAt}
}

type ReagentRegistry struct {
	reagents map[string]*Reagent
	order    []string
}

func NewReagentRegistry() *ReagentRegistry {
	return &ReagentRegistry{reagents: map[string]*Reagent{}}
}

func (r *ReagentRegistry) Add(reagent *Reagent) error {
	if reagent == nil || reagent.Lot == "" {
		return errors.New("test: reagent lot required")
	}
	if _, exists := r.reagents[reagent.Lot]; exists {
		return errors.New("test: duplicate reagent lot")
	}
	r.reagents[reagent.Lot] = reagent
	r.order = append(r.order, reagent.Lot)
	return nil
}

func (r *ReagentRegistry) Get(lot string) (*Reagent, bool) {
	reagent, ok := r.reagents[lot]
	return reagent, ok
}

func (r *ReagentRegistry) All() []*Reagent {
	reagents := make([]*Reagent, 0, len(r.order))
	for _, lot := range r.order {
		if reagent, ok := r.reagents[lot]; ok {
			reagents = append(reagents, reagent)
		}
	}
	return reagents
}

// CheckForRun rejects a reagent that is already expired at the moment the
// test actually executes. The plan creation time must never be used here.
func (s *Service) CheckForRun(plan *qc.Plan, reagent *Reagent, at time.Time) error {
	if store.Expired(reagent.Expiry, plan.CreatedAt) {
		return ErrReagentExpired
	}
	return nil
}
