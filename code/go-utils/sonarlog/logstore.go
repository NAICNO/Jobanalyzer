package sonarlog

import (
	"os"
	"path"
	"runtime"
	"time"

	"go-utils/filesys"
	"go-utils/hostglob"
)

type LogStore struct {
	dataDir string
}

// `dir` is the root directory of the log data store for a cluster.  It contains subdirectory paths
// of the form YYYY/MM/DD for data.  At the leaf of each path are read-only data files for the given
// date:
//
//  - HOSTNAME.csv contain Sonar `ps` log data for the given host
//  - sysinfo-HOSTNAME.json contain Sonar `sysinfo` system data for the given host

func OpenDir(dir string) (*LogStore, error) {
	return &LogStore{dataDir: dir}, nil
}

// In sonalyze, the source and record filters have these options:
//
// - from/to
// - host include, default all
// - user include/exclude
// - command include/exclude
// - job ID include/exclude
//
// The dates and the host can be applied at the file system level under reasonable assumptions.  The
// others require the record to be read and parsed.
//
// One design point here is about the recordFilter.  We must choose whether it must be thread-safe
// and can be applied in parallel while the result list is being constructed, or whether it need not
// be thread-safe and is applied in the master thread.  Probably the latter is better?
//
// A very discriminating recordFilter suggests that records should be allocated singly on the heap,
// not inside a larger slice.
//
// The recordFilter is necessarily primitive: it is *not* a job filter (though the Job ID is a
// thing) and not all fields of the record will make sense unless the record has been rectified
// first.  So we need to be clear about which fields the recordFilter can examine.

func (s *LogStore) LogEntries(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	recordFilter func(*Sample) bool,
	verbose bool,
) (readings SampleStream, dropped int, err error) {
	files, err := filesys.EnumerateFiles(s.dataDir, fromDate, toDate, "*.csv")
	if err != nil {
		return
	}
	if hosts != nil {
		dest := 0
		for src := 0; src < len(files); src++ {
			fn := files[src]
			if hosts.Match(fn[:len(fn)-4]) {
				files[dest] = files[src]
				dest++
			}
		}
		files = files[:dest]
	}
	if verbose {
		println(len(files), " files")
	}
	pendingFiles := make(chan string, len(files))
	type resultrec struct {
		rs SampleStream
		dr int
	}
	results := make(chan *resultrec, len(files))
	for i := 0; i < runtime.NumCPU(); i++ {
		go (func() {
			uf := NewUstrCache()
			for {
				fn := <-pendingFiles
				res := new(resultrec)
				inputFile, err := os.Open(fn)
				if err == nil {
					readings, badRecords, err := ParseSonarLog(inputFile, uf)
					if err == nil {
						res.rs = readings
						res.dr = badRecords
					}
					inputFile.Close()
				}
				results <- res
			}
		})()
	}

	readings = make(SampleStream, 0)

	// TODO: Merge these loops using select, then reduce the sizes of the channels.

	for _, fn := range files {
		pendingFiles <- path.Join(s.dataDir, fn)
	}

	// TODO: If recordFilter is not nil, it needs to be applied somewhow.

	for _, _ = range files {
		res := <-results
		readings = append(readings, res.rs...)
		dropped += res.dr
	}

	// TODO: there's still some basic cleanup of data (a la logclean) to do probably

	// TODO: make sure the goroutines exit, they are holding onto resources

	return
}
