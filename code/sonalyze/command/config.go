package command

import (
	"fmt"
	"log"

	"go-utils/config"
	"sonalyze/sonarlog"
)

// Standard method for cleaning an InputStreamSet relative to a config: the config must have a
// definition for each host.  No config at all is a fatal error.  No config for a host means we
// remove the host from the set, we return the modified set.
//
// Over time this may become more complicated, as the config becomes time-dependent.
func EnsureConfigForInputStreams(
	cfg *config.ClusterConfig,
	streams sonarlog.InputStreamSet,
	reason string,
) (sonarlog.InputStreamSet, error) {
	// Bail if there's no config data at all.
	if cfg == nil {
		return nil, fmt.Errorf("Configuration file required: %s", reason)
	}

	// Remove streams for which we have no config data.
	bad := make(map[sonarlog.InputStreamKey]bool)
	for key, stream := range streams {
		hn := (*stream)[0].Host.String()
		if cfg.LookupHost(hn) == nil {
			bad[key] = true
			log.Printf("Warning: Missing host configuration for %s", hn)
		}
	}

	for b := range bad {
		delete(streams, b)
	}

	return streams, nil
}
