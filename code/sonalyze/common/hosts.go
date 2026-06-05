// There is Hosts, which represents parsed host patterns coming from the user or an API request and
// which can contain names with * wildcards, and Multihost, which represents host patterns that
// don't have those wildcards - to transform the former to the latter, the wildcards have to be
// resolved against the hosts available in a time window.  This is done by ResolveWildcard.
//
// Then a Multihost can be used to match and enumerate host names in various contexts.  The Multihost
// is a wrapper for a hostglob.HostGlobber.
//
// A Multihost is never empty - an empty Hosts must be resolved to a Multihost that contains all
// host names.
//
// It's not obvious that Multihosts and Hostnames [sic] are not two aspects of the same idea.
//
// Previously Hosts served the purpose of Multihost but the semantics around wildcards are just
// plain wrong.
//
// Both Hosts and Multihost are immutable.  The zero values represent empty sets.
//
// TODO: Hosts should probably be renamed as HostPatterns and should just be a container for parsed
// patterns, with wildcards.
//
// TODO: Multihost construction should by itself not be able to error out: patterns that go in
// should all be known good.  It is only ResolveWildcard that should be able to fail.

package common

import (
	"iter"
	"maps"
	"slices"

	"go-utils/hostglob"
)

// var (
// 	// MT: stable after initialization
// 	emptyGlobber *hostglob.HostGlobber
// )

// func init() {
// 	emptyGlobber, _ = hostglob.NewGlobber(true, []string{})
// }

// We use "full" and "compressed" rather than "prefix" and "single" as flags so that a zero value
// has the most sensible flags.
type Hosts struct {
	// full       bool // full names, ie, !prefix
	// compressed bool // patterns may be compressed names
	patterns []string
	// globber    *hostglob.HostGlobber
}

// The host names *must* be single names: No ranges or sets or *; names must not be empty; there
// must be no duplicates.  If a slice is passed, the caller must not retain it.  The API is for use
// only where those conditions are known to hold.
// func NewHostsFromSingle(hns ...string) Hosts {
// 	prefix := true
// 	if len(hns) == 0 {
// 		return Hosts{}
// 	}
// 	globber, err := hostglob.NewGlobber(prefix, hns)
// 	if err != nil {
// 		panic(fmt.Sprintf("Internal error: bad host names: %v", hns))
// 	}
// 	slices.Sort(hns)
// 	return Hosts{
// 		full:     !prefix,
// 		patterns: hns,
// 		globber:  globber,
// 	}
// }

// Create a new Hosts from the list of patterns.  For pattern syntax, see the HostGlobber
// documentation.

func NewHostsFromMultiPatterns(multiPatterns ...string) (Hosts, error) {
	var patterns []string
	for _, mp := range multiPatterns {
		ps, err := hostglob.SplitMultiPattern(mp)
		if err != nil {
			return Hosts{}, err
		}
		patterns = append(patterns, ps...)
	}

	if len(patterns) == 0 {
		return Hosts{}, nil
	}

	for _, p := range patterns {
		if err := hostglob.SyntaxCheckPattern(p); err != nil {
			return Hosts{}, err
		}
	}

	return Hosts{patterns}, nil
	// // This performs the necessary syntax checking.  In most cases, we're going to want this globber
	// // anyway so it's not a disaster to construct it always.
	// globber, err := hostglob.NewGlobber(prefix, patterns)
	// if err != nil {
	// 	return Hosts{}, err
	// }
	// ps := slices.Clone(patterns)
	// slices.Sort(ps)
	// return Hosts{
	// 	full:       !prefix,
	// 	compressed: true,
	// 	patterns:   ps,
	// 	globber:    globber,
	// }, nil
}

// Return the canonical name for the set of hosts.  For an empty set, it is the empty string.  In
// all other cases it is in principle the compressed representation of the (mathematical) set of the
// expanded names in the set of patterns, ie, there are no overlapping ranges among the
// comma-separated patterns and every pattern is maximally compact according to our compression
// algorithm.  Canonicity is important for client code that uses the compressed host name as a hash
// table key for cross-node aggregated information (typically, multi-node sysinfo).
//
// As that is a fairly expensive computation, currently we depend on the patterns being sorted by
// the constructors and then we just join the patterns here.  Overlaps are not removed and names are
// not canonical if they were not canonical in the input.
//
// Nobody should be constructing multi-host names except through this method. (Actually not true:
// Some subsystems use the table.Hostnames structure which has reliable formatting code.  It would
// somewhat be better if we could use that here, but that code is fairly expensive and intended for
// use in queries.  What we want is a representation that can compactly represent ranges without
// expanding them.)
//
// I feel like a simple parse would be helpful here: following the hostglob grammar, and the
// limitations on compression in that code, a parse that returns [head] [range] [tail] where head
// and tail are literal and range is a set of nonnegative integers represented either singly or as
// a-b is sufficient (and minimally this degenerates to the head).  For a hostname like c1-6.xyz
// we'd get "c1-" {6} ".xyz", while for bling.local we'd just get "bling.local".  Then we can bucket
// this into sets based on matching head and tail and merge the number sets.  Then in the end if a
// set has only one element it is represented in its non-compressed form: "c1-6.xyz".  If the set
// has multiple elements it is "c1-[1-5,8].xyz".
//
// This amounts to first splitting multi-patterns into simple patterns, then splitting each pattern
// at the leftmost "." if any, then parsing right to left in the left piece to look for rightmost
// digit string or digit set, if any, and then constructing the "range" from that set if it is there.
//
// Notably after parsing and sorting by the head this *is* the canonical name, no use of
// CompressHostnames in its current form is necessary.
//
// To create the globber, one can create it on the list of stringified patterns I guess.

