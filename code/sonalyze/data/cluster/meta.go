package cluster

import (
	"go-utils/config"
	"sonalyze/db/special"
)

type trivialMeta struct {
	cluster        *special.ClusterEntry
}

func NewMetaFromCluster(cluster *special.ClusterEntry) special.ClusterMeta {
	return &trivialMeta { cluster }
}

func (tm *trivialMeta) ClusterName() string {
	return tm.cluster.Name
}

func (tm *trivialMeta) ExcludedUsers() []string {
	return tm.cluster.ExcludeUser
}

func (tm *trivialMeta) LookupHostByTime(host string, time int64) *config.NodeConfigRecord {
	if tm.cluster.HaveConfig {
		return tm.cluster.Config.LookupHost(host)
	}
	return nil
}

func (tm *trivialMeta) HostsDefinedInTimeWindow(fromIncl, toIncl int64) []string {
	if tm.cluster.HaveConfig {
		return tm.cluster.Config.HostsDefinedInTimeWindow(fromIncl, toIncl)
	}
	return nil
}

// These go away again eventually

func (tm *trivialMeta) DataDir() string {
	if tm.cluster.HaveDataDir {
		return tm.cluster.DataDir
	}
	return ""
}

func (tm *trivialMeta) LogFiles() []string {
	if tm.cluster.HaveLogFiles {
		return tm.cluster.LogFiles
	}
	return nil
}

func (tm *trivialMeta) ReportDir() string {
	if tm.cluster.HaveReportDir {
		return tm.cluster.ReportDir
	}
	return ""
}

func (tm *trivialMeta) ConfigFile() *config.ClusterConfig {
	if tm.cluster.HaveConfig {
		return tm.cluster.Config
	}
	return nil
}
