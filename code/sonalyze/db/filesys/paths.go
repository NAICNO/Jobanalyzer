package filesys

import (
	"path"
)

const (
	dataDirName          = "data"
	reportDirName        = "reports"
	clusterConfigDirName = "cluster-config"
	clusterAliasesFilename = "cluster-aliases.json"
)

// Name of the cluster's config file
func MakeConfigFilePath(jobanalyzerDir, clusterName string) string {
	return path.Join(
		jobanalyzerDir,
		clusterConfigDirName,
		clusterName+"-config.json",
	)
}

// Name of the cluster's data directory
func MakeClusterDataPath(jobanalyzerDir, clusterName string) string {
	return path.Join(jobanalyzerDir, dataDirName, clusterName)
}

// Name of the cluster's reports directory
func MakeReportDirPath(jobanalyzerDir, clusterName string) string {
	return path.Join(jobanalyzerDir, reportDirName, clusterName)
}

func MakeClusterDataDirPath(jobanalyzerDir string) string {
	return path.Join(jobanalyzerDir, dataDirName)
}

func MakeClusterAliasesPath(jobanalyzerDir string) string {
	return path.Join(jobanalyzerDir, clusterConfigDirName, clusterAliasesFilename)
}

