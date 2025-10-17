package cluster

import (
	"slices"

	"go-utils/config"
	"sonalyze/db/special"
)

type clusterMeta struct {
	cluster        *special.ClusterEntry
}

func NewMetaFromCluster(cluster *special.ClusterEntry) special.ClusterMeta {
	return &clusterMeta { cluster }
}

func (tm *clusterMeta) ClusterName() string {
	return tm.cluster.Name
}

func (tm *clusterMeta) ExcludedUsers() []string {
	return tm.cluster.ExcludeUser
}

func (tm *clusterMeta) LookupHostByTime(host string, time int64) *config.NodeConfigRecord {

	// TODO
	//
	// Basically, HaveConfig and Config disappear and are replaced by lazily computed and cached
	// configurations for timespans.  If an explicit config file is provided and it has node data
	// then we always use that, otherwise we will go to the data store and hope for the best.  In
	// the data store, we get sysinfo as for node -newest and present that.  Then the object is
	// cached globally under the triple (clusterName, fromDate, toDate).  When we have only a
	// toDate, as in this case, we can use any cached value that has that toDate, suggesting we'll
	// want a side structure of some kind for that.
	//
	// Possibly we push the entire config management thing into the db layer since in principle
	// the DB can cache the configs for a date.
	//
	// It's LookupHostByTime that will get hit hard, being used by jobs and in postprocessing.  It needs
	// to be fast.
	//
	// So maybe the cache is only for (cluster x host x date) -> config and we do things to make sure
	// that is streamlined.  Contention could be an issue.

	if tm.cluster.HaveConfig {
		return tm.cluster.Config.LookupHost(host)
	}
	return nil
}

func (tm *clusterMeta) NodesDefinedInConfigIfAny() []*config.NodeConfigRecord {
	if tm.cluster.HaveConfig {
		return slices.Clone(tm.cluster.Config.Hosts())
	}
	return make([]*config.NodeConfigRecord, 0)
}

func (tm *clusterMeta) DataDir() string {
	if tm.cluster.HaveDataDir {
		return tm.cluster.DataDir
	}
	return ""
}

func (tm *clusterMeta) LogFiles() []string {
	if tm.cluster.HaveLogFiles {
		return tm.cluster.LogFiles
	}
	return nil
}

func (tm *clusterMeta) ReportDir() string {
	if tm.cluster.HaveReportDir {
		return tm.cluster.ReportDir
	}
	return ""
}
