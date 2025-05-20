package common

import (
	"slices"
	"strings"

	"go-utils/hostglob"
)

// Hosts is a wrapper for a hostglob.HostGlobber that can be used to glob either straight host names
// or file names (based on a set of patterns), depending on need.
type Hosts struct {
	prefix   bool
	patterns []string
	globber  *hostglob.HostGlobber
}

// Create a new Hosts from the list of patterns.  For pattern syntax, see the HostGlobber
// documentation.
func NewHosts(prefix bool, patterns []string) (*Hosts, error) {
	// This performs the necessary syntax checking.  In most cases, we're going to want this globber
	// anyway so it's not a disaster to construct it always.
	globber, err := hostglob.NewGlobber(prefix, patterns)
	if err != nil {
		return nil, err
	}
	return &Hosts{
		prefix: prefix,
		patterns: slices.Clone(patterns),
		globber: globber,
	}, nil
}

// Return true if the set of patterns is empty.
func (h *Hosts) IsEmpty() bool {
	return h.globber.IsEmpty()
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
