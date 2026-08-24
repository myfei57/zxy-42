package audit

import "time"

// journalPath returns the audit journal file for the given day, so the
// journal rotates per day instead of growing without bound.
func journalPath(day time.Time) string {
	return "audit/journal-" + day.Format("20060102") + ".jsonl"
}
