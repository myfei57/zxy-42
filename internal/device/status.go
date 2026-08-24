package device

import "time"

func (s Status) Valid() bool {
	switch s {
	case StatusRegistered, StatusInUse, StatusMaintenance, StatusRetired:
		return true
	}
	return false
}

func (d *Device) SetStatus(next Status, at time.Time) error {
	if !next.Valid() {
		return ErrUnknownStatus
	}
	ok := false
	switch d.Status {
	case StatusRegistered:
		ok = next == StatusInUse || next == StatusRetired
	case StatusInUse:
		ok = next == StatusMaintenance || next == StatusRetired
	case StatusMaintenance:
		ok = next == StatusInUse || next == StatusRetired
	case StatusRetired:
		ok = false
	}
	if !ok {
		return ErrInvalidTransition
	}
	d.Status = next
	d.StatusChangedAt = at
	return nil
}

// Activate moves a registered or maintained device into clinical use. A
// device with an active maintenance lockout cannot be activated.
func (d *Device) Activate(at time.Time) error {
	if d.Lockout.IsActive() {
		return ErrLockoutActive
	}
	return d.SetStatus(StatusInUse, at)
}

func (d *Device) Retire(at time.Time) error {
	return d.SetStatus(StatusRetired, at)
}
