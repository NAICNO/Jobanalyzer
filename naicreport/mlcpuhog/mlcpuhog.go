// The ml-nodes cpuhog analysis runs fairly often (see next) and examines data from a larger time
// window than its running interval, and will append information about CPU hogs to a daily log.  The
// schedule generates a fair amount of redundancy under normal circumstances.
//
// The analysis MUST run often enough for a job ID on a given host never to become reused between
// two consecutive analysis runs.
//
// The present component runs occasionally and filters / resolves the redundancy and creates
// formatted reports about new violations.  For this it maintains state about what it's already seen
// and reported.
//
// For now this code is specific to the ML nodes, hence the "ml" in all the names.
//
// Requirements:
//
//  - a job that appears in the cpuhog log is a cpu hog and should be reported
//  - the report is (for now) some textual output to be emailed
//  - we don't want to report jobs redundantly, so there will have to be persistent state
//  - we don't want the state to grow without bound

// For textual reports, there is persistent state so that reports are not sent redundantly.  This
// state maps a job to a flag that indicates whether a report has been sent or not.
//
// A job is identified by a quadruple (host, id, expired, lastSeen): If expired==false, then the job
// key is (host, id), otherwise the key is (host, id, lastSeen).  The reason for this is that
// non-expired jobs have non-constant lastSeen values, and there is only ever one non-expired job
// for the same (host, id) pair.
//
// The astute reader will have realized that "expired" and "lastSeen" depend on the time window of
// records that this program examines.  Therefore, the persistent state is dependent on that time
// window.  For sane results, the `--to` switch should not be used and the `--from` switch should
// use a constant, relative time, eg `--from=30d` to always be examining the last month.
//
// Records older than the `--from` date will be purged from the persistent state.

package mlcpuhog

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"time"

	"naicreport/jobstate"
	"naicreport/storage"
	"naicreport/util"
)

const (
	cpuhogFilename = "cpuhog-state.csv"
)

func MlCpuhog(progname string, args []string) error {
	progOpts := util.NewStandardOptions(progname + "ml-cpuhog")
	jsonOutput := progOpts.Container.Bool("json", false, "Format output as JSON")
	err := progOpts.Parse(args)
	if err != nil {
		return err
	}

	hogState, err := jobstate.ReadJobDatabaseOrEmpty(progOpts.DataPath, cpuhogFilename)
	if err != nil {
		return err
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d+%d records in database\n", len(hogState.Active), len(hogState.Expired))
	}

	logs, err := readCpuhogLogFiles(progOpts.DataPath, progOpts.From, progOpts.To, progOpts.Verbose)
	if err != nil {
		return err
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d hosts in log\n", len(logs))
		for _, l := range logs {
			fmt.Fprintf(os.Stderr, " %s: %d records\n", l.host, len(l.jobs));
		}
	}

	now := time.Now().UTC()

	new_jobs := 0
	for _, hostrec := range logs {
		for _, job := range hostrec.jobs {
			if jobstate.EnsureJob(hogState, job.id, job.host, job.start, now, job.lastSeen, job.expired, job) {
				new_jobs++
			}
		}
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d new jobs\n", new_jobs)
	}

	purgeDate := util.MinTime(progOpts.From, progOpts.To.AddDate(0, 0, -2))
	purged := jobstate.PurgeJobsBefore(hogState, purgeDate)
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d jobs purged\n", purged)
	}

	events := createCpuhogReport(hogState, logs)
	if *jsonOutput {
		bytes, err := json.Marshal(events)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
	} else {
		writeCpuhogReport(events)
	}

	return jobstate.WriteJobDatabase(progOpts.DataPath, cpuhogFilename, hogState)
}

type perEvent struct {
	Host              string `json:"hostname"`
	Id                uint32 `json:"id"`
	User              string `json:"user"`
	Cmd               string `json:"cmd"`
	StartedOnOrBefore string `json:"started-on-or-before"`
	FirstViolation    string `json:"first-violation"`
	CpuPeak           uint32 `json:"cpu-peak"`
	RCpuAvg           uint32 `json:"rcpu-avg"`
	RCpuPeak          uint32 `json:"rcpu-peak"`
	RMemAvg           uint32 `json:"rmem-avg"`
	RMemPeak          uint32 `json:"rmem-peak"`
}

func createCpuhogReport(
	hogState *jobstate.JobDatabase,
	logs []*cpuhogJobsByHost,
) []*perEvent {

	events := make([]*perEvent, 0)
	for _, jobState := range hogState.Active {
		if !jobState.IsReported {
			jobState.IsReported = true
			events = append(events, makeEvent(jobState))
		}
	}
	for _, jobState := range hogState.Expired {
		if !jobState.IsReported {
			jobState.IsReported = true
			events = append(events, makeEvent(jobState))
		}
	}
	return events
}

