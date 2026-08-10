package common

import (
	"iter"
	"slices"
	"strings"
	"sync/atomic"

	"go-utils/hostglob"
)

type nameInfo struct {
	name  string
	uname Ustr
}

// Hosts represents a set of host names from a cluster.  Hosts structures can be unioned.  One can
// also match host names against a Hosts set, or extract compressed canonical names or uncompressed
// names (individual host names) from it.
//
// For multi-pattern, pattern, and hostname syntax, see the go-utils/hostglob documentation.
type Hosts struct {
	// The patterns are always canonical, ie, disjoint: none of the patterns union with any of the
	// others.
	patterns []HostnameSet
	// Name can be nil.  If not, *name holds nameInfo | (any)nil.
	name     *atomic.Value
}

// The host names *must* be single names: No ranges or sets or *; names must not be empty; there
// must be no duplicates.  If a slice is passed, the caller must not retain it.
func NewHostsFromSingleInfallible(names ...string) Hosts {
	hosts, err := NewHostsFromPatterns(names...)
	if err != nil {
		panic("Internal: bad hostname(s)")
	}
	return hosts
}

// Create a new Hosts from a multi-pattern -- but no * wildcards are allowed!
func NewHostsFromMultiPattern(s string) (Hosts, error) {
	ps, err := hostglob.SplitMultiPattern(s)
	if err != nil {
		return Hosts{}, err
	}
	return NewHostsFromPatterns(ps...)
}

// Create a new Hosts from the list of patterns -- but not multi-patterns, and no * wildcards are
// allowed!
func NewHostsFromPatterns(patterns ...string) (Hosts, error) {
	parsed, err := parseConcretePatterns(patterns)
	if err != nil {
		return Hosts{}, err
	}
	merged := UnionHostnameSets(parsed)
	return Hosts{
		patterns: merged,
		name:     new(atomic.Value),
	}, nil
}

// Union a list of Hosts sets and return a fresh set.
func HostsUnion(hs []Hosts) Hosts {
	if len(hs) == 0 {
		panic("Empty set of hosts in merging")
	}
	l := 0
	for _, x := range hs {
		l += len(x.patterns)
	}
	patterns := make([]HostnameSet, 0, l)
	for _, x := range hs {
		patterns = append(patterns, x.patterns...)
	}
	merged := UnionHostnameSets(patterns)
	return Hosts{
		patterns: merged,
		name:     new(atomic.Value),
	}
}

func (h *Hosts) CanonicalMultiname() string {
	if h.name == nil {
		return ""
	}
	if v := h.name.Load(); v != nil {
		return v.(nameInfo).name
	}
	names := h.CanonicalNames()
	slices.Sort(names)
	n := strings.Join(names, ",")
	u := StringToUstr(n)
	h.name.Store(nameInfo{n, u})
	return n
}

func (h *Hosts) CanonicalMultinameUstr() Ustr {
	if h.name == nil {
		return UstrEmpty
	}
	if v := h.name.Load(); v != nil {
		return v.(nameInfo).uname
	}
	_ = h.CanonicalMultiname()
	return h.name.Load().(nameInfo).uname
}

// Return the string representations of the individual HostnameSets in the Hosts, that is, this is
// like CanonicalMultiname but without joining the resulting strings by ",".
func (h *Hosts) CanonicalNames() []string {
	// We don't cache this currently b/c it's not used much except via CanonicalMultiname, which
	// caches the result and more.
	names := make([]string, len(h.patterns))
	for i, p := range h.patterns {
		names[i] = p.String()
	}
	return names
}

// The Hosts must contain a single host name; return it.
func (h *Hosts) SingleNameInfallible() string {
	if len(h.patterns) != 1 || !h.patterns[0].IsSingle() {
		panic("Invalid use of SingleNameInfallible")
	}
	return h.patterns[0].String()
}

func (h *Hosts) ExpandNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, p := range h.patterns {
			for s := range p.Expand() {
				if !yield(s) {
					return
				}
			}
		}
	}
}

func (h *Hosts) Match(hostname string) bool {
	if h.IsAll() {
		return true
	}
	other, err := parseHostname(hostname)
	if err != nil {
		return false
	}
	for _, p := range h.patterns {
		if p.Match(other) {
			return true
		}
	}
	return false
}

// Return true if the Hosts is empty.
func (h *Hosts) IsAll() bool {
	return h.patterns == nil
}
