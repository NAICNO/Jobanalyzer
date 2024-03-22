// These disappear when we move to Go 1.21 or later.

package minmax

func MaxInt(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

