// The "cluster" table
//
// This is a weird one, for historical reasons - the cluster table is not terribly well-defined but
// is an emergent property of a bunch of data stored in various places.  Some of the weirdness will
// go away, by and by.
//
// This command will do these things:
//
// - enumerate the subdirectories in $jobanalyzer_dir/data and take these to be canonical cluster
//   names
// - look for $jobanalyzer_dir/cluster-config/cluster-aliases.json and take the values therein that
//   have mappings to names found in step 1 to be valid aliases
// - look for $jobanalyzer_dir/cluster-config/<cluster>-config.json and take the values therein for
//   the cluster and associate them with the values computed in steps 1 and 2.
//
// Note that any aliases in the `<cluster>-config.json` files are ignored for the time being, they
// are also unused by all other code.
//
// The cluster table is cached; only explicit invalidation will flush it.  The cluster data must be
// treated as completely read-only.

package special

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sync"

	"go-utils/alias"
	"go-utils/config"
	umaps "go-utils/maps"
)

const (
	dataDirName            = "data"
	reportDirName          = "reports"
	clusterConfigDirName   = "cluster-config"
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

var (
	// MT: Locked + Contained objects are immutable or thread-safe after creation
	clusterCacheLock sync.Mutex

	// The cache has a value if the clusterCache map is not nil.  The clusterAliases may be nil if
	// there was no alias file.
	dataStoreOpen  bool
	clusterCache   map[string]*ClusterEntry
	clusterAliases *alias.Aliases
)

type ClusterEntry struct {
	// These fields must never be modified.
	Name        string
	Description string
	Aliases     []string // Not sorted
	ExcludeUser []string // Not sorted

	// Misc implementation - semi-private, shared with ClusterMeta for now.
	HaveDataDir   bool
	DataDir       string
	HaveLogFiles  bool
	LogFiles      []string
	HaveReportDir bool
	ReportDir     string
	HaveConfig    bool
	Config        *config.ClusterConfig
}

func newClusterEntry() *ClusterEntry {
	// This will become more elaborate
	return new(ClusterEntry)
}

func OpenFullDataStore(jobanalyzerDir string) error {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	var (
		clusters map[string]*ClusterEntry
		aliases  *alias.Aliases
	)

	clusters = make(map[string]*ClusterEntry)

	// Find cluster names from the data directory
	dirEntries, err := os.ReadDir(path.Join(jobanalyzerDir, dataDirName))
	if err != nil {
		return err
	}
	for _, e := range dirEntries {
		if e.IsDir() {
			c := newClusterEntry()
			c.Name = e.Name()
			c.HaveDataDir = true
			c.DataDir = MakeClusterDataPath(jobanalyzerDir, c.Name)
			c.HaveReportDir = true
			c.ReportDir = MakeReportDirPath(jobanalyzerDir, c.Name)
			clusters[c.Name] = c
		}
	}

	// Add aliases to known clusters.  The aliases file is optional, but if something with that name
	// is there it is an error to fail to open it.
	aliasesFile := path.Join(jobanalyzerDir, clusterConfigDirName, clusterAliasesFilename)
	if info, bad := os.Stat(aliasesFile); bad == nil {
		if info.Mode()&fs.ModeType != 0 {
			return errors.New("Cluster alias file is not a regular file")
		}
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
		cfg, err := ReadConfigData(MakeConfigFilePath(jobanalyzerDir, c))
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

	dataStoreOpen = true
	clusterCache = clusters
	clusterAliases = aliases
	return nil
}

// For the following, the cluster name will be set to "data.cluster", "report.cluster",
// "logfiles.cluster", "config.cluster" if configFile is not provided or if the file does not
// provide a cluster name.

func OpenDataStoreFromDataDir(dataDir, configFile string) error {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	cfg, err := maybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := newClusterEntry()
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

	dataStoreOpen = true
	clusterCache = map[string]*ClusterEntry{v.Name: v}
	return nil
}

func OpenDataStoreFromReportDir(reportDir, configFile string) error {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	cfg, err := maybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := newClusterEntry()
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

	dataStoreOpen = true
	clusterCache = map[string]*ClusterEntry{v.Name: v}
	return nil
}

func OpenDataStoreFromLogFiles(logFiles []string, configFile string) error {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	cfg, err := maybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := newClusterEntry()
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

	dataStoreOpen = true
	clusterCache = map[string]*ClusterEntry{v.Name: v}
	return nil
}

func OpenDataStoreFromConfigFile(configFile string) error {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	cfg, err := maybeGetConfig(configFile)
	if err != nil {
		return fmt.Errorf("Could not read config file %s: %v", configFile, err)
	}
	v := newClusterEntry()
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

	dataStoreOpen = true
	clusterCache = map[string]*ClusterEntry{v.Name: v}
	return nil
}

func OpenDataStoreFromConfig(cfg *config.ClusterConfig) error {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	v := newClusterEntry()
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

	dataStoreOpen = true
	clusterCache = map[string]*ClusterEntry{v.Name: v}
	return nil
}

func CloseDataStore() {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if !dataStoreOpen {
		panic("Data store is not open")
	}

	dataStoreOpen = false
	clusterCache = nil
	clusterAliases = nil
}

func GetSingleCluster() *ClusterEntry {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if !dataStoreOpen {
		panic("Data store is not open")
	}

	if len(clusterCache) == 1 {
		for _, v := range clusterCache {
			return v
		}
	}
	return nil
}

func LookupCluster(name string) *ClusterEntry {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if !dataStoreOpen {
		panic("Data store is not open")
	}

	if clusterAliases != nil {
		name = clusterAliases.Resolve(name)
	}
	return clusterCache[name]
}

func ResolveClusterName(name string) string {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if !dataStoreOpen {
		panic("Data store is not open")
	}

	if clusterAliases != nil {
		name = clusterAliases.Resolve(name)
	}
	return name
}

// Return a fresh slice of immutable cluster values.
func AllClusters() []*ClusterEntry {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if !dataStoreOpen {
		panic("Data store is not open")
	}

	return umaps.Values(clusterCache)
}

// Read the config file if the file name is not empty.
func maybeGetConfig(configFileName string) (*config.ClusterConfig, error) {
	if configFileName == "" {
		return nil, nil
	}
	return ReadConfigData(configFileName)
}
