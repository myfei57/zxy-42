package device

import "medops/internal/heart"

// ApplyBeat applies an accepted heartbeat to the device state: the sequence
// must be strictly newer and the live clock follows the platform receive
// time so the online window tracks real traffic.
func (d *Device) ApplyBeat(beat heart.Beat) error {
	if beat.Seq <= d.LastBeatSeq {
		return ErrStaleBeat
	}
	d.LastHeartbeat = beat.SentAt
	d.LastBeatSeq = beat.Seq
	d.Clock.Sync(beat.ReceivedAt)
	return nil
}
