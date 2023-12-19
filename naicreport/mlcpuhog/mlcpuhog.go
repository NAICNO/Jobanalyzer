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
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"go-utils/sonarlog"
	sonartime "go-utils/time"
	"naicreport/joblog"
	"naicreport/jobstate"
	"naicreport/storage"
	"naicreport/util"
)

const (
	cpuhogStateFilename = "cpuhog-state.csv"
	cpuhogDataFilename  = "cpuhog.csv"
)

var verbose bool

// Options logic here and for deadweight:
//
// - If `--state-file` is present then that names the state file; otherwise `--data-path` must be present
//   and we use `${dataPath}/cpuhog-state.csv` as the state file
//
// - If `-- filename ...` are present then those are used for input; otherwise `--data-path` must be present
//   and we enumerate cpuhog.csv files within the dataPath directory
//
// - The `--now` option is strictly for testing, everyone else should ignore it.  It parses a time on
//   the format "YYYY-MM-DD hh:mm" (yes, with the space).

func MlCpuhog(progname string, args []string) error {

	jsonOutput, summaryOutput, stateFilename, now, from, to, dataFiles, err := commandLine()
	if err != nil {
		return err
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Read inputs

	db, err, _ := jobstate.ReadJobDatabaseOrEmpty(stateFilename)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d+%d records in database\n", len(db.Active), len(db.Expired))
	}

	logs, err := joblog.ReadJoblogFiles[*cpuhogJob](
		dataFiles,
		verbose,
		parseCpuhogRecord,
		integrateCpuhogRecords,
	)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d hosts in log\n", len(logs))
		for _, l := range logs {
			fmt.Fprintf(os.Stderr, " %s: %d records\n", l.Host, len(l.Jobs))
		}
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Create the new state

	new_jobs := 0
	for _, hostrec := range logs {
		for _, job := range hostrec.Jobs {
			if jobstate.EnsureJob(db, job.Id, job.Host, job.start, now, job.LastSeen, job.Expired, job) {
				new_jobs++
			}
		}
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d new jobs\n", new_jobs)
	}

	purgeDate := sonartime.MinTime(from, to.AddDate(0, 0, -2))
	purged := jobstate.PurgeJobsBefore(db, purgeDate)
	if verbose {
		fmt.Fprintf(os.Stderr, "%d jobs purged\n", purged)
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Write outputs

	switch {
	case jsonOutput:
		bytes, err := json.Marshal(createCpuhogReport(db, logs, true))
		if err != nil {
			return fmt.Errorf("While marshaling cpuhog data: %w", err)
		}
		fmt.Println(string(bytes))
	case summaryOutput:
		report := createCpuhogReport(db, logs, false)
		for _, r := range report {
			r.jobState.IsReported = true
			fmt.Printf("%s,%d,%s,%d\n", r.User, r.Id, r.Host, r.CpuPeak)
		}
	default:
		writeCpuhogReport(createCpuhogReport(db, logs, false))
	}

	return jobstate.WriteJobDatabase(stateFilename, db)
}

type perJobReport struct {
	Host              string `json:"hostname"`
	Id                uint32 `json:"id"`
	User              string `json:"user"`
	Cmd               string `json:"cmd"`
	StartedOnOrBefore string `json:"started-on-or-before"`
	FirstViolation    string `json:"first-violation"`
	LastSeen          string `json:"last-seen"`
	CpuPeak           uint32 `json:"cpu-peak"`
	RCpuAvg           uint32 `json:"rcpu-avg"`
	RCpuPeak          uint32 `json:"rcpu-peak"`
	RMemAvg           uint32 `json:"rmem-avg"`
	RMemPeak          uint32 `json:"rmem-peak"`
	jobState          *jobstate.JobState
}

func createCpuhogReport(
	db *jobstate.JobDatabase,
	logs []*joblog.JobsByHost[*cpuhogJob],
	allJobs bool,
) []*perJobReport {
	perJobReports := make([]*perJobReport, 0, len(db.Active)+len(db.Expired))
	for _, jobState := range db.Active {
		if allJobs || !jobState.IsReported {
			r := makePerJobReport(jobState)
			if r != nil {
				perJobReports = append(perJobReports, r)
			}
		}
	}
	for _, jobState := range db.Expired {
		if allJobs || !jobState.IsReported {
			r := makePerJobReport(jobState)
			if r != nil {
				perJobReports = append(perJobReports, r)
			}
		}
	}
	return perJobReports
}

// This returns nil if the job was in such a state that no report could be created.  This can
// legitimately happen when the database has more information than the logs.  See issue #220.
func makePerJobReport(jobState *jobstate.JobState) *perJobReport {
	job, found := jobState.Aux.(*cpuhogJob)
	if !found {
		return nil
	}
	return &perJobReport{
		Host:              jobState.Host,
		Id:                jobState.Id,
		User:              job.user,
		Cmd:               job.cmd,
		StartedOnOrBefore: jobState.StartedOnOrBefore.Format(sonarlog.DateTimeFormat),
		FirstViolation:    jobState.FirstViolation.Format(sonarlog.DateTimeFormat),
		LastSeen:          jobState.LastSeen.Format(sonarlog.DateTimeFormat),
		CpuPeak:           uint32(job.cpuPeak / 100),
		RCpuAvg:           uint32(job.rcpuAvg),
		RCpuPeak:          uint32(job.rcpuPeak),
		RMemAvg:           uint32(job.rmemAvg),
		RMemPeak:          uint32(job.rmemPeak),
		jobState:          jobState,
	}
}

func writeCpuhogReport(perJobReports []*perJobReport) {
	reports := make([]*util.JobReport, 0, len(perJobReports))
	for _, r := range perJobReports {
		r.jobState.IsReported = true
		report := fmt.Sprintf(
			`New CPU hog detected (uses a lot of CPU and no GPU) on host "%s":
  Job#: %d
  User: %s
  Command: %s
  Started on or before: %s
  Violation first detected: %s
  Last Seen: %s
  Observed data:
    CPU peak = %d cores
    CPU utilization avg/peak = %d%%, %d%%
    Memory utilization avg/peak = %d%%, %d%%

`,
			r.Host,
			r.Id,
			r.User,
			r.Cmd,
			r.StartedOnOrBefore,
			r.FirstViolation,
			r.LastSeen,
			r.CpuPeak,
			r.RCpuAvg,
			r.RCpuPeak,
			r.RMemAvg,
			r.RMemPeak)
		reports = append(reports, &util.JobReport{Id: r.Id, Host: r.Host, Report: report})
	}

	util.SortReports(reports)
	for _, r := range reports {
		fmt.Print(r.Report)
	}
}

// cpuhogJob implements joblog.Job

type cpuhogJob struct {
	joblog.GenericJob
	user      string    // user's login name
	cmd       string    // ???
	firstSeen time.Time // timestamp of record in which job is first seen
	start     time.Time // the start field of the first record for the job
	end       time.Time // the end field of the last record for the job
	cpuPeak   float64   // this and the following are the Max across all
	gpuPeak   float64   //   records seen for the job, this is necessary
	rcpuAvg   float64   //     as sonalyze will have a limited window in which
	rcpuPeak  float64   //       to gather statistics and its view will change
	rmemAvg   float64   //         over time
	rmemPeak  float64   //
}

func integrateCpuhogRecords(record, probe *cpuhogJob) {
	record.LastSeen = sonartime.MaxTime(record.LastSeen, probe.LastSeen)
	record.firstSeen = sonartime.MinTime(record.firstSeen, probe.firstSeen)
	record.start = sonartime.MinTime(record.start, probe.start)
	record.end = sonartime.MaxTime(record.end, probe.end)
	record.cpuPeak = math.Max(record.cpuPeak, probe.cpuPeak)
	record.gpuPeak = math.Max(record.gpuPeak, probe.gpuPeak)
	record.rcpuAvg = math.Max(record.rcpuAvg, probe.rcpuAvg)
	record.rcpuPeak = math.Max(record.rcpuPeak, probe.rcpuPeak)
	record.rmemAvg = math.Max(record.rmemAvg, probe.rmemAvg)
	record.rmemPeak = math.Max(record.rmemPeak, probe.rmemPeak)
}

func parseCpuhogRecord(r map[string]string) (*cpuhogJob, bool) {
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
		return nil, false
	}

	return &cpuhogJob{
		GenericJob: joblog.GenericJob{
			Id:       id,
			Host:     host,
			LastSeen: timestamp,
			Expired:  false,
		},
		user:      user,
		cmd:       cmd,
		firstSeen: timestamp,
		start:     start,
		end:       end,
		cpuPeak:   cpuPeak,
		gpuPeak:   gpuPeak,
		rcpuAvg:   rcpuAvg,
		rcpuPeak:  rcpuPeak,
		rmemAvg:   rmemAvg,
		rmemPeak:  rmemPeak,
	}, true
}

