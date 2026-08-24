package plan

import (
	"sort"
	"time"

	"medops/internal/device"
)

// Item is one maintenance queue entry used for priority ordering.
type Item struct {
	PlanID   string
	DeviceID string
	Risk     device.Risk
	Due      time.Time
}

// FromPlans converts maintenance plans into sortable queue items.
func FromPlans(plans []*Plan) []Item {
	items := make([]Item, 0, len(plans))
	for _, plan := range plans {
		if plan == nil {
			continue
		}
		items = append(items, Item{
			PlanID:   plan.ID,
			DeviceID: plan.DeviceID,
			Risk:     plan.Risk,
			Due:      plan.DueAt,
		})
	}
	return items
}

// SortItems orders the maintenance queue by clinical risk first and due time
// next, so high-risk devices are never pushed behind low-risk ones.
func SortItems(items []Item) []Item {
	sort.SliceStable(items, func(i, j int) bool {
		return Compare(items[i], items[j]) < 0
	})
	return items
}
