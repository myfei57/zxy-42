package device

// Risk is the clinical priority level of a device.
type Risk int

const (
	RiskLow Risk = iota + 1
	RiskMedium
	RiskHigh
)

func (r Risk) Valid() bool {
	return r >= RiskLow && r <= RiskHigh
}

func (r Risk) String() string {
	switch r {
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	}
	return "unknown"
}

// Weight orders risk so a higher value means higher clinical priority.
func (r Risk) Weight() int { return int(r) }
