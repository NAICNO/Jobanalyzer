// Experiment

package special

import (
	"go-utils/config"
)

// The ClusterMeta is largely a per-command object representing a view on the cluster for that
// command.  It needs to be fairly cheap and it will tend to wrap other data (eg, the methods of
// implementing objects may call into the database layer, and may reference more-persistent cluster
// and configuration data).  It's a bit of a hodgepodge.
type ClusterMeta interface {
	// Return the name of the cluster.  This is assumed to be time-unvarying.
	ClusterName() string

	// This is also time-dependent in the same way as the following, but let's not worry about it
	// yet.
	ExcludedUsers() []string

	// Host names defined in the time window.
	HostsDefinedInTimeWindow(fromIncl, toIncl int64) []string

	// Nodes present in the time window.
	NodesDefinedInTimeWindow(fromIncl, toIncl int64) []*config.NodeConfigRecord

	// This can be nil.  We want the latest host information at or before the given time, which is
	// seconds since Unix epoch.
	LookupHostByTime(host string, time int64) *config.NodeConfigRecord

	// Data for the underlying representation, used by various database implementations.

	// Return a list of logfiles iff we have them, otherwise nil
	LogFiles() []string

	// Return a data directory either from -data-dir or computed from -jobanalyzer-dir, otherwise ""
	DataDir() string

	// Return a data directory either from -report-dir or computed from -jobanalyzer-dir, otherwise ""
	ReportDir() string
}
