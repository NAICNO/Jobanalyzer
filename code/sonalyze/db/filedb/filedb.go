package filedb

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	. "sonalyze/common"
)

const (
	filePermissions = 0644
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Misc utilities

func SniffTypeFromFilenames(names []string, oldType, newType FileAttr) (FileAttr, error) {
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

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// I/O primitives

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
