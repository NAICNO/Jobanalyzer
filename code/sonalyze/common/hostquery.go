package common

import (
	"go-utils/hostglob"
)

// HostQuery is a box that holds user input.  The Patterns are separate host name patterns (ref the
// input grammar) that may contain * and must be resolved to concrete host sets by
// data/common.ResolveHostQuery before they are useful for querying data.
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
