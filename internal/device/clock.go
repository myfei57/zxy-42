package device

import "time"

// Clock is the device-scoped time source. The live time (Current) is what
// every online-window evaluation must follow; Baseline is the enrollment
// anchor used to reject pre-enrollment replays.
type Clock struct {
	Now    time.Time `json:"now"`
	Anchor time.Time `json:"baseline"`
}

func NewClock(at time.Time) *Clock {
	return &Clock{Now: at, Anchor: at}
}

func (c *Clock) Current() time.Time { return c.Now }

func (c *Clock) Baseline() time.Time { return c.Anchor }

func (c *Clock) Advance(d time.Duration) { c.Now = c.Now.Add(d) }

func (c *Clock) Sync(at time.Time) { c.Now = at }

// Reanchor shifts both the baseline and the live time after a clock
// correction; heartbeats before the new anchor are treated as replays.
func (c *Clock) Reanchor(at time.Time) {
	c.Anchor = at
	c.Now = at
}

// Elapsed returns how long the device has been enrolled.
func (c *Clock) Elapsed() time.Duration { return c.Now.Sub(c.Anchor) }
