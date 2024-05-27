package sonarlog

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"go-utils/filesys"
	"go-utils/hostglob"
	utime "go-utils/time"
)

const (
	dirPermissions  = 0755
	filePermissions = 0644
	newline         = 10
)

var (
	BadTimestampErr = errors.New("Bad timestamp")
	LogClosedErr = errors.New("LogStore is closed")
	ReadOnlyDirErr = errors.New("LogDir is read-only list of files")
)

// DOCUMENTME.
//
// dataDir is nonblank iff we've opened a writable directory.  dataDir must have been cleaned.
// files is non-nil iff we've opened a list of files.  The kind is redundant.

type LogDir struct {
	sync.Mutex
	closed           bool
	dataDir          string
	writers          map[string]dirWriter
	files            []string
}

type dirWriter struct {
	file *os.File
	buf  *bufio.Writer
}

func newDir(dir string) *LogDir {
	return &LogDir{
		dataDir:  dir,
		writers:  make(map[string]dirWriter),
	}
}

func newFiles(files []string) *LogDir {
	if len(files) == 0 {
		panic("Empty list of files")
	}
	return &LogDir{
		files: append(make([]string, 0, len(files)), files...),
	}
}

func (ld *LogDir) Flush() error {
	return ld.closeOrFlush(false)
}

func (ld *LogDir) closeOrFlush(isClose bool) error {
	ld.Lock()
	defer ld.Unlock()

	if ld.closed {
		return LogClosedErr
	}
	if ld.files != nil {
		return nil
	}
	ld.closed = isClose

	return ld.flush()
}

// Return the file names that will be passed to os.Open().  The slice should be considered
// immutable.  In particular, it should not be sorted.
//
// Implementation: This will alter ld.files, and it will no longer have the original list
// of input file names, if any.
//
// No caching of the enumerated file list is done here at the moment b/c once we implement caching
// of files, directory enumeration will happen only once and then a different structure will be
// built from that.

func (ld *LogDir) Files(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
) ([]string, error) {
	ld.Lock()
	defer ld.Unlock()

	if ld.closed {
		return nil, LogClosedErr
	}
	if ld.files != nil {
		return ld.files, nil
	}

	// EnumerateFiles takes an exclusive upper bound and then chops off fractional days, so
	// round up to include the files on the `to` date.
	files, err := filesys.EnumerateFiles(ld.dataDir, fromDate, utime.RoundupDay(toDate), "*.csv")
	if err != nil {
		return nil, err
	}

	// Temporary(?) backward compatible: remove cpuhog.csv and bughunt.csv, these later moved
	// into separate directory trees.
	{
		dest := 0
		for src := 0; src < len(files); src++ {
			fn := files[src]
			if strings.HasSuffix(fn, "/cpuhog.csv") || strings.HasSuffix(fn, "/bughunt.csv") {
				continue
			}
			files[dest] = files[src]
			dest++
		}
		files = files[:dest]
	}

	// Retain only file names matching the filter, if present
	if hosts != nil && !hosts.IsEmpty() {
		dest := 0
		for src := 0; src < len(files); src++ {
			fn := files[src]
			base := path.Base(fn)
			if hosts.Match(base[:len(base)-4]) {
				files[dest] = files[src]
				dest++
			}
		}
		files = files[:dest]
	}

	if ld.dataDir != "" {
		for i := range files {
			files[i] = path.Join(ld.dataDir, files[i])
		}
	}

	ld.files = files
	return files, nil
}

func (ld *LogDir) ReadLogEntries(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	verbose bool,
) (samples SampleStream, dropped int, err error) {
	ld.Lock()
	defer ld.Unlock()

	if ld.closed {
		return nil, 0, LogClosedErr
	}
	if ld.files != nil {
		// Flush all pending writes before reading
		ld.flush()
	}

	var files []string
	files, err = ld.Files(fromDate, toDate, hosts)
	if err != nil {
		return
	}
	if verbose {
		log.Printf("%d files", len(files))
	}

	// Here, NumCPU() or NumCPU()+1 seem to be good, this brings us up to about 360% utilization on
	// a quad core VM (probably backed by SSD), testing with 8w of Saga data.  NumCPU()-1 is not
	// good, nor NumCPU()*2 on this machine.  We would expect some blocking on the Ustr table, esp
	// early in the run, and and some waiting for file I/O, but I've not explored these yet.
	//
	// Utilization with Fox data - which look pretty different - is at the same level.
	//
	// With cold data, utilization drops to about 270%, as expected.  This is still pretty good,
	// though in this case a larger number of goroutines might help some.
	numRoutines := runtime.NumCPU()

	pendingFiles := make(chan string, len(files))
	type resultrec struct {
		rs  SampleStream
		dr  int
		err error
	}
	results := make(chan *resultrec, len(files))

	for i := 0; i < numRoutines; i++ {
		go (func() {
			uf := NewUstrCache()
			for {
				fn := <-pendingFiles
				if fn == "" {
					return
				}
				res := new(resultrec)
				inputFile, err := os.Open(fn)
				if err == nil {
					samples, badRecords, err := ParseSonarLog(inputFile, uf, verbose)
					if err == nil {
						res.rs = samples
						res.dr = badRecords
					}
					inputFile.Close()
				} else {
					res.err = err
				}
				results <- res
			}
		})()
	}

	samples = make(SampleStream, 0)

	// TODO: OPTIMIZEME: Merge these loops using select, then reduce the sizes of the channels.
	//
	// TODO: OPTIMIZEME: Probably we would want to be smarter about accumulating in a big array that
	// has to be doubled in size often and may become very large (4 months of data from Saga yielded
	// about 32e6 records).

	for _, fn := range files {
		pendingFiles <- fn
	}

	bad := ""
	for _, _ = range files {
		res := <-results
		if res.err != nil {
			bad += "  " + res.err.Error() + "\n"
		} else {
			samples = append(samples, res.rs...)
			dropped += res.dr
		}
	}

	// Make the goroutines exit.
	close(pendingFiles)

	if bad != "" {
		samples = nil
		err = fmt.Errorf("Failed to process one or more files:\n%s", bad)
		return
	}

	return
}

