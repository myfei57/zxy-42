package qc_test

import (
	"testing"
	"time"

	"medops/internal/qc"
)

func TestCycleNextDueAndOverdue(t *testing.T) {
	anchor := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	next := qc.NextDue(anchor, 7)
	if !next.Equal(anchor.AddDate(0, 0, 7)) {
		t.Fatalf("next due mismatch: %v", next)
	}
	plan := &qc.Plan{NextDue: next, Status: qc.PlanStatusActive}
	if qc.Overdue(plan, anchor) {
		t.Fatal("plan before due reported overdue")
	}
	if !qc.Overdue(plan, anchor.AddDate(0, 0, 8)) {
		t.Fatal("plan after due not reported overdue")
	}
}
