// Utilities for parsing values that are transmitted as JSON strings due to limitations in
// the sonalyze formatter.  These parsers uniformly panic on conversion error

package util

import (
	"fmt"
	"strconv"
	"strings"
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

func JsonGpulist(s string) []uint32 {
	var gpuData []uint32		// Unknown set
	if s != "unknown" {
		gpuData = make([]uint32, 0) // Empty set
		if s != "none" {
			for _, it := range strings.Split(s, ",") {
				n, err := strconv.ParseUint(it, 10, 32)
				if err != nil {
					panic(fmt.Sprintf("Failed to convert JSON value to gpu set, should not happen: %s", s))
				}
				gpuData = append(gpuData, uint32(n))
			}
		}
	}
	return gpuData
}

