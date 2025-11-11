package cluster

import (
	"slices"

	"sonalyze/db/repr"
	"sonalyze/db/special"
)

type clusterMeta struct {
	cluster        *special.ClusterEntry
}

func NewMetaFromCluster(cluster *special.ClusterEntry) special.ClusterMeta {
	return &clusterMeta { cluster }
}

func (tm *clusterMeta) Cluster() *special.ClusterEntry {
	return tm.cluster
}

func (tm *clusterMeta) ClusterName() string {
	return tm.cluster.Name
}

func (tm *clusterMeta) ExcludedUsers() []string {
	return tm.cluster.ExcludeUser
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

func (tm *clusterMeta) HaveLogFilesOfType(dataType special.DataType) bool {
	return tm.cluster.HaveLogFiles && (tm.cluster.LogFileType == 0 || (dataType & tm.cluster.LogFileType) != 0)
}

func (tm *clusterMeta) LogFiles(dataType special.DataType) []string {
	if tm.cluster.HaveLogFiles {
		if dataType == 0 {
			panic("Zero data type")
		}
		if tm.cluster.LogFileType == 0 {
			tm.cluster.LogFileType = dataType
		}
		if tm.cluster.LogFileType == dataType {
			return tm.cluster.LogFiles
		}
	}
	return nil
}

func (tm *clusterMeta) ReportDir() string {
	if tm.cluster.HaveReportDir {
		return tm.cluster.ReportDir
	}
	return ""
}
