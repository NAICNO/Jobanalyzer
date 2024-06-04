// LogFile - API to individual log files
//
// Each LogFile is backed by a particular disk file.  If the file is appendable then there is a
// unique LogFile in the system representing the file; read-only files need not be unique.  But it's
// helpful to keep memory usage down if read-only cacheable files are unique.
//
// When a file is appended to, the data to append are added to a list in the LogFile object but no
// further action is taken.  The file has to be flushed by external action, typically by the Cluster
// marking the file as dirty and performing periodic flushes of dirty data.
//
// Note THERE IS NO FINALIZATION, if a dirty file is dropped on the floor without being flushed its
// data will not be written.
//
// A file may cache its data, mostly transparently - in this case, a read operation returns the
// cached data.  See below.
//
// The files are kept generic through the use of `any`.  We could instead have created a hierarchy
// of interfaces and/or used generic types but that currently seems like needless complexity.
//
// A file is in one of three states (when it is not locked):
//
// (A) on disk, no output pending
// (B) on disk + in cache, no output pending
// (C) on disk, output pending
//
// This implies that if a cached files is appended to, it is first purged from the cache.  This is a
// simplifying assumption that is probably OK, because most files are read-only and will stay in
// cache.
//
// TODO: OPTIMIZEME: If it turns out to be not OK to not cache dirty files (for reasons of poor
// performance due to too many re-reads of data) we can change it so that appended records are
// queued for output and also appended to the in-memory representation.
//
// There is a soft limit on the number of sample records in memory, this is controlled by a command
// line switch.  When this switch is present, the soft limit is in effect.  When the switch is not
// present then no records are held in memory, the file will never be in cache (state (B)).
//
// When a file has been read and is to be cached, we compute its occupancy.  If that + the current
// occupancy of the store exceeds the limit then some older records in the cache have to be purged.
// We purge entire files.  Purging is by 2-random LRU: pick two candidates at random in the cache,
// then remove the least recently used of the two.  Repeat until we're below the limit again.
//
// Purging will not affect records that have been read and are being used by ongoing operations.  So
// the amount of data in memory may for a time exceed the cache limit.

package db

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

const (
	newline = 10
)

type fileAttr int

const (
	fileCacheable fileAttr = 1 << iota
	fileAppendable
	fileSonarSamples // `sonar ps` data
	fileSonarSysinfo // `sonar sysinfo` data
)
const (
	fileContentMask = fileSonarSamples | fileSonarSysinfo
)

var (
	// This is applied to a set of samples newly read from a file, before caching.
	// MT: Constant after initialization; immutable
	SampleRectifier func([]*Sample, *config.ClusterConfig) []*Sample
)

// The components of the Fullname are broken out so as to allow strings to be shared as much as
// possible.  If we were to catenate them up front we'd be creating a lot of unique strings that
// would be held permanently.
//
// TODO: OPTIMIZEME: These strings could usefully be Ustr, though even with 1M files in memory it
// probably won't matter very much?

type Fullname struct {
	cluster  string // .../cluster
	dirname  string // yyyy/mm/dd or other subdir name (when appropriate)
	basename string // hostname.csv, sysinfo-hostname.json, ...
}

func (fn *Fullname) fullName() string {
	return path.Join(fn.cluster, fn.dirname, fn.basename)
}

type sampleCachePayload struct {
	samples []*Sample
	dropped int
}

type LogFile struct {
	// Fullname is immutable and its components can be accessed without holding the lock, and the
	// fullName() method of the LogFile will not take the lock.
	Fullname

	sync.Mutex
	attrs   fileAttr // immutable for now but may store cache metadata?
	pending []any    // string or []byte

	// Cache data owned by the caching code, protected by the LogFile's mutex
	logFileCacheData

	// Cache data owned by the cache purging code, protected by the purgeLock
	logFilePurgeableData
}

func newLogFile(fn Fullname, attrs fileAttr) *LogFile {
	if (attrs & fileContentMask) == 0 {
		panic("Content type must be set")
	}
	return &LogFile{
		Fullname: fn,
		attrs:    attrs,
	}
}

