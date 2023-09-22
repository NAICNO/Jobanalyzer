// Sundry date and time utilities.

package util

import (
	"time"
)

const (
	// The logs use the following format uniformly (for better or worse)
	DateTimeFormat = "2006-01-02 15:04"
)

func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
