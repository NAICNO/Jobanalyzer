// Utilities for parsing values that are transmitted as JSON strings due to limitations in
// the sonalyze formatter.  These parsers uniformly panic on conversion error

package sonalyze

import (
	"fmt"
	"strconv"
	"time"

	"go-utils/sonarlog"
	gut "go-utils/time"
)

func JsonInt(s string) int {
	n, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON value to int, should not happen: %s", s))
	}
	return int(n)
}

func JsonDateTime(s string) time.Time {
	t, err := time.Parse(gut.CommonDateTimeFormat, s)
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

func JsonGpulist(s string) []uint32 {
	gpuData, err := sonarlog.ParseGpulist(s)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert JSON value to gpu set, should not happen: %s", err.Error()))
	}
	return gpuData
}
