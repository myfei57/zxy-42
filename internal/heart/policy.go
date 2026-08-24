package heart

import "time"

// Policy configures the heartbeat online window.
type Policy struct {
	Interval  time.Duration
	Tolerance time.Duration
}

func DefaultPolicy() Policy {
	return Policy{Interval: 30 * time.Minute, Tolerance: 2 * time.Minute}
}

func (p Policy) Window() *Window {
	return NewWindow(p.Interval, p.Tolerance)
}
