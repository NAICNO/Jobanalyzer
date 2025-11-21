// Global database of ClusterEntry objects (internal representation of database connections).

package special

import (
	"sync"

	"go-utils/alias"
	"go-utils/config"
	umaps "go-utils/maps"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

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
	repr.Cluster

	// Semi-private implementation bits, should be initialized by setup code in sonalyze/db and
	// accessed only by dbContext methods, also in sonalyze/db.
	HaveDatabase       bool
	DatabaseConnection any
	HaveDataDir        bool
	DataDir            string
	HaveLogFiles       bool
	LogFiles           []string
	LogFileType        types.DataType
	HaveReportDir      bool
	ReportDir          string
	HaveConfig         bool
	Config             *config.ClusterConfig
}

func DefineClusters(clusters map[string]*ClusterEntry, aliases *alias.Aliases) {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if dataStoreOpen {
		panic("Data store is already open")
	}

	dataStoreOpen = true
	clusterCache = clusters
	clusterAliases = aliases
}

func NewClusterEntry() *ClusterEntry {
	// This will become more elaborate
	return new(ClusterEntry)
}

func ClearClusters() {
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
