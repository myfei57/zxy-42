package qc

// BatchItem is one device result inside a QC batch commit.
type BatchItem struct {
	DeviceID string  `json:"device_id"`
	Result   *Result `json:"result"`
}

// Batch is the unit of a QC round commit.
type Batch struct {
	Items []BatchItem `json:"items"`
}

func NewBatch(items []BatchItem) *Batch {
	return &Batch{Items: items}
}

func (b *Batch) Size() int { return len(b.Items) }

// RemainingIDs returns the device IDs from the given index onward, including
// the failed item itself, so a partial commit can keep them for retry.
func (b *Batch) RemainingIDs(from int) []string {
	if from < 0 {
		from = 0
	}
	if from >= len(b.Items) {
		return nil
	}
	ids := make([]string, 0, len(b.Items)-from)
	for _, item := range b.Items[from:] {
		ids = append(ids, item.DeviceID)
	}
	return ids
}
