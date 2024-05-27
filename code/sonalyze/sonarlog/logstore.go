// LogStore - API to the underlying data logs
//
// The logstore is a global singleton sonarlog.TheStore that manages reading, appending, and
// (eventually) transparent caching.  Clients open individual data (cluster) directories by calling
// OpenDir() or sets of read-only input files by calling OpenFiles().  The returned LogDir structure
// will then manage file I/O within that space.
//
// The main thread would normally `defer sonarlog.TheStore.Close()` to make sure that all pending
// writes are done when the program shuts down.

package sonarlog

import (
	"errors"
	"path"
	"sync"
)

// Currently just a container for individual directories, and a flag to prevent further operation.

type LogStore struct {
	sync.Mutex
	closed bool
	dirs map[string]*LogDir
}

// Global singleton.

var TheStore = LogStore {
	dirs: make(map[string]*LogDir, 10),
}

// Open a directory and attach it to the global logstore.
//
// `dir` is the root directory of the log data store for a cluster.  It contains subdirectory paths
// of the form YYYY/MM/DD for data.  At the leaf of each path are read-only data files for the given
// date:
//
//  - HOSTNAME.csv contain Sonar `ps` log data for the given host
//  - sysinfo-HOSTNAME.json contain Sonar `sysinfo` system data for the given host

func OpenDir(dir string) (*LogDir, error) {
	return TheStore.openDir(dir)
}

// Open a set of read-only files, that are not attached to the global logstore.  `files` is a
// nonempty list of files containing Sonar `ps` log data.  We make a private copy of the list.

func OpenFiles(files []string) (*LogDir, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return newFiles(files), nil
}

// Drain all pending writes, kill all the LogDir nodes, and return when it's all done.  The store is marked
// as closed and no operations to open new directories will work.

func (ls *LogStore) Close() {
	ls.Lock()
	defer ls.Unlock()

	if ls.closed {
		return
	}

	// Close everything synchronously, normally this is fine.  We could do them in parallel but
	// there isn't any reason to do that yet.

	ls.closed = true
	dirs := ls.dirs
	ls.dirs = nil
	for _, d := range dirs {
		d.closeOrFlush(true)
	}
}

func (ls *LogStore) openDir(dir string) (*LogDir, error) {
	ls.Lock()
	defer ls.Unlock()

	if ls.closed {
		return nil, LogClosedErr
	}

	dir = path.Clean(dir)
	if d, ok := ls.dirs[dir]; ok {
		return d, nil
	}

	d := newDir(dir)
	ls.dirs[dir] = d
	return d, nil
}
