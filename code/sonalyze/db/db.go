package db

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"go-utils/alias"
	"go-utils/config"
	"sonalyze/db/special"
	"sonalyze/db/types"
)

// Utility to open a database from a set of parameters.  We must have dataDir xor logFiles.
func OpenReadOnlyDB(meta types.Context, dataType types.DataType) (DataProvider, error) {
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

const (
	dataDirName            = "data"
	clusterConfigDirName   = "cluster-config" // FIXME: duplicated, should abstract it
	clusterAliasesFilename = "cluster-aliases.json"
)

func OpenFullDataStore(jobanalyzerDir, databaseURI string) error {
	var (
		clusters map[string]*special.ClusterEntry
		aliases  *alias.Aliases
	)

	clusters = make(map[string]*special.ClusterEntry)

	if databaseURI != "" {
		theDB, err := OpenDatabaseURI(databaseURI)
		if err != nil {
			return err
		}
		clusterNames, err := theDB.EnumerateClusters()
		if err != nil {
			return err
		}
		for _, name := range clusterNames {
			c := special.NewClusterEntry()
			c.Name = name
			c.HaveDatabase = true
			c.DatabaseConnection = theDB
			clusters[c.Name] = c
		}
	} else {
		// Find cluster names from the data directory
		dirEntries, err := os.ReadDir(path.Join(jobanalyzerDir, dataDirName))
		if err != nil {
			return err
		}
		for _, e := range dirEntries {
			if e.IsDir() {
				c := special.NewClusterEntry()
				c.Name = e.Name()
				c.HaveDataDir = true
				c.DataDir = special.MakeClusterDataPath(jobanalyzerDir, c.Name)
				c.HaveReportDir = true
				c.ReportDir = special.MakeReportDirPath(jobanalyzerDir, c.Name)
				clusters[c.Name] = c
			}
		}
	}

	// Add aliases to known clusters.  The aliases file is optional, but if something with that name
	// is there it is an error to fail to open it.
	aliasesFile := path.Join(jobanalyzerDir, clusterConfigDirName, clusterAliasesFilename)
	if info, bad := os.Stat(aliasesFile); bad == nil {
		if info.Mode()&fs.ModeType != 0 {
			return errors.New("Cluster alias file is not a regular file")
		}
		var err error
		aliases, err = alias.ReadAliases(aliasesFile)
		if err != nil {
			return err
		}
	}
	if aliases != nil {
		for c, as := range aliases.ReverseExpand() {
			if probe, found := clusters[c]; found {
				probe.Aliases = as
			}
		}
	}

	// Find descriptions for known clusters.
	for c, v := range clusters {
		cfg, err := special.ReadConfigData(special.MakeConfigFilePath(jobanalyzerDir, c))
		if err != nil {
			// Arguably we could remove it, but this code will change anyway.
			v.Description = "No configuration found"
			continue
		}
		v.Description = cfg.Description
		v.ExcludeUser = cfg.ExcludeUser
		v.HaveConfig = true
		v.Config = cfg
	}

	special.InitializeDataStore(clusters, aliases)
	return nil
}

// For the following, the cluster name will be set to "data.cluster", "report.cluster",
// "logfiles.cluster", "config.cluster" if configFile is not provided or if the file does not
// provide a cluster name.

func OpenDataStoreFromDataDir(dataDir, configFile string) error {
	cfg, err := special.MaybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := special.NewClusterEntry()
	v.HaveDataDir = true
	v.DataDir = dataDir
	if cfg != nil {
		v.Name = cfg.Name
		v.Description = cfg.Description
		v.HaveConfig = true
		v.Config = cfg
	}
	if v.Name == "" {
		v.Name = "data.cluster"
	}
	if v.Description == "" {
		v.Description = "anonymous cluster (data dir)"
	}
	clusters := map[string]*special.ClusterEntry{v.Name: v}

	special.InitializeDataStore(clusters, nil)
	return nil
}

func OpenDataStoreFromReportDir(reportDir, configFile string) error {
	cfg, err := special.MaybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := special.NewClusterEntry()
	v.HaveReportDir = true
	v.ReportDir = reportDir
	if cfg != nil {
		v.Name = cfg.Name
		v.Description = cfg.Description
		v.HaveConfig = true
		v.Config = cfg
	}
	if v.Name == "" {
		v.Name = "report.cluster"
	}
	if v.Description == "" {
		v.Description = "anonymous cluster (report dir)"
	}
	clusters := map[string]*special.ClusterEntry{v.Name: v}

	special.InitializeDataStore(clusters, nil)
	return nil
}

func OpenDataStoreFromLogFiles(logFiles []string, configFile string) error {
	cfg, err := special.MaybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := special.NewClusterEntry()
	v.HaveLogFiles = true
	v.LogFiles = logFiles
	if cfg != nil {
		v.Name = cfg.Name
		v.Description = cfg.Description
		v.HaveConfig = true
		v.Config = cfg
	}
	if v.Name == "" {
		v.Name = "logfiles.cluster"
	}
	if v.Description == "" {
		v.Description = "anonymous cluster (log files)"
	}
	clusters := map[string]*special.ClusterEntry{v.Name: v}

	special.InitializeDataStore(clusters, nil)
	return nil
}

func OpenDataStoreFromConfigFile(configFile string) error {
	cfg, err := special.MaybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := special.NewClusterEntry()
	v.Name = cfg.Name
	v.Description = cfg.Description
	v.HaveConfig = true
	v.Config = cfg
	if v.Name == "" {
		v.Name = "config.cluster"
	}
	if v.Description == "" {
		v.Description = "anonymous cluster (config file)"
	}
	clusters := map[string]*special.ClusterEntry{v.Name: v}

	special.InitializeDataStore(clusters, nil)
	return nil
}

func OpenDataStoreFromConfig(cfg *config.ClusterConfig) error {
	v := special.NewClusterEntry()
	v.Name = cfg.Name
	v.Description = cfg.Description
	v.HaveConfig = true
	v.Config = cfg
	if v.Name == "" {
		v.Name = "config.cluster"
	}
	if v.Description == "" {
		v.Description = "anonymous cluster (config file)"
	}
	clusters := map[string]*special.ClusterEntry{v.Name: v}

	special.InitializeDataStore(clusters, nil)
	return nil
}
