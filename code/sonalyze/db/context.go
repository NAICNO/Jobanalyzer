// Default implementation of sonalyze/db/types.Context

package db

import (
	"slices"

	"go-utils/config"
	"sonalyze/db/repr"
	"sonalyze/db/special"
	"sonalyze/db/types"
)

type dbContext struct {
	cluster *special.ClusterEntry
}

func NewContextFromCluster(cluster *special.ClusterEntry) types.Context {
	return &dbContext{cluster}
}

func (tm *dbContext) ClusterName() string {
	return tm.cluster.Name
}

func (tm *dbContext) ExcludedUsers() []string {
	return tm.cluster.ExcludeUser
}

func (tm *dbContext) NodesDefinedInConfigIfAny() []*repr.NodeSummary {
	if tm.cluster.HaveConfig {
		return slices.Clone(tm.cluster.Config.Hosts())
	}
	return make([]*repr.NodeSummary, 0)
}

func (tm *dbContext) DataDir() string {
	if tm.cluster.HaveDataDir {
		return tm.cluster.DataDir
	}
	return ""
}

func (tm *dbContext) HaveLogFilesOfType(dataType types.DataType) bool {
	return tm.cluster.HaveLogFiles && (tm.cluster.LogFileType == 0 || (dataType&tm.cluster.LogFileType) != 0)
}

func (tm *dbContext) LogFiles(dataType types.DataType) []string {
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

func (tm *dbContext) HaveDatabaseConnection() bool {
	return tm.cluster.HaveDatabase
}

func (tm *dbContext) ConnectedDB() any {
	return tm.cluster.DatabaseConnection
}

func (tm *dbContext) ReportDir() string {
	if tm.cluster.HaveReportDir {
		return tm.cluster.ReportDir
	}
	return ""
}

func (tm *dbContext) HaveConfig() bool {
	return tm.cluster.HaveConfig
}

func (tm *dbContext) Config() *config.ClusterConfig {
	return tm.cluster.Config
}
