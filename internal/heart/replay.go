package heart

// SeqRegression reports whether the beat repeats or goes backwards relative
// to the last accepted sequence.
func SeqRegression(last int64, beat Beat) bool {
	return beat.Seq <= last
}
