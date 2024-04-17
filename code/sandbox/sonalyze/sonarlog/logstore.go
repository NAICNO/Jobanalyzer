package sonarlog

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"go-utils/filesys"
	"go-utils/hostglob"
)

type LogStore struct {
	// From OpenDir
	dataDir          string
	fromDate, toDate time.Time
	hosts            *hostglob.HostGlobber

	// From OpenFiles, or the result of processing the directory
	files []string

	// Set to true once we have a final list
	processed bool
}

// `dir` is the root directory of the log data store for a cluster.  It contains subdirectory paths
// of the form YYYY/MM/DD for data.  At the leaf of each path are read-only data files for the given
// date:
//
//  - HOSTNAME.csv contain Sonar `ps` log data for the given host
//  - sysinfo-HOSTNAME.json contain Sonar `sysinfo` system data for the given host

func OpenDir(dir string, fromDate, toDate time.Time, hosts *hostglob.HostGlobber) (*LogStore, error) {
	return &LogStore{
		dataDir:  dir,
		fromDate: fromDate,
		toDate:   toDate,
		hosts:    hosts,
	}, nil
}

// `files` is a list of files containing Sonar `ps` log data.  We make a private copy.

func OpenFiles(files []string) (*LogStore, error) {
	return &LogStore{files: append(make([]string, 0, len(files)), files...)}, nil
}

// Return the file names that will be passed to os.Open().  The slice should be considered
// immutable.  In particular, it should not be sorted.
//
// Implementation: This will alter s.files, and it will no longer have the original list
// of input file names, if any.

func (s *LogStore) Files() ([]string, error) {
	if s.processed {
		return s.files, nil
	}

	var files []string
	var err error
	if len(s.files) > 0 {
		files = s.files
	} else {
		files, err = filesys.EnumerateFiles(s.dataDir, s.fromDate, s.toDate, "*.csv")
		if err != nil {
			return nil, err
		}
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
	if s.hosts != nil && !s.hosts.IsEmpty() {
		dest := 0
		for src := 0; src < len(files); src++ {
			fn := files[src]
			base := path.Base(fn)
			if s.hosts.Match(base[:len(base)-4]) {
				files[dest] = files[src]
				dest++
			}
		}
		files = files[:dest]
	}

	if s.dataDir != "" {
		for i := range files {
			files[i] = path.Join(s.dataDir, files[i])
		}
	}

	s.files = files
	s.processed = true
	return files, nil
}

func (s *LogStore) ReadLogEntries(verbose bool) (samples SampleStream, dropped int, err error) {
	var files []string
	files, err = s.Files()
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
