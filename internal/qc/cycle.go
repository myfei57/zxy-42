package qc

import "time"

// NextDue computes the next QC due moment from a cycle anchor.
func NextDue(anchor time.Time, cycleDays int) time.Time {
	return anchor.AddDate(0, 0, cycleDays)
}

// Overdue reports whether the plan is already due at the given moment.
func Overdue(plan *Plan, at time.Time) bool {
	return !at.Before(plan.NextDue)
}
