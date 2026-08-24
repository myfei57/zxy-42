package alert

// Rule is one active alert threshold.
type Rule struct {
	Kind      string  `json:"kind"`
	Severity  string  `json:"severity"`
	Threshold float64 `json:"threshold"`
	Active    bool    `json:"active"`
}

func (r Rule) Matches(kind string, value float64) bool {
	return r.Active && r.Kind == kind && value > r.Threshold
}
