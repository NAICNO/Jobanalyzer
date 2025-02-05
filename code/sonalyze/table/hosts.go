package table

import (
	"slices"
	"strings"

	umaps "go-utils/maps"
)

type Hostnames struct {
	names  map[string]bool
	serial uint64
}

func NewHostnames() *Hostnames {
	return &Hostnames{names: make(map[string]bool)}
}

func (hn *Hostnames) IsEmpty() bool {
	return len(hn.names) == 0
}

func (hn *Hostnames) Add(name string) {
	hn.names[name] = true
	hn.serial++
}

func (hn *Hostnames) FormatFull() string {
	names := umaps.Keys(hn.names)
	slices.Sort(names)
	return strings.Join(names, ",")
}

func (hn *Hostnames) FormatBrief() string {
	short := make(map[string]bool)
	for _, n := range umaps.Keys(hn.names) {
		pre, _, _ := strings.Cut(n, ".")
		short[pre] = true
	}
	names := umaps.Keys(short)
	slices.Sort(names)
	return strings.Join(names, ",")
}
