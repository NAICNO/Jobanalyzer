package cluster

import (
	"slices"
	_ "time"

	. "sonalyze/common"
	_ "sonalyze/data/config"
	_ "sonalyze/db"
	"sonalyze/db/repr"
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

func (tm *clusterMeta) LookupHostByTime(host Ustr, t int64) *repr.NodeSummary {
	if tm.cluster.HaveConfig {
		return tm.cluster.Config.LookupHost(host.String())
	}
	return nil
}

func (tm *clusterMeta) NodesDefinedInConfigIfAny() []*repr.NodeSummary {
	if tm.cluster.HaveConfig {
		return slices.Clone(tm.cluster.Config.Hosts())
	}
	return make([]*repr.NodeSummary, 0)
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
