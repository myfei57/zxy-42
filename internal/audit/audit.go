package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"medops/internal/store"
)

// Record is one operation audit entry.
type Record struct {
	ID     string    `json:"id"`
	Actor  string    `json:"actor"`
	Action string    `json:"action"`
	Target string    `json:"target"`
	Detail string    `json:"detail"`
	At     time.Time `json:"at"`
}

// Recorder appends audit entries to a file journal.
type Recorder struct {
	fs *store.FileStore
}

func NewRecorder(fs *store.FileStore) *Recorder {
	return &Recorder{fs: fs}
}

func (r *Recorder) Append(actor, action, target, detail string, at time.Time) (Record, error) {
	record := Record{
		ID:     uuid.NewString(),
		Actor:  actor,
		Action: action,
		Target: target,
		Detail: detail,
		At:     at,
	}
	line, err := json.Marshal(record)
	if err != nil {
		return Record{}, err
	}
	if err := r.fs.AppendLine(journalPath(at), line); err != nil {
		return Record{}, err
	}
	return record, nil
}

func (r *Recorder) List() ([]Record, error) {
	names, err := r.fs.List("audit")
	if err != nil {
		return nil, err
	}
	var records []Record
	for _, name := range names {
		lines, err := r.fs.ReadLines("audit/" + name)
		if err != nil {
			return nil, err
		}
		for _, line := range lines {
			var record Record
			if json.Unmarshal([]byte(line), &record) == nil {
				records = append(records, record)
			}
		}
	}
	return records, nil
}
