package cmd

import (
	"fmt"

	. "sonalyze/common"
	"sonalyze/data/config"
	"sonalyze/data/sample"
)

// Standard method for cleaning an InputStreamSet relative to a config: the config must have a
// definition for each host.  No config at all is a fatal error.  No config for a host means we
// remove the host from the set, we return the modified set.
//
// Over time this may become more complicated, as the config becomes time-dependent.
func EnsureConfigForInputStreams(
	cfg *config.ConfigDataProvider,
	streams sample.InputStreamSet,
	reason string,
) (sample.InputStreamSet, error) {
	// Remove streams for which we have no config data.
	bad := make(map[sample.InputStreamKey]bool)
	for key, stream := range streams {
		hn := (*stream)[0].Hostname
		ts := (*stream)[0].Timestamp
		if cfg.LookupHostByTime(hn, ts) == nil {
			bad[key] = true
			Log.Infof("Warning: Missing host configuration for %s", hn.String())
		}
	}

	for b := range bad {
		delete(streams, b)
	}

	// Bail if there's no config data at all.
	if len(streams) == 0 {
		return nil, fmt.Errorf("All configuration data missing: %s", reason)
	}

	return streams, nil
}
