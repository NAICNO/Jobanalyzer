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
	"io/fs"
	"os"
	"path"
	"sync"

	"go-utils/alias"
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

// This data structure must be treated as completely read-only, including the Aliases slice.
type ClusterEntry struct {
	Name        string
	Description string
	Aliases     []string // Not sorted
	ExcludeUser []string // Not sorted
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
			n := e.Name()
			clusters[n] = &ClusterEntry{Name: n}
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
		cfg, err := MaybeGetConfig(MakeConfigFilePath(jobanalyzerDir, c))
		if err != nil {
			continue
		}
		v.Description = cfg.Description
		v.ExcludeUser = cfg.ExcludeUser
	}

	dataStoreOpen = true
	clusterCache = clusters
	clusterAliases = aliases
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

// The entry will be nil if the cluster is not defined.  The alias resolver is always consulted.
// Panics if the database is not open.
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
