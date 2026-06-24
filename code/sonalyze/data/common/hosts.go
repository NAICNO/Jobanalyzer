package common

import (
	"errors"
	"strings"
	"time"

	. "sonalyze/common"
)

// If any h pattern contains * then it is resolved against the available hosts here.

func ResolveHostQuery(meta any, h HostQuery, from, to time.Time) (Hosts, error) {
	// FIXME: Implement this properly for wildcards by querying AvailableHosts() as in the config
	// code and then filtering against those names.  For now, just return error if * is used.
	for _, p := range h.Patterns {
		if strings.IndexByte(p, '*') != -1 {
			return Hosts{}, errors.New("Wildcards not currently allowed: " + p)
		}
	}

	return NewHostsFromPatterns(h.Patterns...)
}
