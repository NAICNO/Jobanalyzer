// Jobanalyzer data store abstraction (evolving)
//
// The data storage tree is rooted at some directory $DATA with the following structure:
//
//    $DATA/$CLUSTERNAME/$YEAR/$MONTH/$DAY/datafile
//
// where $MONTH and $DAY are always two decimal digits and `datafile` has one of several forms:
//
//    <hostname>.csv               Sonar log data (free CSV form)
//    sysinfo-<hostname>.json      Sysinfo log data (unseparated JSON objects?)
//
// Notably *all* csv files in this tree are Sonar log data.

// datastore.Store is a a representation of the data store.
//
// Its Write() method writes records to files in the data storage tree, possibly with retries.
//
// AT THE MOMENT, there will be only one ingestor running on the server, so data files have only one
// writing process, and we don't need a lock in the file system for concurrent writing.  There is a
// danger that there's a reader while we're writing but that danger is already there and should be
// fixed separately.
//
// However, each HTTP request in the ingestor is served on a separate goroutine and we don't want to
// have to deal with mutexing the log files *internally* anywhere, so we run a single writer on a
// dedicated goroutine, behind the scenes.

package datastore

import (
	"fmt"
	"os"
	"path"
	"time"

	"go-utils/status"
)

const (
	maxWriteAttempts  = 6
	writeRetryMinutes = 5
	channelCapacity   = 1000
	dirPermissions    = 0755
	filePermissions   = 0644
)

type Store struct {
	dataPath        string
	dataChannel     chan *dataRecord
	dataStopChannel chan bool
	verbose         bool
}

type dataRecord struct {
	dirname  string // path underneath w.dataPath
	filename string // path underneath w.dataPath
	payload  []byte // text to write
	attempts int    // number of write attempts for this record so far
}

func Open(dataPath string, verbose bool) *Store {
	w := &Store{
		dataPath:        dataPath,
		dataChannel:     make(chan *dataRecord, channelCapacity),
		dataStopChannel: make(chan bool),
		verbose:         verbose,
	}
	go w.runWriter()
	return w
}

func (w *Store) Close() {
	// Send nil on dataChannel to make the writer exit its loop in an orderly way.  Wait for a
	// response from the writer on stopChannel, and we are done.
	//
	// We don't care about any of the pending retry writes - their sleeping goroutines will either be
	// terminated when the program exits, or will wake up and send data on a channel nobody's listening
	// on before that.  Either way this is invisible.
	w.dataChannel <- nil
	<-w.dataStopChannel
}

// Here the `format` must be a format string with exactly one %s parameter, it should produce a full
// file name with extension from a node name.
//
// writeRecord is infallible, all operations that could fail are performed in the runWriter loop.

func (w *Store) Write(cluster, host, timestamp, format string, payload []byte) {
	// The path will be (below dataPath) cluster/year/month/day/$FILENAME
	tval, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		if w.verbose {
			status.Warningf("Bad timestamp %s, dropping record", timestamp)
			return
		}
	}
	dirname := fmt.Sprintf("%s/%04d/%02d/%02d", cluster, tval.Year(), tval.Month(), tval.Day())
	filename := fmt.Sprintf("%s/%s", dirname, fmt.Sprintf(format, host))

	const newline = 10

	if len(payload) > 0 {
		if payload[len(payload)-1] != newline {
			payload = append(payload, newline)
		}
		w.dataChannel <- &dataRecord{dirname, filename, payload, 0}
	}
}

// TODO: Optimization: Cache open files for a time.

func (w *Store) runWriter() {
	for {
		r := <-w.dataChannel
		if r == nil {
			break
		}
		r.attempts++
		if w.verbose {
			fmt.Printf("Storing: %s", string(r.payload))
		}

		err := os.MkdirAll(path.Join(w.dataPath, r.dirname), dirPermissions)
		if err != nil {
			// Could be disk full, fs went away, element of path exists as file, wrong permissions
			w.maybeRetryWrite(r, fmt.Sprintf("Failed to create path (%v)", err))
			return
		}
		f, err := os.OpenFile(
			path.Join(w.dataPath, r.filename),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			filePermissions,
		)
		if err != nil {
			// Could be disk full, fs went away, file is directory, wrong permissions
			w.maybeRetryWrite(r, fmt.Sprintf("Failed to open/create file (%v)", err))
			return
		}
		n, err := f.Write(r.payload)
		f.Close()
		if err != nil {
			if n == 0 {
				// Nothing was written so try to recover by restarting.
				w.maybeRetryWrite(r, fmt.Sprintf("Failed to write file (%v)", err))
			} else {
				// Partial data were written.
				//
				// The usual and benign reason for a partial write on Unix is that a signal was
				// delivered in the middle of a write and the write needs to restart with the rest
				// of the data; this is signalled with an EINTR error return from write(2).  The Go
				// libraries try to hide that problem - see internal/poll/fd_unix.go in the Go
				// sources, the function Write() is the normal destination for file output.  It
				// calls ignoringEINTRIO(syscall.Write) to perform the write, and that in turn will
				// restart the write in the case of EINTR.
				//
				// Of course, "transparently restarting after writing some data" in O_APPEND mode is
				// a complete fiction, but what can you do.
				//
				// Anyway, if we get here with a partial data write it's going to be something more
				// serious than EINTR, such as a disk full.  Trying to recover is probably not worth
				// our time.  Just log the failure and hope somebody sees it.
				status.Errorf("Write error on log (%v), %d bytes written of %d",
					err, n, len(r.payload))
			}
		}
	}
	w.dataStopChannel <- true
}

func (w *Store) maybeRetryWrite(r *dataRecord, msg string) {
	if r.attempts < maxWriteAttempts {
		if w.verbose {
			status.Info(msg + ", retrying later")
		}
		go func() {
			// Obviously some kind of backoff is possible, but do we care?
			time.Sleep(time.Duration(writeRetryMinutes * time.Minute))
			w.dataChannel <- r
		}()
	} else {
		status.Warning(msg + ", too many retries, abandoning")
	}
}
