package common

import (
	"iter"
	"slices"
	"strings"

	"go-utils/hostglob"
)

// Hosts is a wrapper for a hostglob.HostGlobber that can be used to glob either straight host names
// or file names (based on a set of patterns), depending on need.
type Hosts struct {
	prefix   bool
	expanded bool                  // true iff patterns has only plain names, no ranges
	patterns []string              // never nil
	globber  *hostglob.HostGlobber // never nil
}

// Create a new Hosts from the list of patterns.  For pattern syntax, see the HostGlobber
// documentation.
func NewHosts(prefix bool, patterns []string) (*Hosts, error) {
	return newHosts(prefix, patterns, false)
}

func newHosts(prefix bool, patterns []string, knownSingle bool) (*Hosts, error) {
	if patterns == nil {
		patterns = make([]string, 0)
	}
	// NewGlobber performs the necessary syntax checking.  In most cases, we're going to want this
	// globber anyway so it's not a disaster to construct it always.
	globber, err := hostglob.NewGlobber(prefix, patterns)
	if err != nil {
		return nil, err
	}
	expanded := true
	if !knownSingle {
		for _, p := range patterns {
			if !IsPlainName(p) {
				expanded = false
				break
			}
		}
	}
	return &Hosts{
		prefix:   prefix,
		expanded: expanded,
		patterns: slices.Clone(patterns),
		globber:  globber,
	}, nil
}

func IsPlainName(n string) bool {
	return !strings.ContainsAny(n, "[]*")
}

func EmptyHosts() *Hosts {
	hosts, _ := newHosts(false, nil, true)
	return hosts
}

func SingleHostInfallible(hn string) *Hosts {
	if !IsPlainName(hn) {
		panic("BUG: Not a single host name: " + hn)
	}
	hs, _ := newHosts(true, []string{hn}, true)
	return hs
}

func ExpandedHostsInfallible(hns []string) *Hosts {
	for _, hn := range hns {
		if !IsPlainName(hn) {
			panic("BUG: Not a single host name: " + hn)
		}
	}
	hs, _ := newHosts(true, hns, true)
	return hs
}

func (h *Hosts) String() string {
	s := strings.Join(h.patterns, ",")
	if h.prefix {
		s = "prefix: " + s
	}
	return s
}

// Return true if the set of patterns is empty.
func (h *Hosts) IsEmpty() bool {
	return h.globber.IsEmpty()
}

func (h *Hosts) ExpandedPatterns() iter.Seq[string] {
	if h.expanded {
		return slices.Values(h.patterns)
	}
	ps := h.patterns
	return func(yield func(string) bool) {
		for _, p := range ps {
			if IsPlainName(p) {
				if !yield(p) {
					return
				}
			} else {
				// Oh boy.  This is pretty creaky - a mix of lazy and eager expansion, no caching,
				// etc.  There are several uses of ExpandPattern, all could probably be smarter.
				xs, err := hostglob.ExpandPattern(p)
				if err != nil {
					// The patterns comes from a compiled globber, so they are well-formed, but
					// ExpandPattern can't do "*", so expansion might still fail.
					continue
				}
				for _, hn := range xs {
					if !yield(hn) {
						return
					}
				}
			}
		}
	}
}

func (h *Hosts) SingleHostInfallible() string {
	if !h.expanded || len(h.patterns) != 1 {
		panic("BUG: Not a single host name in hosts bag")
	}
	return h.patterns[0]
}

func (h *Hosts) Patterns() []string {
	return h.patterns
}

func (h *Hosts) IsPrefix() bool {
	return h.prefix
}

// Return the cached globber that matches strings against the hosts in the set.
func (h *Hosts) HostnameGlobber() *hostglob.HostGlobber {
	return h.globber
}

// Create a new globber from the patterns by first inserting the host patterns into the globs and
// then creating the globber from that.  The globs are file name patterns with exactly one *.
// FilenameGlobber expands those patterns into sets of file name patterns by substituting the host
// name patterns that are in this Hosts object for the *, and returns a matcher that can be used to
// match against file names.
func (h *Hosts) FilenameGlobber(globs []string) *hostglob.HostGlobber {
	globbers := make([]*hostglob.HostGlobber, 0)
	for _, glob := range globs {
		if strings.Count(glob, "*") != 1 {
			panic("Host glob must have exactly one '*'")
		}
		before, after, _ := strings.Cut(glob, "*")
		globber, err := hostglob.NewGlobberWithFix(h.prefix, before, h.patterns, after)
		if err != nil {
			panic("Host glob compilation should not have failed")
		}
		globbers = append(globbers, globber)
	}
	globber := hostglob.Join(globbers)
	return globber
}
