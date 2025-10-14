package cluster

import (
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
	if tm.cluster.HaveConfig {
		return tm.cluster.Config.LookupHost(host)
	}
	return nil
}

func (tm *clusterMeta) HostsDefinedInTimeWindow(fromIncl, toIncl int64) []string {
	if tm.cluster.HaveConfig {
		return tm.cluster.Config.HostsDefinedInTimeWindow(fromIncl, toIncl)
	}
	return nil
}

func (tm *clusterMeta) ConfigAtTime(_ int64) *config.ClusterConfig {
	if tm.cluster.HaveConfig {
		return tm.cluster.Config
	}
	return nil
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
