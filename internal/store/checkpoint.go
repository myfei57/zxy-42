package store

import "time"

// Checkpoint records the last heartbeat the platform durably accepted for a
// device, so recovery never re-processes older traffic.
type Checkpoint struct {
	DeviceID    string    `json:"device_id"`
	LastBeatSeq int64     `json:"last_beat_seq"`
	LastBeatAt  time.Time `json:"last_beat_at"`
	WrittenAt   time.Time `json:"written_at"`
}

type CheckpointStore struct {
	fs *FileStore
}

func NewCheckpointStore(fs *FileStore) *CheckpointStore {
	return &CheckpointStore{fs: fs}
}

func (c *CheckpointStore) Write(deviceID string, seq int64, at time.Time) error {
	return c.fs.WriteJSON("checkpoints/"+deviceID+".json", Checkpoint{
		DeviceID:    deviceID,
		LastBeatSeq: seq,
		LastBeatAt:  at,
		WrittenAt:   time.Now(),
	})
}

func (c *CheckpointStore) Read(deviceID string) (Checkpoint, bool) {
	var cp Checkpoint
	if err := c.fs.ReadJSON("checkpoints/"+deviceID+".json", &cp); err != nil {
		return Checkpoint{}, false
	}
	return cp, true
}