func (lf *LogFile) AppendAsync(payload any) error {
	if (lf.attrs & fileAppendable) == 0 {
		panic("Read-only file")
	}
	switch x := payload.(type) {
	case []byte:
		if len(x) == 0 {
			return nil
		}
	case string:
		if len(x) == 0 {
			return nil
		}
	default:
		return errors.New("Payload must be string or []byte")
	}

	lf.Lock()
	defer lf.Unlock()

	// Purge the cache here because writes are pending.  We would do this anyway in ReadSync and
	// this eases cache pressure earlier.
	//
	// TODO: OPTIMIZEME: Optimize ReadSync so that appendable data can be appended to cached data
	// without re-reading everything.
	lf.cachePurgeLocked("internal:dirty")
	if lf.pending == nil {
		lf.pending = make([]any, 0, 5)
	}
	lf.pending = append(lf.pending, payload)
	return nil
}

var (
	// MT: Constant after initialization; immutable
	perSampleSize int64
)

func init() {
	var s Sample
	perSampleSize = int64(unsafe.Sizeof(s) + unsafe.Sizeof(&s))
}

// The `data` is concretely a []T, specifically not a type with ~[]T.  Generic reader code depends
// on this to collect read results.

func (lf *LogFile) ReadSync(
	uf *UstrCache,
	verbose bool,
	cfg *config.ClusterConfig,
) (data any, badRecords int, err error) {
	lf.Lock()
	defer lf.Unlock()

	// Flush pending data before reading.
	if len(lf.pending) != 0 && lf.isCachedLocked() {
		// There should be nothing in the cache.  See comment in AppendAsync re optimizing this.
		Log.Warningf("cache: File should not have cached data.")
	}
	err = lf.flushSyncLocked()
	if err != nil {
		return
	}

	switch {
	case (lf.attrs & fileSonarSamples) != 0:
		if isCached, cachedData := lf.cacheReadLocked(); isCached {
			c := cachedData.(*sampleCachePayload)
			data, badRecords = c.samples, c.dropped
		} else {
			var inputFile *os.File
			inputFile, err = os.Open(lf.fullName())
			if err == nil {
				defer inputFile.Close()
				var samples []*Sample
				samples, badRecords, err = ParseSonarLog(inputFile, uf, verbose)
				if SampleRectifier != nil {
					// There is a tricky invariant here that we're not going to check: If data are
					// cached, then we require not just that there is a ClusterConfig for the
					// cluster, but config info for each node. This is so that we can rectify the
					// input data based on system info.  If there is no config for the cluster or
					// the code then these data may remain wonky.
					//
					// The reason we don't check the invariant is that there effects of not having a
					// config are fairly benign, and also that so much else depends on having a
					// config that we'll get a more thorough check in other ways.
					samples = SampleRectifier(samples, cfg)
				}
				data = samples
				if err == nil && CacheEnabled() {
					lf.cacheWriteLocked(
						&sampleCachePayload{samples, badRecords},
						perSampleSize*int64(len(samples)),
					)
				}
			}
		}
	case (lf.attrs & fileSonarSysinfo) != 0:
		var inputFile *os.File
		inputFile, err = os.Open(lf.fullName())
		if err == nil {
			defer inputFile.Close()
			data, err = ParseSysinfoLog(inputFile, verbose)
		}
	default:
		panic("Unknown content type")
	}

	return
}

// Reason codes with the prefix "internal:" are reserved for internal use by the file layer.
func (lf *LogFile) PurgeCache(reason string) {
	lf.Lock()
	defer lf.Unlock()
	lf.cachePurgeLocked(reason)
}

func (lf *LogFile) FlushSync() error {
	lf.Lock()
	defer lf.Unlock()

	return lf.flushSyncLocked()
}

func (lf *LogFile) flushSyncLocked() (err error) {
	if len(lf.pending) == 0 {
		return nil
	}
	items := lf.pending
	lf.pending = make([]any, 0, 5)

	// We assume that the directory of the file exits because it's the responsibility of the
	// PersistentCluster to create it when the LogFile is created.
	f, err := os.OpenFile(lf.fullName(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		// Could be disk full, fs went away, file is directory, wrong permissions
		//
		// Could also be too many open files, in which case we really want to close all open
		// files and retry.
		err = fmt.Errorf("Failed to open/create file (%v)", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, item := range items {
		needNewline := false
		switch x := item.(type) {
		case string:
			_, err = w.WriteString(x)
			needNewline = x[len(x)-1] != newline
		case []byte:
			_, err = w.Write(x)
			needNewline = x[len(x)-1] != newline
		}
		if err == nil && needNewline {
			err = w.WriteByte(newline)
		}
		if err != nil {
			return
		}
	}
	return
}
