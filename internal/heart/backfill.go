package heart

// RejectLate reports whether a heartbeat is older than the enrollment
// baseline: late arrivals must never backfill the online history.
func RejectLate(clock Clock, beat Beat) bool {
	return beat.SentAt.Before(clock.Baseline())
}
