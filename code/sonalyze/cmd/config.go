package cmd

import (
	"fmt"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/data/sample"
)

// Standard method for cleaning an InputStreamSet relative to a config: the config must have a
// definition for each host.  No config at all is a fatal error.  No config for a host means we
// remove the host from the set (imperatively).
//
// Over time this may become more complicated, as the config becomes time-dependent.
func EnsureConfigForInputStreams(
	cfg *config.ClusterConfig,
	streams sample.InputStreamSet,
	reason string,
) error {
	// Bail if there's no config data at all.
	if cfg == nil {
		return fmt.Errorf("Configuration file required: %s", reason)
	}

	// Remove streams for which we have no config data.
	bad := make(map[sample.InputStreamKey]bool)
	for key, stream := range streams {
		hn := (*stream)[0].Hostname.String()
		if cfg.LookupHost(hn) == nil {
			bad[key] = true
			Log.Infof("Warning: Missing host configuration for %s", hn)
		}
	}

	for b := range bad {
		delete(streams, b)
	}

	return nil
}
