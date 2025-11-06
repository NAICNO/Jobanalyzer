package special

import (
	. "sonalyze/common"
	"sonalyze/db/repr"
)

// The ClusterMeta is largely a per-command object representing an implementation-near view on the
// cluster for that command.  It needs to be fairly cheap and it will tend to wrap other data (eg,
// the methods of implementing objects may call into the database layer, and may reference
// more-persistent cluster and configuration data).  It is part of a manually managed "cluster
// table" that does not quite exist yet, but also an API to dynamic cluster configuration data.
// It's a bit of a hodgepodge and will evolve over time.
type ClusterMeta interface {
	// Return the name of the cluster.  This is assumed to be time-unvarying.
	ClusterName() string

	// The set of excluded users is time-dependent but we're going to ignore that for now.
	ExcludedUsers() []string

	// A fresh list of nodes present in a static config if we have a static config, otherwise an
	// empty slice.
	//
	// NOTE: This API will likely go away; in the future, configs will not provide node data.
	NodesDefinedInConfigIfAny() []*repr.NodeSummary

	// This can be nil.  We want the latest host information at or before the given time, which is
	// seconds since Unix epoch.  If the database has to be queried, the query window into the past
	// may be limited to 14 days.  The result is not necessarily stable, it may change if new data
	// come in, but will never revert to older data.  New data that replace a prior non-nil result
	// may or may not be honored in a timely manner.  A static cluster configuration, should it
	// exist, will be consulted only if the information can't be found in the database.
	LookupHostByTime(host Ustr, time int64) *repr.NodeSummary

	// Return a list of logfiles iff we have them, otherwise nil
	LogFiles() []string

	// Return a data directory either from -data-dir or computed from -jobanalyzer-dir, otherwise ""
	DataDir() string

	// Return a data directory either from -report-dir or computed from -jobanalyzer-dir, otherwise ""
	ReportDir() string
}
