// An abstraction that reads and cleans up job log files with possibly-overlapping and
// partly-redundant data records, as for the cpuhog and deadweight reports.
//
// The output of ReadJoblogFiles is a list of hosts, and for each host a list of individual jobs for
// that host (the former sorted ascending by name, the latter descending by LastSeen() timestamp).
// Notably job IDs may be reused in this list: it's a fact of life.  But they are still different
// jobs.  Within the list, `IsExpired` is false for the first (most recent) record that has each job
// ID, and true for all subsequent records with that ID.
//
// We can assume that the log records can be ingested and partitioned by host and then bucketed by
// timestamp and the buckets sorted, and if a job ID is present in two consecutive buckets in the
// list of buckets then it's the same job, and otherwise those are two different jobs with a reused
// ID.  The assumption is valid because there is a requirement stated throughout the system,
// including in mlcpuhog.go, mldeadweight.go, the shell scripts, and the cron jobs, that the
// analysis producing the logs runs often enough for the assumption to hold.

package joblog

import (
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"time"

	"naicreport/storage"
)

// The jobs being read from the logs satisfy this interface

type Job interface {
	GetId() uint32
	SetId(id uint32)
	GetHost() string
	GetLastSeen() time.Time
	IsExpired() bool
	SetExpired(flag bool)
}

// GenericJob implements Job and can be used as an embedded field by clients when that makes sense,
// but there are no dependencies in this package about its use.  (Though things might be a little
// simpler if we could make use of a GenericJob always being embedded in the concrete type of Job.)

type GenericJob struct {
	Id uint32
	Host string
	LastSeen time.Time
	Expired bool
}

func (s *GenericJob) GetId() uint32 {
	return s.Id
}
func (s *GenericJob) SetId(id uint32) {
	s.Id = id
}
func (s *GenericJob) GetHost() string {
	return s.Host
}
func (s *GenericJob) GetLastSeen() time.Time {
	return s.LastSeen
}
func (s *GenericJob) IsExpired() bool {
	return s.Expired
}
func (s *GenericJob) SetExpired(flag bool) {
	s.Expired = flag
}

// Host is redundantly stored in the Jobs field: the value is always the same in JobsByHost and in
// the individual jobs.

type JobsByHost[T Job] struct {
	Host string
	Jobs []T
}

type bucket_t[T Job] []T
type bucketList_t[T Job] []bucket_t[T]

func ReadJoblogFiles[T Job](
    // root of data directory
	dataPath string,

	// filename without a path, eg "cpuhog.csv"
	logFilename string,

	// the window of time we're interested in
	from, to time.Time,

	// true if we want some stderr diagnostics
	verbose bool,

	// Parse a single csv string->string map.  This must set LastSeen=timestamp, IsExpired=false,
	// and then other fields as required by integrateRecords.
	parseLogRecord func(map[string]string) (T, bool),

	// Integrate values from b into a, updating a in-place
	integrateRecords func(a, b T),

) ([]*JobsByHost[T], error) {
	files, err := storage.EnumerateFiles(dataPath, from, to, logFilename)
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d files\n", len(files))
	}

	// Collect all the buckets, there will be many entries in this list for the same host
	bucketList := make(bucketList_t[T], 0)
	for _, filePath := range files {
		records, err := storage.ReadFreeCSV(path.Join(dataPath, filePath))
		if err != nil {
			continue
		}

		// By design, all jobs in a (host, time) bucket are consecutive in a single file.

		bucket := make(bucket_t[T], 0)
		for _, unparsed := range records {
			parsed, ok := parseLogRecord(unparsed)
			if !ok {
				continue
			}

			if len(bucket) == 0 {
				bucket = append(bucket, parsed)
				last := bucket[len(bucket)-1]
				if last.GetHost() != parsed.GetHost() || last.GetLastSeen() != parsed.GetLastSeen() {
					bucketList = append(bucketList, bucket)
					bucket = make(bucket_t[T], 0)
				}
				bucket = append(bucket, parsed)
			}
		}
		if len(bucket) > 0 {
			bucketList = append(bucketList, bucket)
		}
	}

	// Sort host list by ascending name
	sort.Slice(bucketList, func(i, j int) bool {
		return bucketList[i][0].GetHost() < bucketList[j][0].GetHost()
	})

	// Collect runs for the same host and process them
	bucketListIx := 0
	bucketListLim := len(bucketList)
	result := make([]*JobsByHost[T], 0)
	for bucketListIx < bucketListLim {
		endIx := bucketListIx + 1
		host := bucketList[bucketListIx][0].GetHost()
		for endIx < bucketListLim && host == bucketList[endIx][0].GetHost() {
			endIx++
		}
		result = append(result,
			&JobsByHost[T]{
				Host: host,
				Jobs: processLogRecordsForHost(bucketList[bucketListIx:endIx], integrateRecords),
			})
		bucketListIx = endIx
	}

	return result, nil
}

