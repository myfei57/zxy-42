package heart

import (
	"errors"
	"time"
)

var (
	ErrInvalidBeat    = errors.New("heart: invalid beat")
	ErrReplayRejected = errors.New("heart: beat rejected as replay")
	ErrSeqRegression  = errors.New("heart: beat sequence regression")
	ErrOfflineWindow  = errors.New("heart: device outside online window")
)

// Beat is one heartbeat report from a device.
type Beat struct {
	DeviceID   string    `json:"device_id"`
	Seq        int64     `json:"seq"`
	SentAt     time.Time `json:"sent_at"`
	ReceivedAt time.Time `json:"received_at"`
	Battery    int       `json:"battery"`
	Rssi       int       `json:"rssi"`
}

func NewBeat(deviceID string, seq int64, sentAt, receivedAt time.Time) Beat {
	return Beat{
		DeviceID:   deviceID,
		Seq:        seq,
		SentAt:     sentAt,
		ReceivedAt: receivedAt,
		Battery:    -1,
		Rssi:       -127,
	}
}

func (b Beat) Validate() error {
	if b.DeviceID == "" || b.Seq <= 0 || b.SentAt.IsZero() || b.ReceivedAt.IsZero() {
		return ErrInvalidBeat
	}
	return nil
}
