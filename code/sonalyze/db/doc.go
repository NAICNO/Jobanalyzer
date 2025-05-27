// Jobanalyzer has a simple data base interface that provides access to individual data streams,
// basically a trivial time series database at the moment.  The data base can be backed by either a
// list of individual immutable files or a (mutable) directory tree, as described in filedb/.  In
// the future, proper time series databases will be added.
//
// Lists of files can be opened with the functions in fileliststore.go, directory trees with the
// functions in dirtreestore.go.  For traditional reasons these are referred to throughout the code
// as "transient clusters" and "persistent clusters", respectively.
//
// The main thread would normally `defer db.Close()` to make sure that all pending writes are done
// when the program shuts down, if read-write databases are opened.
//
// Locking (internals).
//
// We have these major pieces:
//
//   - ClusterStore, a global singleton
//   - PersistentCluster, a per-cluster unique object for directory-backed cluster data
//   - TransientSampleCluster, a per-file-set object non-unique object for a set of read-only files
//   - LogFile, a per-file (read-write or read-only) object, unique within directory-backed data
//   - Purgable data, a global singleton for the cache
//
// There are multiple goroutines that handle file I/O, and hence there is concurrent access to all
// the pieces.  To handle this we mostly use mutexes.
//
// Locks are in a strict hierarchy (a DAG) as follows.
//
//   - ClusterStore has a lock that covers its internal data.  This lock may be held while methods are
//     being called on individual cluster objects.
//
//   - Each PersistentCluster and (in principle, though not at present since it's not needed)
//     TransientCluster has a lock that covers its internal data.  These locks may be held while
//     methods are being called on individual files.
//
//   - Each LogFile has a lock.  The main lock covers the file's main mutable data structures.  This
//     lock may be held when the file code calls into the cache code, and usually it must be, so that
//     the cache code can know the state of the file.
//
//   - The cache code has a lock on a global singleton data structure, the purgeable set.  This lock
//     also covers the part of the purgeable set data structure that lives in each LogFile.
//
// The cache code can call back up to the LogFile code (notably to purge a file).  In this case it
// must hold *no* nocks at all, as the LogFile code may call down into the cache code again.  Thus
// the cache code can't know for a fact that the file's state hasn't changed between the time it
// selects it for purging and the time it purges it.  It must be resilient to that.
package db
