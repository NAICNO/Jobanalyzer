// Utilities for parsing values that are transmitted as JSON strings due to limitations in
// the sonalyze formatter.  These parsers uniformly panic on conversion error

package util

import (
	"fmt"
	"strconv"
	"time"
)

func JsonInt(s string) int {
	n, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON value to int, should not happen: %s", s))
	}
	return int(n)
}

func JsonDateTime(s string) time.Time {
	t, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON value to time, should not happen: %s", s))
	}
	return t
}

func JsonFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON value to f64, should not happen: %s", s))
	}
	return f
}
