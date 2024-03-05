// ml-cpuhog - consolidate periodic cpuhog reports into a persistent database.
//
// End-user options:
//
//  -state-dir directory
//  -data-path directory (obsolete name)
//     The directory that holds the database file and is the root of the directory tree
//     containing the periodic reports.
//
//  -from timestamp
//     The start of the window of time we're interested in, default 1d (1 day ago).  Normally, the
//     `-from` switch should use a constant, relative time, eg `-from 30d` to always be examining
//     the last month.
//
//     Note, records older than the `-from` date will be purged from the persistent state.
//
//  -json
//     Produce json output, not the default formatted-text output.
//
// Debugging / development options:
//
//  -v
//     Print diagnostic information
//
//  -summary
//     Print some summary information
//
//  -state-file filename
//     Name the file holding the database, overrides the standard computation of this file name.
//     Normally, the database is <state-dir>/cpuhog-state.csv.
//
//  -to timestamp
//     The end of the window of time, default now.  Normally, the `-to` switch should not be used,
//     as the resulting purging of the database can be surprising.
//
//  -now time
//     Define `now` so that relative times like `1d` are relative to a fixed known time.  The time
//     must have the precise format "YYYY-MM-DD hh:mm" (with the embedded space)
//
//  -- filename ... (at end of options list only)
//     Name the files holding the periodic reports.  Normally those data files are the cpuhog.csv
//     log files found in the state directory in the <year>/<month>/<day> subdirectories appropriate
//     for the selected date range.
//
// Description:
//
// The ML nodes cpuhog analysis runs fairly often and examines log data from a larger time window
// than its running interval.  It will append information about CPU hogs to a daily log.  The
// schedule generates a fair amount of redundancy under normal circumstances.
//
// The cpuhog analysis MUST run often enough for a job ID on a given host never to become reused
// between two consecutive analysis runs.
//
// The ml-cpuhog component runs occasionally and filters / resolves the redundancy and creates
// formatted reports about new violations.  For this it maintains state about what it's already seen
// and reported.
//
// For now this code is specific to the ML nodes, hence the "ml" in all the names.
//
// -------------------------------------------------------------------------------------------------
//
// Implementation notes.
//
// Program requirements:
//
//  - a job that appears in the cpuhog log is a cpu hog and should be reported
//  - the report is (for now) some textual output to be emailed
//  - we don't want to report jobs redundantly, so there will have to be persistent state
//  - we don't want the state to grow without bound
//
// Persistent state:
//
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
// window.  For sane results, the `-to` switch should not be used and the `-from` switch should
// use a constant, relative time, eg `-from 30d` to always be examining the last month.

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

	"go-utils/freecsv"
	gut "go-utils/time"
	"naicreport/joblog"
	"naicreport/jobstate"
	"naicreport/util"
)

const (
	cpuhogStateFilename = "cpuhog-state.csv"
	cpuhogDataFilename  = "cpuhog.csv"
)

var verbose bool

