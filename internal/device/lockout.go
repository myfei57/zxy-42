package device

import "time"

// Lockout is the maintenance exclusion state of a device. A locked device is
// not allowed back into clinical use until the lockout is cleared.
type Lockout struct {
	Active bool      `json:"active"`
	Reason string    `json:"reason"`
	SetAt  time.Time `json:"set_at"`
}

func NewLockout() *Lockout {
	return &Lockout{}
}

func (l *Lockout) Set(reason string, at time.Time) {
	l.Active = true
	l.Reason = reason
	l.SetAt = at
}

func (l *Lockout) Clear() {
	l.Active = false
	l.Reason = ""
	l.SetAt = time.Time{}
}

func (l *Lockout) IsActive() bool { return l.Active }
