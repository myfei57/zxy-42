package heart

import (
	"time"

	"medops/internal/store"
)

// Verdict is the outcome of one heartbeat evaluation.
type Verdict struct {
	DeviceID string    `json:"device_id"`
	Online   bool      `json:"online"`
	LastBeat time.Time `json:"last_beat"`
	Baseline time.Time `json:"baseline"`
}

// Beater is anything that can receive an accepted heartbeat.
type Beater interface {
	DeviceID() string
	LiveClock() Clock
	LastSeq() int64
	ApplyBeat(Beat) error
}

// Service accepts heartbeats, rejects replays and stale sequences, applies
// the online window and persists a checkpoint for recovery.
type Service struct {
	window     *Window
	checkpoint *store.CheckpointStore
}

func NewService(window *Window, checkpoint *store.CheckpointStore) *Service {
	return &Service{window: window, checkpoint: checkpoint}
}

func (s *Service) Record(b Beater, beat Beat) (Verdict, error) {
	if err := beat.Validate(); err != nil {
		return Verdict{}, err
	}
	if RejectLate(b.LiveClock(), beat) {
		return Verdict{}, ErrReplayRejected
	}
	if SeqRegression(b.LastSeq(), beat) {
		return Verdict{}, ErrSeqRegression
	}
	if !s.window.Evaluate(b.LiveClock(), beat) {
		return Verdict{}, ErrOfflineWindow
	}
	if delta := beat.ReceivedAt.Sub(b.LiveClock().Current()); delta > 0 {
		b.LiveClock().Advance(delta)
	}
	if err := b.ApplyBeat(beat); err != nil {
		return Verdict{}, err
	}
	if s.checkpoint != nil {
		_ = s.checkpoint.Write(b.DeviceID(), beat.Seq, beat.ReceivedAt)
	}
	return Verdict{
		DeviceID: b.DeviceID(),
		Online:   true,
		LastBeat: beat.SentAt,
		Baseline: b.LiveClock().Current(),
	}, nil
}

// LastCheckpoint returns the durably accepted heartbeat checkpoint of a
// device, when one exists.
func (s *Service) LastCheckpoint(deviceID string) (store.Checkpoint, bool) {
	if s.checkpoint == nil {
		return store.Checkpoint{}, false
	}
	return s.checkpoint.Read(deviceID)
}
