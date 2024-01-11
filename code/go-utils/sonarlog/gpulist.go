package sonarlog

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseGpulist(s string) ([]uint32, error) {
	var gpuData []uint32 // Unknown set
	if s != "unknown" {
		gpuData = make([]uint32, 0) // Empty set
		if s != "none" {
			for _, it := range strings.Split(s, ",") {
				n, err := strconv.ParseUint(it, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("While parsing GPU list: %w", err)
				}
				gpuData = append(gpuData, uint32(n))
			}
		}
	}
	return gpuData, nil
}
