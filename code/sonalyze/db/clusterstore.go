// ClusterStore - API to the underlying data logs
//
// The cluster store is a global singleton that manages reading, appending, and transparent caching.
// Clients open individual data (cluster) directories by calling OpenPersistentCluster() or sets of
// read-only sonar log input files by calling OpenTransientSampleCluster().  The returned Cluster
// structure will then manage file I/O within that space.
//
// The main thread would normally `defer sonarlog.Close()` to make sure that all pending writes are
// done when the program shuts down.
//
// Locking.
//
// We have these major pieces:
//
//  - ClusterStore, a global singleton
//  - PersistentCluster, a per-cluster unique object for directory-backed cluster data
//  - TransientSampleCluster, a per-file-set object non-unique object for a set of read-only files
//  - LogFile, a per-file (read-write or read-only) object, unique within directory-backed data
//  - Purgable data, a global singleton for the cache
//
// There are multiple goroutines that handle file I/O, and hence there is concurrent access to all
// the pieces.  To handle this we mostly use traditional locks (mutexes).
//
// Locks are in a strict hierarchy (a DAG) as follows.
//
// - ClusterStore has a lock that covers its internal data.  This lock may be held while methods are
//   being called on individual cluster objects.
//
// - Each PersistentCluster and TransientCluster has a lock that covers its internal data.  These
//   locks may be held while methods are being called on individual files.
//
// - Each LogFile has a lock.  The main lock covers the file's main mutable data structures.  This
//   lock may be held when the file code calls into the cache code, and usually it must be, so that
//   the cache code can know the state of the file.
//
// - The cache code has a lock on a global singleton data structure, the purgeable set.  This lock
//   also covers the part of the purgeable set data structure that lives in each LogFile.
//
// The cache code can call back up to the LogFile code (notably to purge a file).  In this case it
// must hold *no* nocks at all, as the LogFile code may call down into the cache code again.  Thus
// the cache code can't know for a fact that the file's state hasn't changed between the time it
// selects it for purging and the time it purges it.  It must be resilient to that.

package db

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/repr"
)

const (
	dirPermissions  = 0755
	filePermissions = 0644
)