// Each entry in `buckets` is a bucket of records with the same timestamp.  All hosts in all buckets
// are the same (and can be ignored).
//
// On return, the jobs are sorted descending by lastSeen, and `expired` is set for all but the first
// job with a given ID.

func processLogRecordsForHost[T Job](
	buckets bucketList_t[T],
	integrateRecords func(a, b T),
) []T {
	const deletedMark = math.MaxUint32

	// Sort the buckets by descending time
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i][0].GetLastSeen().After(buckets[j][0].GetLastSeen())
	})

	// Merge buckets that have the same timestamp
	newBuckets := make(bucketList_t[T], 0)
	bucketIdx := 0
	for bucketIdx < len(buckets) {
		bucket := buckets[bucketIdx]
		probeIdx := bucketIdx + 1
		for probeIdx < len(buckets) && buckets[probeIdx][0].GetLastSeen() == bucket[0].GetLastSeen() {
			bucket = append(bucket, buckets[probeIdx]...)
			probeIdx++
		}
		newBuckets = append(newBuckets, bucket)
		bucketIdx = probeIdx
	}
	buckets = newBuckets

	// Now there is a bucket for each time the report was run, and the bucket list is sorted
	// descending by lastSeen timestamp.  No two buckets have the same timestamp.  Then (ignoring
	// the specific values of the timestamps) if a job ID appears in records in consecutive buckets
	// it is the same job, and those records should be merged into one; a gap in the bucket list for
	// a job ID signifies that the next time the ID is encountered it is a different job.
	//
	// This might be most easily implemented by the following per-host algorithm:
	//
	//  - start with the first bucket
	//  - pick a job, and remove it from the bucket
	//  - advance
	//  - while the next bucket has the same job
	//    - remove the job from that bucket and integrate the data
	//    - advance
	//  - push the integrated job
	//  - repeat until the first bucket is empty
	//  - discard the empty bucket and start over until the list of buckets is empty

	results := make([]T, 0)
	for bucketIdx, bucket := range buckets {
		for _, record := range bucket {
			if record.GetId() == deletedMark {
				continue
			}

		probeLoop:
			for _, probeBucket := range buckets[bucketIdx+1:] {
				any := false
				for _, probe := range probeBucket {
					if probe.GetId() == record.GetId() {
						any = true
						probe.SetId(deletedMark)

						integrateRecords(record, probe)

						// At most one hit per probeBucket
						continue probeLoop
					}
				}

				// If no hit then we're done with this job
				if !any {
					break probeLoop
				}
			}
			results = append(results, record)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].GetLastSeen().After(results[j].GetLastSeen())
	})

	for jobIdx, job := range results {
		for _, otherJob := range results[jobIdx+1:] {
			if otherJob.GetId() == job.GetId() {
				otherJob.SetExpired(true)
			}
		}
	}

	return results
}
