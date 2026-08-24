package device

import "time"

// MoveWard reassigns the device to another ward and records the transfer.
// QC dispatch and trend evaluation must follow the current ward.
func (d *Device) MoveWard(wardID string, at time.Time) error {
	if wardID == "" {
		return ErrWardRequired
	}
	if wardID == d.WardID {
		return nil
	}
	d.WardHistory = append(d.WardHistory, WardMove{FromWard: d.WardID, ToWard: wardID, MovedAt: at})
	d.WardID = wardID
	return nil
}