var (
	// MT: Constant after initialization; immutable
	BadTimestampErr  = errors.New("Bad timestamp")
	ClusterClosedErr = errors.New("ClusterStore is closed")
	ReadOnlyDirErr   = errors.New("Cluster is read-only list of files")
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Various kinds of clusters and their capabilities.
//
// Cluster stores can be persistent (backed by a writable directory on disk) or transient (backed by
// a list of read-only files).  Data can be appended to a persistent store, and data read from a
// persistent store will be cached in memory for faster access.  Eventually we'll also maintain some
// intra-cluster (in-memory) indices to speed record selection.

type Cluster interface {
	// Return the config info for the cluster, if any is attached.
	Config() *config.ClusterConfig

	// Close the cluster: flush all files and mark them as closed, and mark the cluster as closed.
	// Returns when all files have been flushed.  All subsequent operations on the cluster will
	// return ClusterClosedErr.
	//
	// (This is not usually the method you want.  Persistent clusters can be synced to disk with
	// FlushAsync.)
	Close() error
}

// A SampleCluster can provide data from `sonar ps` samples: per-process samples, and per-system
// load data.
type SampleCluster interface {
	Cluster

	// Find all filenames for Sonar `ps` samples in the cluster selected by the date range and the
	// host matcher, if any.  In clusters backed by a set of read-only files, all names will be
	// returned.  Times must be UTC.
	SampleFilenames(
		fromDate, toDate time.Time,
		hosts *Hosts,
	) ([]string, error)

	// Read `ps` samples from all the files selected by SampleFilenames() and extract the
	// per-process sample data.  Times must be UTC.  The inner slices of the result, and the
	// records they point to, must not be mutated in any way.
	ReadSamples(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (sampleBlobs [][]*repr.Sample, dropped int, err error)

	// Read `ps` samples from all the files selected by SampleFilenames() and extract the load data.
	// Times must be UTC.  The inner slices of the result, and the records they point to, must not
	// be mutated in any way.
	ReadLoadData(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (dataBlobs [][]*repr.LoadDatum, dropped int, err error)

	// Read `ps` samples from all the files selected by SampleFilenames() and extract the gpu data.
	// Times must be UTC.  The inner slices of the result, and the records they point to, must not
	// be mutated in any way.
	ReadGpuData(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (dataBlobs [][]*repr.GpuDatum, dropped int, err error)
}

// A SysinfoCluster can provide `sonar sysinfo` data: per-system hardware configuration data.
type SysinfoCluster interface {
	Cluster

	// Find all filenames for Sonar `sysinfo` data in the cluster selected by the date range and the
	// host matcher, if any.  Times must be UTC.
	SysinfoFilenames(
		fromDate, toDate time.Time,
		hosts *Hosts,
	) ([]string, error)

	// Read `sysinfo` records from all the files selected by SysinfoFilenames().  Times must be UTC.
	// The inner slices of the result, and the records they point to, must not be mutated in any
	// way.
	ReadSysinfoData(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (sysinfoBlobs [][]*repr.SysinfoData, dropped int, err error)
}

// There is no HostGlobber here, as the sacct data are not mostly node-oriented.  Any analysis
// needing to filter by host should apply the host filter after reading.
type SacctCluster interface {
	Cluster

	// Find all filenames for Slurm `sacct` data in the cluster selected by the date range and the
	// host matcher, if any.  Times must be UTC.
	SacctFilenames(
		fromDate, toDate time.Time,
	) ([]string, error)

	// Read `sacct` records from all the files selected by SacctFilenames().  Times must be UTC.
	// The inner slices of the result, and the records they point to, must not be mutated in any
	// way.
	ReadSacctData(
		fromDate, toDate time.Time,
		verbose bool,
	) (recordBlobs [][]*repr.SacctInfo, dropped int, err error)
}

type CluzterCluster interface {
	Cluster

	CluzterFilenames(
		fromDate, toDate time.Time,
	) ([]string, error)

	ReadCluzterData(
		fromDate, toDate time.Time,
		verbose bool,
	) (recordBlobs [][]*repr.CluzterInfo, dropped int, err error)
}

// An AppendableCluster (not yet well developed, this could be split into appending different types
// of data) allows data to be appended to the cluster store.
//
// Read operations subsequent to append operations must provide a consistent view of the data:
// either the data before the append, or the data after.
type AppendableCluster interface {
	SampleCluster
	SysinfoCluster
	SacctCluster
	CluzterCluster

	// Trigger flushing of all pending data.  In principle the flushing is asynchronous, but
	// synchronously flushing the data is also allowed.
	FlushAsync()

	// Append data to the data store.
	//
	// `payload` is string or []byte, exclusively.  Each should in general be a single record.  The
	// payload may optionally be terminated with \n to indicate end-of-record; any embedded \n are
	// technically considered part of the record and is only allowed if the record format allows
	// that (JSON does, CSV does not).
	AppendSamplesAsync(ty FileAttr, host, timestamp string, payload any) error
	AppendSysinfoAsync(ty FileAttr, host, timestamp string, payload any) error
	AppendSlurmSacctAsync(ty FileAttr, timestamp string, payload any) error
	AppendCluzterAsync(ty FileAttr, timestamp string, payload any) error
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Cluster store API

// TODO: IMPROVEME?  These could be passed as parameters to OpenPersistentCluster and
// OpenTransientSampleCluster instead of being global variables.

var (
	// This is applied to a set of samples newly read from a file, before caching.
	// MT: Constant after initialization; immutable
	SampleRectifier func([]*repr.Sample, *config.ClusterConfig) []*repr.Sample

	// This is applied to a set of load data newly read from a file, before caching.
	// MT: Constant after initialization; immutable
	LoadDatumRectifier func([]*repr.LoadDatum, *config.ClusterConfig) []*repr.LoadDatum

	// This is applied to a set of GPU data newly read from a file, before caching.
	// MT: Constant after initialization; immutable
	GpuDatumRectifier func([]*repr.GpuDatum, *config.ClusterConfig) []*repr.GpuDatum
)

// Open a directory and attach it to the global logstore.
//
// `dir` is the root directory of the log data store for a cluster.  It contains subdirectory paths
// of the form YYYY/MM/DD for data.  At the leaf of each path are read-only data files for the given
// date.
//
// FILE NAME SCHEMES.
//
// Older data files follow these naming patterns:
//
//  - <hostname>.csv contain Sonar `ps` (ie sample) log data for the host
//  - sysinfo-<hostname>.json contain Sonar `sysinfo` system data for the host
//  - slurm-sacct.csv contains Sonar `sacct` data from Slurm for the given cluster
//
// Newer data files follow the naming pattern <version>+<type>-<originator>.json:
//
//  - 0+sample-<hostname>.json contains Sonar sample data for the host
//  - 0+sysinfo-<hostname>.json contains Sonar sysinfo data for the host
//  - 0+job-slurm.json contains Sonar slurm job data for the cluster (more general than sacct)
//  - 0+cluzter-slurm.json contains Sonar cluzter status data for the cluster
//
// In the latter scheme, `0` indicates version 0 of the new file format, for more see
// github.com/NordicHPC/sonar/util/formats/newfmt/types.go.  Extensions to the format are always
// backward compatible and require no new version number, however should there ever be reason to
// move to version 1, we can increment the file name scheme version number.  The intent is that the
// new file names have enough information to parse the contents correctly and index them coarsely
// within a given calendar day.
//
// For correctness, we assume host names cannot contain '+' (per spec they cannot).
//
// (In very old directories there may also be files `bughunt.csv` and `cpuhog.csv` that are state
// files used by some reports, these should be considered off-limits.  And note that in the old
// data, hosts cannot be named "slurm-sacct", or there will be a conflict between sacct job data and
// normal sample data.)

func OpenPersistentCluster(dir string, cfg *config.ClusterConfig) (*PersistentCluster, error) {
	return gClusterStore.openPersistentCluster(dir, cfg)
}

// Open a set of read-only files, that are not attached to the global logstore.  `files` is a
// nonempty list of files containing Sonar `ps` log data.  We make a private copy of the list.

func OpenTransientSampleCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*TransientSampleCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := sniffTypeFromFilenames(files, FileSampleCSV, FileSampleV0JSON)
	if err != nil {
		return nil, err
	}
	return newTransientSampleCluster(files, ty, cfg), nil
}

func OpenTransientSacctCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*TransientSacctCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := sniffTypeFromFilenames(files, FileSlurmCSV, FileSlurmV0JSON)
	if err != nil {
		return nil, err
	}
	return newTransientSacctCluster(files, ty, cfg), nil
}

func OpenTransientSysinfoCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*TransientSysinfoCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := sniffTypeFromFilenames(files, FileSysinfoOldJSON, FileSysinfoV0JSON)
	if err != nil {
		return nil, err
	}
	return newTransientSysinfoCluster(files, ty, cfg), nil
}

func OpenTransientCluzterCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*TransientCluzterCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return newTransientCluzterCluster(files, FileCluzterV0JSON, cfg), nil
}

func sniffTypeFromFilenames(names []string, oldType, newType FileAttr) (FileAttr, error) {
	var oldCount, newCount int
	for _, name := range names {
		if strings.HasPrefix(name, "0+") {
			newCount++
		} else {
			oldCount++
		}
	}
	if oldCount > 0 && newCount > 0 {
		return oldType, errors.New("Files must all have the same representation")
	}
	if oldCount > 0 {
		return oldType, nil
	}
	return newType, nil
}

// Drain all pending writes in the global logstore, close all the attached Cluster nodes, and return
// when it's all done.  The store is marked as closed and no operations to open new directories will
// work, nor will operations on attached clusters work.
//
// Errors are not signalled because they are not generally useful at this point but there are error
// conditions, notably in the I/O.

func Close() {
	gClusterStore.close()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// The singleton clusterStore is currently just a container for individual cluster directories.

type clusterStore struct {
	sync.Mutex
	closed   bool
	clusters map[string]*PersistentCluster
}

// MT: Constant after initialization; thread-safe
var gClusterStore clusterStore

func unsafeResetClusterStore() {
	gClusterStore = clusterStore{
		clusters: make(map[string]*PersistentCluster, 10),
	}
}

// Note, this does not reinitialize I/O goroutines or their state

func init() {
	unsafeResetClusterStore()
}

func (ls *clusterStore) openPersistentCluster(
	clusterDir string,
	cfg *config.ClusterConfig,
) (*PersistentCluster, error) {
	ls.Lock()
	defer ls.Unlock()
	if ls.closed {
		return nil, ClusterClosedErr
	}

	// Normally the path will have been cleaned by command line parsing, but do it anyway.
	clusterDir = path.Clean(clusterDir)
	if d := ls.clusters[clusterDir]; d != nil {
		return d, nil
	}

	d := newPersistentCluster(clusterDir, cfg)
	ls.clusters[clusterDir] = d
	return d, nil
}

func (ls *clusterStore) close() {
	ls.Lock()
	defer ls.Unlock()
	if ls.closed {
		return
	}

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

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// I/O goroutines.
//
// A number of background goroutines are on call to perform I/O, this helps (a lot) with reading
// large numbers of log files, a common case when starting up or moving focus to a new cluster for
// some query.

// A parseRequest is posted on parseRequests when a file needs to be parsed.  The id should be
// anything that the caller needs it to be.  The file must never be nil (a nil file means the queue
// is closed and the worker will exit).

type parseRequest struct {
	file    *LogFile
	id      any
	reader  ReadSyncMethods
	results chan parseResult
	verbose bool
}

var (
	// MT: Constant after initialization; thread-safe
	parseRequests = make(chan parseRequest, 100)
)

// A parseResult must *always* be returned in response to the parse request with a non-nil file.
// The id is the id from the request.  `data` will have the data read by the file layer but may be
// nil in the case of an error, `dropped` is the number of benignly dropped records.

type parseResult struct {
	id      any
	data    any
	dropped int
	err     error
}

// Fork off the shared parser workers.
//
// About performance:
//
// NumCPU() or NumCPU()+1 seem to be good, this brings us up to about 360% utilization on a quad
// core VM (probably backed by SSD), testing with 8w of Saga old-style `sonar ps` data.  NumCPU()-1
// is not good, nor NumCPU()*2 on this machine.  We would expect some blocking on the Ustr table,
// esp early in the run, and and some waiting for file I/O, but I've not explored these yet.
//
// Utilization with new-style Fox data - which look pretty different - is at the same level.
//
// With cold data, utilization drops to about 270%, as expected.  This is still pretty good, though
// in this case a larger number of goroutines might help some.  Cold data is in some sense the
// expected case, if caching works well, so worth exploring maybe.

func init() {
	workers := runtime.NumCPU()
	//workers = 1
	for i := 0; i < workers; i++ {
		uf := NewUstrCache()
		go Forever(
			func() {
				for {
					request := <-parseRequests
					if request.file == nil {
						return
					}
					var result parseResult
					result.id = request.id
					result.data, result.dropped, result.err =
						request.file.ReadSync(uf, request.verbose, request.reader)
					request.results <- result
				}
			},
			os.Stderr,
		)
	}
}

// Read a set of records from a set of files and return a slice of slices, normally one inner slice
// per file.  The outer slice is always freshly allocated but the inner slices, though immutable,
// are owned by the database layer, as is their underlying storage.  The objects pointed to from the
// inner slices are also shared and immutable.  The inner slices may be in any order due to
// concurrency in the database access layer (reading is implemented by by passing the read request
// to the pool of readers and collecting the results).
//
// To be safe, clients should iterate over the data structure as soon as they can and not retain
// references to the slices longer than necessary.  The inner slices *MUST* not be mutated; in
// particular, they must not be sorted.
//
// On return, `dropped` is an indication of the number of benign errors, but it conflates dropped
// records and dropped fields.  err is non-nil only for non-benign records, in which case it
// attempts to collect information about all errors encountered.
//
// TODO: IMPROVEME: The API is a little crusty.  We could distinguish dropped fields vs dropped
// records, and we could sensibly return partial results too.
//
// TODO: We could strengthen the notion of immutability of the results by wrapping the result set in
// a typesafe container that can be iterated over, esp in Go 1.23 or later.  But even defining a
// type a la ResultSet[T] and returning that would help, probably.

func readRecordsFromFiles[T any](
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (recordBlobs [][]*T, dropped int, err error) {
	if verbose {
		Log.Infof("%d files", len(files))
	}

	// TODO: OPTIMIZEME: Probably we would want to be smarter about accumulating in a big array that
	// has to be doubled in size often and may become very large (4 months of data from Saga yielded
	// about 32e6 records).

	results := make(chan parseResult, 100)
	recordBlobs = make([][]*T, 0)

	toReceive := len(files)
	toSend := 0
	bad := ""
	// Note that the appends here ensure that we never mutate the spines of the cached data arrays,
	// but always return a fresh list of records.
	for toSend < len(files) && toReceive > 0 {
		select {
		case parseRequests <- parseRequest{
			file:    files[toSend],
			results: results,
			verbose: verbose,
			reader:  reader,
		}:
			toSend++
		case res := <-results:
			if res.err != nil {
				bad += "  " + res.err.Error() + "\n"
			} else {
				recordBlobs = append(recordBlobs, res.data.([]*T))
				dropped += res.dropped
			}
			toReceive--
		}
	}
	for toReceive > 0 {
		res := <-results
		if res.err != nil {
			bad += "  " + res.err.Error() + "\n"
		} else {
			recordBlobs = append(recordBlobs, res.data.([]*T))
			dropped += res.dropped
		}
		toReceive--
	}

	if bad != "" {
		recordBlobs = nil
		err = fmt.Errorf("Failed to process one or more files: %s", bad)
		return
	}

	return
}
