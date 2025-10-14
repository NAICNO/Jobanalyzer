// Experiment

package special

import (
	"go-utils/config"
)

type ClusterMeta interface {
	ClusterName() string

	// Arguably this is also time-dependent but let's not worry about it yet
	ExcludedUsers() []string

	// Host names defined in the time window
	HostsDefinedInTimeWindow(fromIncl, toIncl int64) []string

	// This can be nil.  We want the latest host information at or before the given time, which is
	// seconds since Unix epoch
	LookupHostByTime(host string, time int64) *config.NodeConfigRecord

	// Return the cluster configuration at the given time.
	ConfigAtTime(time int64) *config.ClusterConfig

	// Data for the underlying representation.

	// Return a list of logfiles iff we have them, otherwise nil
	LogFiles() []string

	// Return a data directory either from -data-dir or computed from -jobanalyzer-dir, otherwise ""
	DataDir() string

	// Return a data directory either from -report-dir or computed from -jobanalyzer-dir, otherwise ""
	ReportDir() string
}