func makeEvent(jobState *jobstate.JobState) *perEvent {
	job := jobState.Aux.(*cpuhogJob)
	return &perEvent{
		Host:              jobState.Host,
		Id:                jobState.Id,
		User:              job.user,
		Cmd:               job.cmd,
		StartedOnOrBefore: jobState.StartedOnOrBefore.Format(util.DateTimeFormat),
		FirstViolation:    jobState.FirstViolation.Format(util.DateTimeFormat),
		CpuPeak:           uint32(job.cpuPeak / 100),
		RCpuAvg:           uint32(job.rcpuAvg),
		RCpuPeak:          uint32(job.rcpuPeak),
		RMemAvg:           uint32(job.rmemAvg),
		RMemPeak:          uint32(job.rmemPeak),
	}
}

func writeCpuhogReport(events []*perEvent) {
	reports := make([]*util.JobReport, 0)
	for _, e := range events {
		report := fmt.Sprintf(
			`New CPU hog detected (uses a lot of CPU and no GPU) on host "%s":
  Job#: %d
  User: %s
  Command: %s
  Started on or before: %s
  Violation first detected: %s
  Observed data:
    CPU peak = %d cores
    CPU utilization avg/peak = %d%%, %d%%
    Memory utilization avg/peak = %d%%, %d%%

`,
			e.Host,
			e.Id,
			e.User,
			e.Cmd,
			e.StartedOnOrBefore,
			e.FirstViolation,
			e.CpuPeak,
			e.RCpuAvg,
			e.RCpuPeak,
			e.RMemAvg,
			e.RMemPeak)
		reports = append(reports, &util.JobReport{Id: e.Id, Host: e.Host, Report: report})
	}

	util.SortReports(reports)
	for _, r := range reports {
		fmt.Print(r.Report)
	}
}

// The output of this algorithm is a list of hosts, and for each host a list of individual jobs for
// that host (the former sorted ascending by name, the latter descending by lastSeen timestamp).
// Notably job IDs may be reused in this list: fact of life.  But they are still different jobs.
//
// We are going to assume that the log records can be ingested and partitioned by host and then
// bucketed by timestamp and the buckets sorted, and if a job ID is present in two consecutive
// buckets in the list of buckets then it's the same job, and otherwise those are two different jobs
// with a reused ID.  The assumption is valid because there is a requirement stated above, and in
// the shell scripts and cron jobs, that the analysis producing our input runs often enough for the
// assumption to hold.
//
// The cpuhogJob has a field 'expired' which will be false for the first record that has each job ID
// and true for all subsequent records with that ID.

type cpuhogJob struct {
	id        uint32    // job id
	host      string    // a single host name, since ml nodes
	user      string    // user's login name
	cmd       string    // ???
	firstSeen time.Time // timestamp of record in which job is first seen
	lastSeen  time.Time // ditto the record in which the job is last seen
	start     time.Time // the start field of the first record for the job
	end       time.Time // the end field of the last record for the job
	expired   bool      // false iff this is the latest for this id on this host
	cpuPeak   float64   // this and the following are the Max across all
	gpuPeak   float64   //   records seen for the job, this is necessary
	rcpuAvg   float64   //     as sonalyze will have a limited window in which
	rcpuPeak  float64   //       to gather statistics and its view will change
	rmemAvg   float64   //         over time
	rmemPeak  float64   //
}

type cpuhogJobsByHost struct {
	host string
	jobs []*cpuhogJob
}

type bucket_t []*cpuhogJob
type bucketList_t []bucket_t

