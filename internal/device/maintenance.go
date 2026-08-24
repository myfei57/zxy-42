package device

import "time"

// MaintenanceRecord is the durable history of one maintenance intervention.
type MaintenanceRecord struct {
	DeviceID  string    `json:"device_id"`
	Reason    string    `json:"reason"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// BeginMaintenance moves an in-use device into maintenance and applies the
// lockout so it cannot return to clinical use before maintenance completes.
func (d *Device) BeginMaintenance(reason string, at time.Time) error {
	if err := d.SetStatus(StatusMaintenance, at); err != nil {
		return err
	}
	d.Lockout.Set(reason, at)
	return nil
}

// EndMaintenance releases the maintenance lockout and returns the device to
// clinical use in a single step. Clearing the lockout and reactivating the
// device must happen together: completing maintenance without releasing the
// lockout leaves the device stuck in maintenance with no path back to in-use,
// so the release and the status transition are not separable.
func (d *Device) EndMaintenance(at time.Time) error {
	d.Lockout.Clear()
	return d.Activate(at)
}
