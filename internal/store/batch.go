package store

import (
	"path/filepath"
)

// BatchStore persists the per-device results of a QC batch commit and the
// devices that still need a retry after a partial commit.
type BatchStore struct {
	fs *FileStore
}

func NewBatchStore(fs *FileStore) *BatchStore {
	return &BatchStore{fs: fs}
}

func (b *BatchStore) WriteResult(planID, deviceID string, payload interface{}) error {
	return b.fs.WriteJSON("results/batches/"+planID+"/"+deviceID+".json", payload)
}

func (b *BatchStore) MarkPending(planID string, deviceIDs []string) error {
	return b.fs.WriteJSON("results/batches/"+planID+"/pending.json", deviceIDs)
}

func (b *BatchStore) Pending(planID string) ([]string, bool) {
	var ids []string
	if err := b.fs.ReadJSON("results/batches/"+planID+"/pending.json", &ids); err != nil {
		return nil, false
	}
	return ids, true
}

func (b *BatchStore) ResultExists(planID, deviceID string) bool {
	return b.fs.Exists("results/batches/" + planID + "/" + deviceID + ".json")
}

// Path returns the absolute filesystem path a result file would land at.
func (b *BatchStore) Path(planID, deviceID string) string {
	return filepath.Join(b.fs.Root, "results", "batches", planID, deviceID+".json")
}
