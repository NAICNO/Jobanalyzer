// The "cluster" table
//
// This is a weird one, for historical reasons - the cluster table is not terribly well-defined but
// is an emergent property of a bunch of data stored in various places.  Some of the weirdness will
// go away, by and by.
//
// This command will do these things:
//
// - enumerate the subdirectories in $jobanalyzer_dir/data and take these to be canonical
//   cluster names
// - look for $jobanalyzer_dir/cluster-aliases.json and take the values therein that have
//   mappings to names found in step 1 to be valid aliases
// - look for $jobanalyzer_dir/scripts/<cluster>/<cluster>-config.json and take the values
//   therein for the cluster and associate them with the values computed in steps 1 and 2.
//
// Then the cluster names, aliases and other data will be printed, current keys are `cluster`,
// `aliases`, and `description`.  The alias list is always a string of comma-separated alias names
// (even for JSON format).  Some clusters may have no aliases or description.
//
// Note that any aliases in the `<cluster>-config.json` files are ignored for the time being, they
// are also unused by all other code.
//
// The cluster table is cached; only explicit invalidation will flush it.

package db

import (
	"errors"
	"io/fs"
	"path"
	"os"
	"sync"

	"go-utils/alias"
)

const (
	dataDirName = "data"
	clusterConfigDirName = "scripts"
	clusterAliasesFilename = "cluster-aliases.json"
)

// Name of the cluster's config file
func MakeConfigFilePath(jobanalyzerDir, clusterName string) string {
	return path.Join(
		jobanalyzerDir,
		clusterConfigDirName,
		clusterName,
		clusterName+"-config.json",
	)
}

// Name of the cluster's data directory
func MakeClusterDataPath(jobanalyzerDir, clusterName string) string {
	return path.Join(jobanalyzerDir, dataDirName, clusterName)
}

var (
	// MT: Locked + Contained objects are immutable or thread-safe after creation
	clusterCacheLock sync.Mutex

	// The cache has a value if the clusterCache map is not nil.  The clusterAliases may be nil if
	// there was no alias file.
	clusterCache     map[string]*ClusterEntry
	clusterAliases   *alias.Aliases
)

type ClusterEntry struct {
	Name        string
	Description string
	Aliases     []string
}

// The cluster table is returned as a pair: a shared immutable map from cluster name to cluster
// information and (for historical reasons) a thread-safe alias resolver object.

func ReadClusterData(
	jobanalyzerDir string,
) (clusters map[string]*ClusterEntry, aliases *alias.Aliases, err error) {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	if clusterCache != nil {
		clusters = clusterCache
		aliases = clusterAliases
		return
	}

	clusters = make(map[string]*ClusterEntry)

	// Find cluster names from the data directory
	dirEntries, err := os.ReadDir(path.Join(jobanalyzerDir, dataDirName))
	if err != nil {
		return
	}
	for _, e := range dirEntries {
		if e.IsDir() {
			n := e.Name()
			clusters[n] = &ClusterEntry{Name: n}
		}
	}

	// Add aliases to known clusters.  The aliases file is optional, but if something with that name
	// is there it is an error to fail to open it.
	aliasesFile := path.Join(jobanalyzerDir, clusterAliasesFilename)
	if info, bad := os.Stat(aliasesFile); bad == nil {
		if info.Mode()&fs.ModeType != 0 {
			err = errors.New("Cluster alias file is not a regular file")
			return
		}
		aliases, err = alias.ReadAliases(aliasesFile)
		if err != nil {
			return
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
			continue
		}
		v.Description = cfg.Description
	}

	clusterCache = clusters
	clusterAliases = aliases
	return
}

func InvalidateClusterCache() {
	clusterCacheLock.Lock()
	defer clusterCacheLock.Unlock()

	clusterCache = nil
	clusterAliases = nil
}