// Append the data to the file (adding termination or other formatting as necessary by the file
// format).
//
// This caches the open file and its buffered writer, because the normal situation is that there
// will be many records per node and date when a batch of records arrives.
//
// The `format` must be a format string with exactly one %s parameter, it should produce a full file
// name with extension from a node name.
//
// When BadTimestampErr is returned, the record can in principle be dropped silently.  All other
// errors are basically I/O errors.

func (ld *LogDir) AppendBytes(host, timestamp, format string, payload []byte) error {
	ld.Lock()
	defer ld.Unlock()

	if ld.closed {
		return LogClosedErr
	}
	if ld.files != nil {
		return ReadOnlyDirErr
	}
	if len(payload) == 0 {
		return nil
	}

	w, err := ld.getWriter(host, timestamp, format)
	if err != nil {
		return err
	}

	_, err = w.buf.Write(payload)
	if err != nil {
		return fmt.Errorf("Failed to append to file (%v)", err)
	}

	if payload[len(payload)-1] != newline {
		err := w.buf.WriteByte(newline)
		if err != nil {
			return fmt.Errorf("Failed to append to file (%v)", err)
		}
	}

	return nil
}

func (ld *LogDir) AppendString(host, timestamp, format string, payload string) error {
	ld.Lock()
	defer ld.Unlock()

	if ld.closed {
		return LogClosedErr
	}
	if ld.files != nil {
		return ReadOnlyDirErr
	}
	if len(payload) == 0 {
		return nil
	}

	w, err := ld.getWriter(host, timestamp, format)
	if err != nil {
		return err
	}

	_, err = w.buf.WriteString(payload)
	if err != nil {
		return fmt.Errorf("Failed to append to file (%v)", err)
	}

	if payload[len(payload)-1] != newline {
		err := w.buf.WriteByte(newline)
		if err != nil {
			return fmt.Errorf("Failed to append to file (%v)", err)
		}
	}

	return nil
}

// Pre: LOCK HELD
// Pre: ld.files == nil
func (ld *LogDir) flush() error {
	var err error

	if len(ld.writers) == 0 {
		return nil
	}

	for _, w := range ld.writers {
		err = errors.Join(err, w.buf.Flush())
		err = errors.Join(err, w.file.Close())
	}
	// TODO: In Go 1.21, we can use clear()
	ld.writers = make(map[string]dirWriter)

	return err
}

// Pre: LOCK HELD
// Pre: !ld.closed
// Pre: ld.files == nil
func (ld *LogDir) getWriter(host, timestamp, format string) (dirWriter, error) {
	// The path will be (below ld.dataDir) yyyy/mm/dd/$FILENAME
	tval, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return dirWriter{}, BadTimestampErr
	}
	if ld.files != nil {
		return dirWriter{}, ReadOnlyDirErr
	}

	// TODO: OPTIMIZEME: Mostly, dirname (indeed the timestamp) will be the same for many records in
	// the same bundle.  We may want to optimize that by caching the timestamp and avoiding the
	// parsing and mkdir.
	dirname := fmt.Sprintf("%04d/%02d/%02d", tval.Year(), tval.Month(), tval.Day())
	err = os.MkdirAll(path.Join(ld.dataDir, dirname), dirPermissions)
	if err != nil {
		return dirWriter{}, fmt.Errorf("Failed to create path (%v)", err)
	}

	filename := path.Join(ld.dataDir, dirname, fmt.Sprintf(format, host))
	if probe, ok := ld.writers[filename]; ok {
		return probe, nil
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		// Could be disk full, fs went away, file is directory, wrong permissions
		//
		// Could also be too many open files, in which case we really want to close all open
		// files and retry.
		return dirWriter{}, fmt.Errorf("Failed to open/create file (%v)", err)
	}
	w := dirWriter{file: f, buf: bufio.NewWriter(f)}
	ld.writers[filename] = w
	return w, nil
}
