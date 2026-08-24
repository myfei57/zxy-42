package quota

import "time"

// Policy configures the quota limits and the sliding window length.
type Policy struct {
	HeartbeatLimit int
	ResultLimit    int
	Window         time.Duration
}

func DefaultPolicy() Policy {
	return Policy{
		HeartbeatLimit: 1000,
		ResultLimit:    200,
		Window:         24 * time.Hour,
	}
}
