package sonarlog

import (
	"fmt"
	"strconv"
	"strings"
)

// Representation:
// - the set is a bit vector
// - bits 0..30 represent GPUs, 0x0000_0000 is "empty"
// - if bit i is set then GPU i is in the set
// - value 0x8000_0000 represents "unknown"
//
// There will be *a lot* of these both in input and in memory, so:
// - representation compactness is important
// - avoiding pointers in the representation is important (or SonarReading will have pointers too)
// - avoiding a lot of garbage generation during parsing is important

type GpuSet uint32

const (
	unknown = GpuSet(0x8000_0000)
	empty   = GpuSet(0x0000_0000)
)

func EmptyGpuSet() GpuSet {
	return empty
}

func NewGpuSet(s string) (GpuSet, error) {
	gpuData := unknown
	if s != "unknown" {
		gpuData = empty
		if s != "none" {
			for {
				before, after, found := strings.Cut(s, ",")
				n, err := strconv.ParseUint(before, 10, 32)
				if err != nil {
					return unknown, fmt.Errorf("While parsing GPU list: %w", err)
				}
				if n > 30 {
					return unknown, fmt.Errorf("While parsing GPU list: GPU #%d", n)
				}
				gpuData |= (1 << n)
				if !found {
					break
				}
				s = after
			}
		}
	}
	return gpuData, nil
}

func (g GpuSet) IsEmpty() bool {
	return g == empty
}

func (g GpuSet) IsUnknown() bool {
	return g == unknown
}

func (g GpuSet) Size() int {
	if g == unknown {
		panic("Size of unknown set")
	}
	g = (g & 0x55555555) + ((g >> 1) & 0x55555555)
	g = (g & 0x33333333) + ((g >> 2) & 0x33333333)
	g = (g & 0x0f0f0f0f) + ((g >> 4) & 0x0f0f0f0f)
	g = (g & 0x00ff00ff) + ((g >> 8) & 0x00ff00ff)
	g = (g & 0x0000ffff) + ((g >> 16) & 0x0000ffff)
	return int(g)
}

func (g GpuSet) IsSet(n int) bool {
	if g == unknown {
		return false
	}
	return (g & (1 << n)) != 0
}

func (g GpuSet) AsSlice() []int {
	xs := make([]int, 0)
	if g != unknown {
		for k := 0; k < 31; k++ {
			if (g & (1 << k)) != 0 {
				xs = append(xs, k)
			}
		}
	}
	return xs
}

func (g GpuSet) String() string {
	if g == unknown {
		return "unknown"
	}
	if g == empty {
		return "none"
	}
	s := ""
	for k := 0; k < 31; k++ {
		if (g & (1 << k)) != 0 {
			if s != "" {
				s += ","
			}
			s += strconv.Itoa(k)
		}
	}
	return s
}