func MlCpuhog(progname string, args []string) error {

	jsonOutput, summaryOutput, stateFilename, now, fileOpts, filterOpts, err := commandLine()
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

	// fileOpts will establish the either-or invariant here
	var dataFiles []string
	if fileOpts.Files != nil {
		dataFiles = fileOpts.Files
	} else {
		dataFiles, err = joblog.FindJoblogFiles(
			fileOpts.Path,
			cpuhogDataFilename,
			filterOpts.From,
			filterOpts.To)
		if err != nil {
			return err
		}
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

	from := filterOpts.From
	to := filterOpts.To
	purgeDate := gut.MinTime(from, to.AddDate(0, 0, -2))
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
		StartedOnOrBefore: jobState.StartedOnOrBefore.Format(gut.CommonDateTimeFormat),
		FirstViolation:    jobState.FirstViolation.Format(gut.CommonDateTimeFormat),
		LastSeen:          jobState.LastSeen.Format(gut.CommonDateTimeFormat),
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
	record.LastSeen = gut.MaxTime(record.LastSeen, probe.LastSeen)
	record.firstSeen = gut.MinTime(record.firstSeen, probe.firstSeen)
	record.start = gut.MinTime(record.start, probe.start)
	record.end = gut.MaxTime(record.end, probe.end)
	record.cpuPeak = math.Max(record.cpuPeak, probe.cpuPeak)
	record.gpuPeak = math.Max(record.gpuPeak, probe.gpuPeak)
	record.rcpuAvg = math.Max(record.rcpuAvg, probe.rcpuAvg)
	record.rcpuPeak = math.Max(record.rcpuPeak, probe.rcpuPeak)
	record.rmemAvg = math.Max(record.rmemAvg, probe.rmemAvg)
	record.rmemPeak = math.Max(record.rmemPeak, probe.rmemPeak)
}

func parseCpuhogRecord(r map[string]string) (*cpuhogJob, bool) {
	success := true

	tag := freecsv.GetString(r, "tag", &success)
	success = success && tag == "cpuhog"
	timestamp := freecsv.GetCommonDateTime(r, "now", &success)
	id := freecsv.GetJobMark(r, "jobm", &success)
	user := freecsv.GetString(r, "user", &success)
	host := freecsv.GetString(r, "host", &success)
	cmd := freecsv.GetString(r, "cmd", &success)
	cpuPeak := freecsv.GetFloat64(r, "cpu-peak", &success)
	gpuPeak := freecsv.GetFloat64(r, "gpu-peak", &success)
	rcpuAvg := freecsv.GetFloat64(r, "rcpu-avg", &success)
	rcpuPeak := freecsv.GetFloat64(r, "rcpu-peak", &success)
	rmemAvg := freecsv.GetFloat64(r, "rmem-avg", &success)
	rmemPeak := freecsv.GetFloat64(r, "rmem-peak", &success)
	start := freecsv.GetCommonDateTime(r, "start", &success)
	end := freecsv.GetCommonDateTime(r, "end", &success)

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
	now time.Time,
	fileOpts *util.DataFilesOptions,
	filterOpts *util.DateFilterOptions,
	err error,
) {
	opts := flag.NewFlagSet(os.Args[0]+" ml-cpuhog", flag.ContinueOnError)
	fileOpts = util.AddDataFilesOptions(opts, "state-dir", "Root `directory` of state data store")
	filterOpts = util.AddDateFilterOptions(opts)
	opts.BoolVar(&jsonOutput, "json", false, "Format output as JSON")
	opts.BoolVar(&summaryOutput, "summary", false, "Format output for testing")
	opts.StringVar(&stateFilename, "state-file", "", "Store computation state in `filename` (optional)")
	var nowOpt string
	opts.StringVar(&nowOpt, "now", "", "ISO `timestamp` to use as the present time (for testing)")
	opts.BoolVar(&verbose, "v", false, "Verbose (debugging) output")
	var dataPath string
	opts.StringVar(&dataPath, "data-path", "", "Obsolete name for -state-dir")
	err = opts.Parse(os.Args[2:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}
	err = util.RectifyDateFilterOptions(filterOpts, opts)
	if err != nil {
		return
	}
	if fileOpts.Path == "" && fileOpts.Files == nil && dataPath != "" {
		fileOpts.Path = dataPath
	}
	err = util.RectifyDataFilesOptions(fileOpts, opts)
	if err != nil {
		return
	}
	if stateFilename == "" {
		if fileOpts.Path == "" {
			err = errors.New("If -state-file is not present then -state-dir must be")
			return
		}
		stateFilename = path.Join(fileOpts.Path, cpuhogStateFilename)
	}
	if nowOpt != "" {
		var n time.Time
		n, err = time.Parse(nowOpt, gut.CommonDateTimeFormat)
		if err != nil {
			err = fmt.Errorf("Argument to -now could not be parsed: %w", err)
			return
		}
		now = n.UTC()
	} else {
		now = time.Now().UTC()
	}
	return
}
