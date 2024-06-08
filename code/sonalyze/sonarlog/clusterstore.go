// ClusterStore - API to the underlying data logs
//
// The cluster store is a global singleton that manages reading, appending, and (eventually)
// transparent caching.  Clients open individual data (cluster) directories by calling
// OpenPersistentCluster() or sets of read-only sonar log input files by calling
// OpenTransientSampleCluster().  The returned Cluster structure will then manage file I/O within
// that space.
//
// The main thread would normally `defer sonarlog.Close()` to make sure that all pending writes are
// done when the program shuts down.

package sonarlog

import (
	"errors"
	"fmt"
	"log"
	"path"
	"runtime"
	"sync"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	. "sonalyze/common"
)

const (
	dirPermissions  = 0755
	filePermissions = 0644
)

var (
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
	// Close the cluster: flush all files and mark them as closed, and mark the cluster as closed.
	// Returns when all files have been flushed.  All subsequent operations on the cluster will
	// return ClusterClosedErr.
	//
	// (This is not usually the method you want.  Persistent clusters can be synced to disk with
	// FlushAsync.)
	Close() error
}

// A SampleCluster can provide `sonar ps` samples.
type SampleCluster interface {
	Cluster

	// Find all filenames for Sonar `ps` samples in the cluster selected by the date range and the
	// host matcher, if any.  In clusters backed by a set of read-only files, all names will be
	// returned.  Times must be UTC.
	SampleFilenames(
		fromDate, toDate time.Time,
		hosts *hostglob.HostGlobber,
	) ([]string, error)

	// Read `ps` samples from all the files selected by SampleFilenames().  Times must be UTC.
	ReadSamples(
		fromDate, toDate time.Time,
		hosts *hostglob.HostGlobber,
		verbose bool,
	) (samples SampleStream, dropped int, err error)
}

// A SysinfoCluster can provide `sonar sysinfo` data.
type SysinfoCluster interface {
	Cluster

	// Find all filenames for Sonar `sysinfo` data in the cluster selected by the date range and the
	// host matcher, if any.  Times must be UTC.
	SysinfoFilenames(
		fromDate, toDate time.Time,
		hosts *hostglob.HostGlobber,
	) ([]string, error)

	// Read `sysinfo` records from all the files selected by SysinfoFilenames().  Times must be UTC.
	ReadSysinfo(
		fromDate, toDate time.Time,
		hosts *hostglob.HostGlobber,
		verbose bool,
	) (records []*config.NodeConfigRecord, dropped int, err error)
}

// An AppendableCluster (not yet well developed, this could be split into appending different types
// of data) allows data to be appended to the cluster store.
//
// Read operations subsequent to append operations must provide a consistent view of the data:
// either the data before the append, or the data after.
type AppendableCluster interface {
	SampleCluster
	SysinfoCluster

	// Trigger flushing of all pending data.  In principle the flushing is asynchronous, but
	// synchronously flushing the data is also allowed.
	FlushAsync()

	// Append data to the data store.
	//
	// `payload` is string or []byte, exclusively.  Each should in general be a single record.  The
	// payload may optionally be terminated with \n to indicate end-of-record; any embedded \n are
	// technically considered part of the record and is only allowed if the record format allows
	// that (JSON does, CSV does not).
	AppendSamplesAsync(host, timestamp string, payload any) error
	AppendSysinfoAsync(host, timestamp string, payload any) error
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Cluster store API

// Open a directory and attach it to the global logstore.
//
// `dir` is the root directory of the log data store for a cluster.  It contains subdirectory paths
// of the form YYYY/MM/DD for data.  At the leaf of each path are read-only data files for the given
// date:
//
//  - HOSTNAME.csv contain Sonar `ps` log data for the given host
//  - sysinfo-HOSTNAME.json contain Sonar `sysinfo` system data for the given host
//  - in older directories there may also be files `bughunt.csv` and `cpuhog.csv` that are state
//    files used by some reports, these should be considered off-limits

func OpenPersistentCluster(dir string) (*PersistentCluster, error) {
	return gClusterStore.openPersistentCluster(dir)
}

// Open a set of read-only files, that are not attached to the global logstore.  `files` is a
// nonempty list of files containing Sonar `ps` log data.  We make a private copy of the list.

func OpenTransientSampleCluster(files []string) (*TransientSampleCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return newTransientSampleCluster(files), nil
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

func (ls *clusterStore) openPersistentCluster(clusterDir string) (*PersistentCluster, error) {
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

	d := newPersistentCluster(clusterDir)
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
	results chan parseResult
	verbose bool
}

var (
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
	for i := 0; i < runtime.NumCPU(); i++ {
		go (func() {
			uf := NewUstrCache()
			for {
				request := <-parseRequests
				if request.file == nil {
					return
				}
				var result parseResult
				result.id = request.id
				result.data, result.dropped, result.err = request.file.ReadSync(uf, request.verbose)
				request.results <- result
			}
		})()
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Sundry helpers on sets of files.
//
// The files must not be locked in the caller because the methods on the files called by the code
// below (directly or indirectly) may wish to lock them.

// For a list of files, produce a list of full names (not necessarily absolute names, though)

func filenames(files []*LogFile) []string {
	names := make([]string, len(files))
	for i, fn := range files {
		names[i] = fn.fullName()
	}
	return names
}

// Read a set of records from a set of files and return the resulting list, which may be in any
// order (b/c concurrency).  We do this by passing read request to the pool of readers and
// collecting the results.
//
// On return, `dropped` is an indication of the number of benign errors, but it conflates dropped
// records and dropped fields.  err is non-nil only for non-benign records, in which case it
// attempts to collect information about all errors encountered.
//
// TODO: IMPROVEME: The API is a little crusty.  We could distinguish dropped fields vs dropped
// records, and we could sensibly return partial results too.

func readRecordsFromFiles[T any](
	files []*LogFile,
	verbose bool,
) (records []*T, dropped int, err error) {
	if verbose {
		log.Printf("%d files", len(files))
	}

	// TODO: OPTIMIZEME: Probably we would want to be smarter about accumulating in a big array that
	// has to be doubled in size often and may become very large (4 months of data from Saga yielded
	// about 32e6 records).

	results := make(chan parseResult, 100)
	records = make([]*T, 0)

	toReceive := len(files)
	toSend := 0
	bad := ""
	for toSend < len(files) && toReceive > 0 {
		select {
		case parseRequests <- parseRequest{
			file:    files[toSend],
			results: results,
			verbose: verbose,
		}:
			toSend++
		case res := <-results:
			if res.err != nil {
				bad += "  " + res.err.Error() + "\n"
			} else {
				records = append(records, res.data.([]*T)...)
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
			records = append(records, res.data.([]*T)...)
			dropped += res.dropped
		}
		toReceive--
	}

	if bad != "" {
		records = nil
		err = fmt.Errorf("Failed to process one or more files:\n%s", bad)
		return
	}

	return
}

// Read sonar `ps` samples.

func readSamples(
	files []*LogFile,
	verbose bool,
) (samples SampleStream, dropped int, err error) {
	records, dropped, err := readRecordsFromFiles[Sample](files, verbose)
	samples = SampleStream(records)
	return
}

// Read sonar `sysinfo` records.

func readSysinfo(
	files []*LogFile,
	verbose bool,
) (records []*config.NodeConfigRecord, dropped int, err error) {
	records, dropped, err = readRecordsFromFiles[config.NodeConfigRecord](files, verbose)
	return
}
