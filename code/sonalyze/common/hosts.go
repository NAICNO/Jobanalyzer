package common

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"go-utils/hostglob"
)

// Hosts is a wrapper for a hostglob.HostGlobber that can be used to glob either straight host names
// or file names (based on a set of patterns), depending on need.
type Hosts struct {
	ranges   bool
	patterns []string
	globber  *hostglob.HostGlobber
}

var (
	// MT: immutable after initialization
	emptyGlobber *hostglob.HostGlobber
)

func init() {
	emptyGlobber, _ = hostglob.NewGlobber(false, []string{})
}

// The host names *must* be single names: No ranges or sets or *; names must not be empty; there
// must be no duplicates.  If a slice is passed, the caller must not retain it.  The API is for use
// only where those conditions are known to hold.
func NewHostsFromSingle(names ...string) Hosts {
	hosts, _ := NewHostsFromPatterns(names...)
	hosts.ranges = false
	return hosts
}

// Create a new Hosts from the list of patterns -- but not multi-patterns, and no * wildcards are
// allowed!  For pattern syntax, see the HostGlobber documentation.
func NewHostsFromPatterns(patterns ...string) (Hosts, error) {
	// Globber compilation performs some syntax checking (but allows *).  In most cases, we're going
	// to want this globber anyway so it's not a disaster to construct it always.
	globber, err := hostglob.NewGlobber(true, patterns)
	if err != nil {
		return Hosts{}, err
	}
	patterns = slices.Clone(patterns)
	// FIXME: Sorting is insufficient for canonicalizing names, but it's a start.
	slices.Sort(patterns)
	return Hosts{
		ranges:   true,
		patterns: patterns,
		globber:  globber,
	}, nil
}

func HostsMerge(hs []Hosts) Hosts {
	if len(hs) == 0 {
		panic("Empty set of hosts in merging")
	}
	uniquePatterns := make(map[string]bool, 0)
	var ranges bool
	for _, h := range hs {
		for _, p := range h.patterns {
			uniquePatterns[p] = true
		}
		ranges = ranges || h.ranges
	}
	patterns := slices.Collect(maps.Keys(uniquePatterns))
	globber, _ := hostglob.NewGlobber(true, patterns)
	// FIXME: Sorting is insufficient for canonicalizing names, but it's a start.
	slices.Sort(patterns)
	return Hosts{
		ranges:   ranges,
		patterns: patterns,
		globber:  globber,
	}
}

func (h *Hosts) CanonicalName() string {
	// FIXME: Must be cached.
	compressed := hostglob.CompressHostnames(h.patterns)
	slices.Sort(compressed)
	return strings.Join(compressed, ",")
}

func (h *Hosts) CanonicalNameUstr() Ustr {
	// FIXME: Should be cached maybe.
	return StringToUstr(h.CanonicalName())
}

func (h *Hosts) SingleNameInfallible() string {
	if h.ranges || len(h.patterns) != 1 {
		panic("Invalid use of SingleNameInfallible")
	}
	return h.patterns[0]
}

func (h *Hosts) ExpandNames() iter.Seq[string] {
	if !h.ranges {
		return slices.Values(h.patterns)
	}
	// Annoying that hostglob.ExpandPattern returns a slice and not an iterator.
	return func(yield func(string) bool) {
		for _, p := range h.patterns {
			ss, err := hostglob.ExpandPattern(p)
			if err != nil {
				continue
			}
			for _, hn := range ss {
				if !yield(hn) {
					return
				}
			}
		}
	}
}

// Return true if the set of patterns is empty.
func (h *Hosts) IsEmpty() bool {
	if h.globber == nil {
		return true
	}
	return h.globber.IsEmpty()
}

func (h *Hosts) Patterns() []string {
	return h.patterns
}

// Return the cached globber that matches strings against the hosts in the set.
func (h *Hosts) HostnameGlobber() *hostglob.HostGlobber {
	if h.globber == nil {
		return emptyGlobber
	}
	return h.globber
}

// The HostQuery is a box that holds user input.  These are separate patterns but they may contain *
// and must be resolved to concrete host sets by data/common.ResolveHostQuery before they are useful
// for querying data.

type HostQuery struct {
	Patterns []string
}

func NewHostQueryFromMultiPatterns(multiPatterns ...string) (HostQuery, error) {
	var patterns []string
	for _, mp := range multiPatterns {
		ps, err := hostglob.SplitMultiPattern(mp)
		if err != nil {
			return HostQuery{}, err
		}
		patterns = append(patterns, ps...)
	}
	if len(patterns) == 0 {
		return HostQuery{}, nil
	}
	for _, p := range patterns {
		if err := hostglob.SyntaxCheckPattern(p); err != nil {
			return HostQuery{}, err
		}
	}
	return HostQuery{patterns}, nil
}
