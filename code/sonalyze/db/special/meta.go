// Experiment

package special

import (
	"go-utils/config"
)

type ClusterMeta interface {
	ClusterName() string

	// This is of dubious correctness but it's how we've had it traditionally.  Really this attribute is
	// connected to specific nodes.
	HasCrossNodeJobs() bool

	// Arguably this is also time-dependent but let's not worry about it yet
	ExcludedUsers() []string

	// Host names defined in the time window
	HostsDefinedInTimeWindow(fromIncl, toIncl int64) []string

	// This can be nil.  We want the latest host information at or before the given time, which is
	// seconds since Unix epoch
	LookupHostByTime(host string, time int64) *config.NodeConfigRecord

	// FIXME: the provider must be a db.DataProvider, circular packages, clean later
	SetDataProvider(provider any)
}
