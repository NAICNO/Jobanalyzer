// LogFile - API to individual log files
//
// Each LogFile is backed by a particular disk file.  If the file is appendable then there is a
// unique LogFile in the system representing the file; read-only files need not be unique.  But it's
// useful for memory use if read-only cacheable files are unique.
//
// When a file is appended to, the data to append is added to a list in the LogFile object but no
// further action is taken.  The file has to be flushed by external action, typically by the Cluster
// marking the file as dirty and performing periodic flushes of dirty data.
//
// A file may cache its data, mostly transparently - in this case, a read operation returns the
// cached data.  (Not yet implemented.)
//
// The files are kept generic through the use of `any`.  We could instead have created a hierarchy
// of interfaces and/or used templated types but that currently seems like needless complexity.
//
// ---- stop reading here ----
//
// (Clean this up - not all of it belongs here - it's about the cached state, not yet implemented.)
//
// A file is in one of several states:
//
// - on disk, no output pending
// - on disk, output pending
// - in memory, no output pending
// - in memory, output pending
// - locked, state transition ongoing
//
// A file can be appended to whether it is on disk or in memory, the output is queued and written
// asynchronously.
//
// When a file is read, it is always brought in from disk to memory.  This does not require flushing
// its output first.  (Though it would be nice, I guess, to do so.)
//
// A file that is on disk with output pending may not actually exist on the disk yet.
//
// There may be a soft limit on the number of sample records in memory, this is controlled by a
// command line switch.  When this switch is present, the soft limit is in effect.  When the switch
// is not present then no records are held in memory, the file will never be in-memory.
//
// When a file has been read and is to be cached, we compute its occupancy O.  If O + the current
// occupancy of the store, C, exceeds the limit then some older records in the cache have to be
// purged.  We purge entire files.  Purging is by random-LRU: pick some candidates at random in the
// cache, then remove the LRU order.  Repeat until we're below the limit again.
//
// Purging will not affect records that have been obtained by ongoing operations: They retain their
// records unchanged.  So the amount of data in memory may for a time exceed the limit.

package sonarlog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

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

type LogFile struct {
	// Fullname is immutable and its components can be accessed without holding the lock, and the
	// fullName() method of the LogFile will not take the lock.
	Fullname

	sync.Mutex
	attrs   fileAttr // immutable for now but may store cache metadata?
	pending []any    // string or []byte
	// cached    any
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
	if lf.pending == nil {
		lf.pending = make([]any, 0, 5)
	}
	lf.pending = append(lf.pending, payload)
	return nil
}

// The `data` is either []*Sample or []*SomethingElse, but it is specifically not eg SampleStream.

func (lf *LogFile) ReadSync(
	uf *UstrCache,
	verbose bool,
) (data any, badRecords int, err error) {
	lf.Lock()
	defer lf.Unlock()

	err = lf.flushSyncLocked()
	if err != nil {
		return
	}

	inputFile, err := os.Open(lf.fullName())
	if err == nil {
		defer inputFile.Close()
		switch {
		case (lf.attrs & fileSonarSamples) != 0:
			data, badRecords, err = ParseSonarLog(inputFile, uf, verbose)
		case (lf.attrs & fileSonarSysinfo) != 0:
			data, err = ParseSysinfoLog(inputFile, verbose)
		default:
			panic("Unknown content type")
		}
	}

	return
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
