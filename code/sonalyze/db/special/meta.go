package special

import (
	"sonalyze/db/repr"
)

// ClusterMeta is an atrophied shim around a *ClusterEntry.  It no longer has any utility (except
// marginally for testing since we can mock a ClusterEntry).  It should be removed and the methods
// of the main implementor, clusterMeta, should be moved onto ClusterEntry.  Probably ClusterEntry
// should not be called that.

type ClusterMeta interface {
	// The underlying cluster object.
	Cluster() *ClusterEntry

	// Return the name of the cluster.  This is assumed to be time-unvarying.
	ClusterName() string

	// The set of excluded users is time-dependent but we're going to ignore that for now.
	ExcludedUsers() []string

	// A fresh list of nodes present in a static config if we have a static config, otherwise an
	// empty slice.
	//
	// NOTE: This API will likely go away; in the future, configs will not provide node data.
	NodesDefinedInConfigIfAny() []*repr.NodeSummary

	// Return true if we have log files of the given type, or of any type if the type is not
	// yet set.
	HaveLogFilesOfType(dataType DataType) bool

	// Return a list of logfiles iff we have them and they are of the given type, otherwise nil.  If
	// no type has been set, we freeze the type with this type (actually the set of types that
	// incorporate the type).
	LogFiles(dataType DataType) []string

	// Return a data directory either from -data-dir or computed from -jobanalyzer-dir, otherwise ""
	DataDir() string

	// Return a data directory either from -report-dir or computed from -jobanalyzer-dir, otherwise ""
	ReportDir() string
}
