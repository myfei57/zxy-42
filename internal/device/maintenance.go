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
