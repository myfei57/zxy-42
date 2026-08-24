package audit

import "time"

// Filter narrows an audit listing by actor, action and time range.
type Filter struct {
	Actor  string
	Action string
	From   time.Time
	To     time.Time
}

func (f Filter) Matches(record Record) bool {
	if f.Actor != "" && record.Actor != f.Actor {
		return false
	}
	if f.Action != "" && record.Action != f.Action {
		return false
	}
	if !f.From.IsZero() && record.At.Before(f.From) {
		return false
	}
	if !f.To.IsZero() && record.At.After(f.To) {
		return false
	}
	return true
}

func (r *Recorder) Filter(f Filter) ([]Record, error) {
	records, err := r.List()
	if err != nil {
		return nil, err
	}
	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		if f.Matches(record) {
			filtered = append(filtered, record)
		}
	}
	return filtered, nil
}
