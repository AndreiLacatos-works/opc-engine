package valuecomputers

func mapValueToNewRange(originalRangeStart, originalRangeEnd, value, newRangeStart, newRangeEnd float64) float64 {
	originalDelta := originalRangeEnd - originalRangeStart
	if originalDelta == 0 {
		return 0.0
	}

	proportion := (value - originalRangeStart) / originalDelta
	return newRangeStart + proportion*(newRangeEnd-newRangeStart)
}
