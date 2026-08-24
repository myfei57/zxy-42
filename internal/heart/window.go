package heart

import "time"

// Clock is the live time source the online window must follow.
type Clock interface {
	Current() time.Time
	Baseline() time.Time
	Advance(d time.Duration)
}

// Window decides whether a device is online using a sliding window anchored
// on the current clock, never on an earlier snapshot.
type Window struct {
	Interval  time.Duration
	Tolerance time.Duration
}

func NewWindow(interval, tolerance time.Duration) *Window {
	return &Window{Interval: interval, Tolerance: tolerance}
}

// Evaluate reports whether the beat falls inside the live clock window.
func (w *Window) Evaluate(clock Clock, beat Beat) bool {
	baseline := clock.Baseline()
	earliest := baseline.Add(-w.Interval)
	latest := baseline.Add(w.Tolerance)
	return !beat.SentAt.Before(earliest) && !beat.SentAt.After(latest)
}
