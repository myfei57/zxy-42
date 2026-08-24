package plan

// Compare orders maintenance work by clinical risk first, then due time.
// A negative result means a precedes b.
func Compare(a, b Item) int {
	if a.Risk.Weight() != b.Risk.Weight() {
		if a.Risk.Weight() > b.Risk.Weight() {
			return -1
		}
		return 1
	}
	switch {
	case a.Due.Before(b.Due):
		return -1
	case a.Due.After(b.Due):
		return 1
	}
	return 0
}
