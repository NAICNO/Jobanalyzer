package common

import (
	"iter"
	"maps"
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
//
// TODO: At the moment, this does not have a representation that allows compressed names to be
// canonical.  See TODO comments below.
type Hosts struct {
	ranges   bool
	patterns []string
	globber  *hostglob.HostGlobber
	name     *atomic.Value // can be nil.  If not, holds nameInfo or (any)nil.
}

// The host names *must* be single names: No ranges or sets or *; names must not be empty; there
// must be no duplicates.  If a slice is passed, the caller must not retain it.  The API is for use
// only where those conditions are known to hold.
func NewHostsFromSingleInfallible(names ...string) Hosts {
	hosts, _ := NewHostsFromPatterns(names...)
	hosts.ranges = false
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
	// Globber compilation performs some syntax checking (but allows *).  In most cases, we're going
	// to want this globber anyway so it's not a disaster to construct it always.  But it could be
	// cached in the same way the canonicalName is.
	//
	// TODO: This must change in various ways.  The patterns must be compiled into HostnameSet
	// values that can be merged, we must merge them here, and then represent them directly, not as
	// a globber - the globber is probably obsolete at that point.
	globber, err := hostglob.NewGlobber(true, patterns)
	if err != nil {
		return Hosts{}, err
	}
	patterns = slices.Clone(patterns)
	return Hosts{
		ranges:   true,
		patterns: patterns,
		globber:  globber,
		name:     new(atomic.Value),
	}, nil
}

// The hostnames must be single-host names - no sets, no wildcards.  Returns a *canonical*
// multi-pattern for the set of hosts in the input.
func CompressHostnamesInfallible(hostnames ...string) string {
	h := NewHostsFromSingleInfallible(hostnames...)
	return h.CanonicalMultiname()
}

// Union a list of Hosts sets and return a fresh set.
func HostsUnion(hs []Hosts) Hosts {
	// TODO: This must change - merging must perform proper unioning of HostnameSet values
	// represented in the Hosts, see comment in NewHostsFromPatterns.
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
	return Hosts{
		ranges:   ranges,
		patterns: patterns,
		globber:  globber,
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
	// TODO: Not what we want, but this will probably be fixed by itself once NewHostsFromPatterns
	// and HostsUnion are changed, we'll probably just join the individual string conversions of
	// HostnameSet values here.
	compressed := hostglob.CompressHostnames(h.patterns)
	slices.Sort(compressed)
	n := strings.Join(compressed, ",")
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
	// We don't cache this currently b/c it's not used much.
	return h.patterns
}

// The Hosts must contain a single host name; return it.
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
	// Annoying that ExpandPattern returns a slice and not an iterator.
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

func (h *Hosts) Match(hostname string) bool {
	if h.IsAll() {
		return true
	}
	return h.globber.Match(hostname)
}

// Return true if the set of patterns is empty.
func (h *Hosts) IsAll() bool {
	if h.globber == nil {
		return true
	}
	return h.globber.IsEmpty()
}