// This is virtually identical to the function in deadweight.go
func commandLine() (
	jsonOutput, summaryOutput bool,
	stateFilename string,
	now, from, to time.Time,
	dataFiles []string,
	err error,
) {
	opts := flag.NewFlagSet(os.Args[0]+" ml-cpuhog", flag.ContinueOnError)
	logOpts := util.AddSonarLogOptions(opts)
	opts.BoolVar(&jsonOutput, "json", false, "Format output as JSON")
	opts.BoolVar(&summaryOutput, "summary", false, "Format output for testing")
	opts.StringVar(&stateFilename, "state-file", "", "Store computation state in `filename` (optional)")
	var nowOpt string
	opts.StringVar(&nowOpt, "now", "", "ISO `timestamp` to use as the present time (for testing)")
	opts.BoolVar(&verbose, "v", false, "Verbose (debugging) output")
	err = opts.Parse(os.Args[2:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}
	err = util.RectifySonarLogOptions(logOpts, opts)
	if err != nil {
		return
	}
	if stateFilename == "" {
		if logOpts.DataPath == "" {
			err = errors.New("If --state-file is not present then --data-path must be")
			return
		}
		stateFilename = path.Join(logOpts.DataPath, cpuhogStateFilename)
	}

	// logOpts will establish the either-or invariant here
	if logOpts.DataFiles != nil {
		dataFiles = logOpts.DataFiles
	} else {
		var files []string
		files, err = joblog.FindJoblogFiles(logOpts.DataPath, cpuhogDataFilename, logOpts.From, logOpts.To)
		if err != nil {
			err = fmt.Errorf("Could not enumerate files: %w", err)
			return
		}
		dataFiles = files
	}

	if nowOpt != "" {
		var n time.Time
		n, err = time.Parse(nowOpt, sonarlog.DateTimeFormat)
		if err != nil {
			err = fmt.Errorf("Argument to --now could not be parsed: %w", err)
			return
		}
		now = n.UTC()
	} else {
		now = time.Now().UTC()
	}
	from = logOpts.From
	to = logOpts.To
	return
}
