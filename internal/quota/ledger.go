package quota

import (
	"encoding/json"
	"time"

	"medops/internal/store"
)

// LedgerEntry is one usage change recorded for auditability.
type LedgerEntry struct {
	Key    string    `json:"key"`
	Delta  int       `json:"delta"`
	At     time.Time `json:"at"`
	Reason string    `json:"reason"`
}

type ledger struct {
	fs *store.FileStore
}

func newLedger(fs *store.FileStore) *ledger {
	return &ledger{fs: fs}
}

func (l *ledger) Append(key string, delta int, at time.Time, reason string) error {
	line, err := json.Marshal(LedgerEntry{Key: key, Delta: delta, At: at, Reason: reason})
	if err != nil {
		return err
	}
	return l.fs.AppendLine("quota/ledger.jsonl", line)
}

func (l *ledger) Entries(key string) ([]LedgerEntry, error) {
	lines, err := l.fs.ReadLines("quota/ledger.jsonl")
	if err != nil {
		return nil, err
	}
	var entries []LedgerEntry
	for _, line := range lines {
		var entry LedgerEntry
		if json.Unmarshal([]byte(line), &entry) == nil && entry.Key == key {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// Ledger returns the recorded usage changes for one scope/kind pair.
func (s *Service) Ledger(scope string, kind Kind) ([]LedgerEntry, error) {
	return s.ledger.Entries(s.key(scope, kind))
}