func readCpuhogLogFiles(
	dataPath string,
	from, to time.Time,
	verbose bool,
) ([]*cpuhogJobsByHost, error) {
	files, err := storage.EnumerateFiles(dataPath, from, to, "cpuhog.csv")
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d files\n", len(files))
	}

	// Collect all the buckets, there will be many entries in this list for the same host
	bucketList := make(bucketList_t, 0)
	for _, filePath := range files {
		records, err := storage.ReadFreeCSV(path.Join(dataPath, filePath))
		if err != nil {
			continue
		}

		// By design, all jobs in a (host, time) bucket are consecutive in a single file.

		bucket := make(bucket_t, 0)
		for _, unparsed := range records {
			parsed := parseRecord(unparsed)
			if parsed == nil {
				continue
			}

			if len(bucket) == 0 {
				bucket = append(bucket, parsed)
				last := bucket[len(bucket)-1]
				if last.host != parsed.host || last.lastSeen != parsed.lastSeen {
					bucketList = append(bucketList, bucket)
					bucket = make(bucket_t, 0)
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
		return bucketList[i][0].host < bucketList[j][0].host
	})

	// Collect runs for the same host and process them
	bucketListIx := 0
	bucketListLim := len(bucketList)
	result := make([]*cpuhogJobsByHost, 0)
	for bucketListIx < bucketListLim {
		endIx := bucketListIx + 1
		host := bucketList[bucketListIx][0].host
		for endIx < bucketListLim && host == bucketList[endIx][0].host {
			endIx++
		}
		result = append(result,
			&cpuhogJobsByHost{
				host: host,
				jobs: processRecordsForHost(bucketList[bucketListIx:endIx]),
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

func processRecordsForHost(buckets bucketList_t) []*cpuhogJob {
	const deletedRecordMark = math.MaxUint32

	// Sort the buckets by descending time
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i][0].lastSeen.After(buckets[j][0].lastSeen)
	})

	// Merge buckets that have the same timestamp
	newBuckets := make(bucketList_t, 0)
	bucketIdx := 0
	for bucketIdx < len(buckets) {
		bucket := buckets[bucketIdx]
		probeIdx := bucketIdx + 1
		for probeIdx < len(buckets) && buckets[probeIdx][0].lastSeen == bucket[0].lastSeen {
			bucket = append(bucket, buckets[probeIdx]...)
			probeIdx++
		}
		newBuckets = append(newBuckets, bucket)
		bucketIdx = probeIdx
	}
	buckets = newBuckets

	/*
		for _, b := range buckets {
			fmt.Println("==========")
			for _, r := range b {
				fmt.Printf("%v %v %v\n", r.host, r.id, r.lastSeen)
			}
		}
	*/

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

	results := make([]*cpuhogJob, 0)
	for bucketIdx, bucket := range buckets {
		for _, record := range bucket {
			if record.id == deletedRecordMark {
				continue
			}

		probeLoop:
			for _, probeBucket := range buckets[bucketIdx+1:] {
				any := false
				for _, probe := range probeBucket {
					if probe.id == record.id {
						any = true
						probe.id = deletedRecordMark

						// Integrate probe into record
						record.firstSeen = util.MinTime(record.firstSeen, probe.firstSeen)
						record.lastSeen = util.MaxTime(record.lastSeen, probe.lastSeen)
						record.start = util.MinTime(record.start, probe.start)
						record.end = util.MaxTime(record.end, probe.end)
						record.cpuPeak = math.Max(record.cpuPeak, probe.cpuPeak)
						record.gpuPeak = math.Max(record.gpuPeak, probe.gpuPeak)
						record.rcpuAvg = math.Max(record.rcpuAvg, probe.rcpuAvg)
						record.rcpuPeak = math.Max(record.rcpuPeak, probe.rcpuPeak)
						record.rmemAvg = math.Max(record.rmemAvg, probe.rmemAvg)
						record.rmemPeak = math.Max(record.rmemPeak, probe.rmemPeak)

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
		return results[i].lastSeen.After(results[j].lastSeen)
	})

	for jobIdx, job := range results {
		for _, otherJob := range results[jobIdx+1:] {
			if otherJob.id == job.id {
				otherJob.expired = true
			}
		}
	}

	return results
}

// This always sets firstSeen = lastSeen = timestamp, and expired = false.

func parseRecord(r map[string]string) *cpuhogJob {
	success := true

	tag := storage.GetString(r, "tag", &success)
	success = success && tag == "cpuhog"
	timestamp := storage.GetDateTime(r, "now", &success)
	id := storage.GetJobMark(r, "jobm", &success)
	user := storage.GetString(r, "user", &success)
	host := storage.GetString(r, "host", &success)
	cmd := storage.GetString(r, "cmd", &success)
	cpuPeak := storage.GetFloat64(r, "cpu-peak", &success)
	gpuPeak := storage.GetFloat64(r, "gpu-peak", &success)
	rcpuAvg := storage.GetFloat64(r, "rcpu-avg", &success)
	rcpuPeak := storage.GetFloat64(r, "rcpu-peak", &success)
	rmemAvg := storage.GetFloat64(r, "rmem-avg", &success)
	rmemPeak := storage.GetFloat64(r, "rmem-peak", &success)
	start := storage.GetDateTime(r, "start", &success)
	end := storage.GetDateTime(r, "end", &success)

	if !success {
		return nil
	}

	firstSeen := timestamp
	lastSeen := timestamp

	return &cpuhogJob{
		id,
		host,
		user,
		cmd,
		firstSeen,
		lastSeen,
		start,
		end,
		false,
		cpuPeak,
		gpuPeak,
		rcpuAvg,
		rcpuPeak,
		rmemAvg,
		rmemPeak,
	}
}
