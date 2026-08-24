package store

import "time"

// Expired reports whether a validity bound (such as a reagent expiry) is
// already past at the given reference moment. A moment exactly at the bound
// counts as expired.
func Expired(bound, at time.Time) bool {
	return at.After(bound) || at.Equal(bound)
}

// WindowEnd returns the moment a validity window that opened at start and
// lasts for the duration actually ends.
func WindowEnd(start time.Time, duration time.Duration) time.Time {
	return start.Add(duration)
}

// WithinWindow reports whether at falls inside [start, start+duration).
func WithinWindow(start time.Time, duration time.Duration, at time.Time) bool {
	return !at.Before(start) && at.Before(WindowEnd(start, duration))
}
