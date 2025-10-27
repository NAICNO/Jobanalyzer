// Interface to a database based on directory trees.  See doc.go in this directory and in filedb/
// for more information.

package db

import (
	"path"
	"sync"

	"sonalyze/db/errs"
	"sonalyze/db/filedb"
	"sonalyze/db/special"
)

type clusterStore struct {
	sync.Mutex
	cacheSize   int64
	initialized bool
	closed      bool
	clusters    map[string]*filedb.PersistentCluster
}

// MT: Constant after initialization; thread-safe
var gClusterStore clusterStore

func unsafeResetClusterStore() {
	gClusterStore = clusterStore{
		clusters: make(map[string]*filedb.PersistentCluster, 10),
	}
}

// Note, this does not reinitialize I/O goroutines or their state

func init() {
	unsafeResetClusterStore()
}

// SetCacheSize can be called to size the memory cache for the database, in some
// implementation-defined way, before the first database operation is performed.
func SetCacheSize(size int64) {
	gClusterStore.setCacheSize(size)
}

// Open a date-keyed directory tree as a read-only persistent database.
func OpenPersistentDirectoryDB(
	meta special.ClusterMeta,
) (PersistentDataProvider, error) {
	return gClusterStore.openPersistentCluster(meta, meta.DataDir())
}

// Open a date-keyed directory tree as a read-write persistent database.
func OpenAppendablePersistentDirectoryDB(
	meta special.ClusterMeta,
) (AppendablePersistentDataProvider, error) {
	return gClusterStore.openPersistentCluster(meta, meta.DataDir())
}

// Drain all pending writes in the database, close all the attached Cluster nodes, and return when
// it's all done.  The store is marked as closed and no operations to open new directories will
// work, nor will operations on attached clusters work.
//
// Errors are not signalled because they are not generally useful at this point but there are error
// conditions, notably in the I/O.
func Close() {
	gClusterStore.close()
}

// For testing use.
func openPersistentCluster(meta special.ClusterMeta, dir string) (*filedb.PersistentCluster, error) {
	return gClusterStore.openPersistentCluster(meta, dir)
}

func (s *clusterStore) setCacheSize(size int64) {
	s.Lock()
	defer s.Unlock()
	if s.closed || s.initialized {
		return
	}
	if size > 0 {
		s.cacheSize = size
	}
}

func (s *clusterStore) lazyInitLocked() {
	if !s.initialized {
		s.initialized = true
		if s.cacheSize > 0 {
			filedb.CacheInit(s.cacheSize)
		}
	}
}

func (ls *clusterStore) openPersistentCluster(
	meta special.ClusterMeta,
	clusterDir string,
) (*filedb.PersistentCluster, error) {
	ls.Lock()
	defer ls.Unlock()
	if ls.closed {
		return nil, errs.ClusterClosedErr
	}
	ls.lazyInitLocked()

	// Normally the path will have been cleaned by command line parsing, but do it anyway.
	clusterDir = path.Clean(clusterDir)
	if d := ls.clusters[clusterDir]; d != nil {
		return d, nil
	}

	d := filedb.NewPersistentCluster(clusterDir, meta)
	ls.clusters[clusterDir] = d
	return d, nil
}

func (ls *clusterStore) close() {
	ls.Lock()
	defer ls.Unlock()
	if ls.closed {
		return
	}
	ls.lazyInitLocked()

	// Close everything synchronously, normally this is fine.  We could do them in parallel but
	// there isn't any reason to do that yet.  Drop errors on the floor, not much to be done about
	// them at this stage (except retry, maybe).

	ls.closed = true
	clusters := ls.clusters
	ls.clusters = nil
	for _, d := range clusters {
		d.Close()
	}
}
