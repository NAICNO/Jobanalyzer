// This is a mess.  It's now mostly cluster cache plus ClusterEntry.  How much can we move one level up?
//
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
	"path"
	"sync"

	"go-utils/alias"
	"go-utils/config"
	umaps "go-utils/maps"
	"sonalyze/db/repr"
)

const (
	dataDirName            = "data"
	reportDirName          = "reports"
	clusterConfigDirName   = "cluster-config"
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

func InitializeDataStore(clusters map[string]*ClusterEntry, aliases *alias.Aliases) {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	dataStoreOpen = true
	clusterCache = clusters
	clusterAliases = aliases
}

type ClusterEntry struct {
	repr.Cluster

	// Misc implementation - semi-private, shared with ClusterMeta for now.
	HaveDatabase  bool
	DatabaseConnection any
	HaveDataDir   bool
	DataDir       string
	HaveLogFiles  bool
	LogFiles      []string
	LogFileType   DataType
	HaveReportDir bool
	ReportDir     string
	HaveConfig    bool
	Config        *config.ClusterConfig
}

func NewClusterEntry() *ClusterEntry {
	// This will become more elaborate
	return new(ClusterEntry)
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
func MaybeGetConfig(configFileName string) (*config.ClusterConfig, error) {
	if configFileName == "" {
		return nil, nil
	}
	return ReadConfigData(configFileName)
}