// func (h *Hosts) CanonicalName() string {
// 	if len(h.patterns) == 0 {
// 		return ""
// 	}
// 	s := strings.Join(h.patterns, ",")
// 	return s
// }

// func (h *Hosts) SingleNameInfallible() string {
// 	if h.compressed {
// 		panic("Compressed names in receiver of SingleNameInfallible")
// 	}
// 	if len(h.patterns) != 1 {
// 		panic("Not exactly 1 name in receiver of SingleNameInfallible")
// 	}
// 	return h.patterns[0]
// }

// func (h *Hosts) ExpandNames() iter.Seq[string] {
// 	if !h.compressed {
// 		return slices.Values(h.patterns)
// 	}
// 	// len(h.patterns) > 0
// 	// Annoying that hostglob.ExpandPattern returns a slice and not an iterator.
// 	return func(yield func(string) bool) {
// 		for _, p := range h.patterns {
// 			ss, err := hostglob.ExpandPattern(p)
// 			if err != nil {
// 				continue
// 			}
// 			for _, hn := range ss {
// 				if !yield(hn) {
// 					return
// 				}
// 			}
// 		}
// 	}
// }

// Return true if the set of patterns is empty.
// func (h *Hosts) IsEmpty() bool {
// 	if h.globber == nil {
// 		return true
// 	}
// 	return h.globber.IsEmpty()
// }

// func (h *Hosts) Patterns() []string {
// 	return h.patterns
// }

// func (h *Hosts) IsPrefix() bool {
// 	return !h.full
// }

// Return the cached globber that matches strings against the hosts in the set.  This will never
// return nil.
// func (h *Hosts) HostnameGlobber() *hostglob.HostGlobber {
// 	if h.globber == nil {
// 		return emptyGlobber
// 	}
// 	return h.globber
// }

// A Multihost is a container for an immutable set of concrete host names.  The set has a canonical
// name (string representation), which is a sorted list of the compressed names of the patterns (in
// the sense of the hostglob grammar) in the set.
//
// Normally these will be small because they are constructed by stream merging and stream merging
// normally operates on small sets.
//
// It's not at all obvious why this should not be like (or just be) table.Hostnames.
//
// The Multihost should never be empty.
//
// 'All' is strictly an optimization flag to avoid filtering a result set further against the set of
// host names.  If it is set, all hosts are still present in the Multihost, and filtering against
// the Multihost will produce the same set whether IsAll is tested or not.  But a producer that knows
// it's including all hosts can set it, and this may help keep things efficient.

type Multihost struct {
	complex  bool
	patterns []string // never nil
	All      bool
}

// Arguments *must* be single hostnames without * or ranges.  If a slice is passed, ownership of the
// slice is passed to the returned object.
func MultihostFromSingle(hostnames ...string) Multihost {
	return Multihost{
		patterns: hostnames,
	}
}

// Arguments *must* be syntactically valid host name patterns without *, and with ranges only in the
// first element.  (Actually we normally require at most a single range.)  They cannot be
// multi-patterns.  If a slice is passed, ownership of the slice is passed to the returned object.
func MultihostFromNonWildcardPattern(pattern ...string) Multihost {
	if len(pattern) == 0 {
		panic("Empty set of Multihost patterns")
	}
	return Multihost{
		complex:  true,
		patterns: pattern,
	}
}

func MultihostMerge(ms []Multihost) Multihost {
	if len(ms) == 0 {
		panic("Empty set of multihosts in merging")
	}
	// Remove exact duplicates because it's easy.  That may be redundant, will examine later.
	patterns := make(map[string]bool, 0)
	var complex bool
	for _, mh := range ms {
		for _, p := range mh.patterns {
			patterns[p] = true
		}
		complex = complex || mh.complex
	}
	return Multihost{
		complex:  complex,
		patterns: slices.Collect(maps.Keys(patterns)),
	}
}

func (m *Multihost) SingleNameInfallible() string {
	if len(m.patterns) != 1 || m.complex {
		panic("SingleNameInfallible failed")
	}
	return m.patterns[0]
}

func (m *Multihost) Match(hn string) bool {
	if m.All {
		return true
	}
	panic("FIXME: Implement")
}

// For a given set of hosts, NO MATTER HOW THEY WERE REPRESENTED WHEN THE SET WAS CREATED, this must
// always return the same string.  The canonical name is used to index metadata for a merged host's
// data stream, among other things.  See extensive comment above (should be moved here).
func (m *Multihost) CanonicalName() string {
	panic("FIXME: Implement")
}

func (m *Multihost) ExpandNames() iter.Seq[string] {
	panic("FIXME: Implement")
}

// This lends the slice: the slice or the underlying array in the range of the slice must not be
// modified.  The return value is never nil.  This may or may not return the input patterns, but
// no redundant patterns will have been added.
func (m *Multihost) Patterns() []string {
	return m.patterns
}

func (m *Multihost) HostnameGlobber() *hostglob.HostGlobber {
	// This can be lazy since a number of commands do not need it, but at the same time, anything
	// that depends on samples will need it, so there may not be much to be gained by not just
	// constructing it always.
	panic("FIXME: Implement")
}
