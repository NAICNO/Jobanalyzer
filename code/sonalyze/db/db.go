package db

import (
	"fmt"

	"go-utils/config"
	"sonalyze/db/special"
)

// Utility to open a database from a set of parameters.  We must have dataDir xor logFiles.
func OpenReadOnlyDB(meta special.ClusterMeta, dataType special.DataType) (DataProvider, error) {
	var theLog DataProvider
	var err error
	if meta.HaveLogFilesOfType(dataType) {
		theLog, err = OpenFileListDB(meta, dataType)
	} else if meta.HasDatabaseConnection() {
		theLog, err = OpenConnectedDB(meta)
	} else {
		theLog, err = OpenPersistentDirectoryDB(meta)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to open log store: %v", err)
	}
	return theLog, nil
}

func OpenFullDataStore(jobanalyzerDir, databaseURI string) error {
	return special.OpenFullDataStore(jobanalyzerDir, databaseURI)
}

func OpenDataStoreFromDataDir(dataDir, configFile string) error {
	return special.OpenDataStoreFromDataDir(dataDir, configFile)
}

func OpenDataStoreFromReportDir(reportDir, configFile string) error {
	return special.OpenDataStoreFromReportDir(reportDir, configFile)
}

func OpenDataStoreFromLogFiles(logFiles []string, configFile string) error {
	return special.OpenDataStoreFromLogFiles(logFiles, configFile)
}

func OpenDataStoreFromConfigFile(configFile string) error {
	return special.OpenDataStoreFromConfigFile(configFile)
}

func OpenDataStoreFromConfig(cfg *config.ClusterConfig) error {
	return OpenDataStoreFromConfig(cfg)
}


