package common

import (
	"time"

	. "sonalyze/common"
)

// A Multihost is never empty - an empty Hosts must be resolved to a Multihost that contains all
// host names here.
//
// If any h pattern contains * then it is resolved against the available hosts here.

func ResolveWildcard(meta any, h Hosts, from, to time.Time) (Multihost, error) {
	// Code from config.go to handle an empty Hosts set, will disappear from there.  In this case,
	// must set the IsAll flag of the returned Multihost.  Note the set of hosts can be large, and
	// we'd like something efficient here - both caching and compression.  See comments in
	// config.go.  Probably the caching/compression logic should be there and we should not call
	// AvailableHosts directly.
	/*
		if h.IsEmpty() {
		 	// TODO: Make a cdp
			// TODO: Check if the cdp is not valid
			hosts, err := cdp.AvailableHosts(qa.FromDate, qa.ToDate)
			if err != nil {
				return nil, err
			}
			qa.Host = MultihostFromSingle(umaps.Keys(hosts)...)
	   		// TODO: Must set 'All' on the result object.
		}

	*/

	panic("FIXME: Implement")
}
