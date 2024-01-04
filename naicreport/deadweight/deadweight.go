// deadweight - consolidate periodic deadweight reports into a persistent database.
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
//     Normally, the database is <state-dir>/deadweight-state.csv.
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
//     Name the files holding the periodic reports.  Normally those data files are the
//     deadweight.csv log files found in the state directory in the <year>/<month>/<day>
//     subdirectories appropriate for the selected date range.
//
// Description:
//
// The ml-nodes deadweight analysis runs fairly often and examines data from a larger time window
// than its running interval, and will append information about zombies, defunct processes and other
// dead weight to a daily log.  The schedule generates a fair amount of redundancy under normal
// circumstances.
//
// The analysis MUST run often enough for a job ID on a given host never to become reused between
// two consecutive analysis runs.
//
// The present component runs occasionally and filters / resolves the redundancy and creates
// formatted reports about new problems.  For this it maintains state about what it's already seen
// and reported.
//
// -------------------------------------------------------------------------------------------------
//
// Implementation notes.
//
// Requirements:
//
//  - a job that appears in the deadweight log is dead weight and should be reported
//  - the report is (for now) just textual output to be emailed
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

package deadweight

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	deadweightStateFilename = "deadweight-state.csv"
	deadweightDataFilename  = "deadweight.csv"
)

var verbose bool

func Deadweight(progname string, args []string) error {

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

	var dataFiles []string
	// fileOpts will establish the either-or invariant here
	if fileOpts.Files != nil {
		dataFiles = fileOpts.Files
	} else {
		dataFiles, err = joblog.FindJoblogFiles(
			fileOpts.Path,
			deadweightDataFilename,
			filterOpts.From,
			filterOpts.To)
		if err != nil {
			return err
		}
	}

	logs, err := joblog.ReadJoblogFiles[*deadweightJob](
		dataFiles,
		verbose,
		parseDeadweightRecord,
		integrateDeadweightRecords,
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
	purgeDate := sonartime.MinTime(from, to.AddDate(0, 0, -2))
	purged := jobstate.PurgeJobsBefore(db, purgeDate)
	if verbose {
		fmt.Fprintf(os.Stderr, "%d purged\n", purged)
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Write outputs

	switch {
	case jsonOutput:
		bytes, err := json.Marshal(createDeadweightReport(db, logs, true))
		if err != nil {
			return fmt.Errorf("While marshaling deadweight data: %w", err)
		}
		fmt.Println(string(bytes))
	case summaryOutput:
		report := createDeadweightReport(db, logs, false)
		for _, r := range report {
			r.jobState.IsReported = true
			fmt.Printf("%s,%d,%s\n", r.User, r.Id, r.Host)
		}
	default:
		writeDeadweightReport(createDeadweightReport(db, logs, false))
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
	jobState          *jobstate.JobState
}

func createDeadweightReport(
	db *jobstate.JobDatabase,
	logs []*joblog.JobsByHost[*deadweightJob],
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
	job, found := jobState.Aux.(*deadweightJob)
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
		jobState:          jobState,
	}
}

func writeDeadweightReport(perJobReports []*perJobReport) {
	reports := make([]*util.JobReport, 0, len(perJobReports))
	for _, r := range perJobReports {
		r.jobState.IsReported = true
		report := fmt.Sprintf(
			`New pointless job detected (zombie, defunct, or hung) on host "%s":
  Job#: %d
  User: %s
  Command: %s
  Started on or before: %s
  Violation first detected: %s
  Last seen: %s
`,
			r.Host,
			r.Id,
			r.User,
			r.Cmd,
			r.StartedOnOrBefore,
			r.FirstViolation,
			r.LastSeen)
		reports = append(reports, &util.JobReport{Id: r.Id, Host: r.Host, Report: report})
	}

	util.SortReports(reports)
	for _, r := range reports {
		fmt.Print(r.Report)
	}
}

// deadweightJob implements joblog.Job

type deadweightJob struct {
	joblog.GenericJob
	user      string
	cmd       string
	firstSeen time.Time
	start     time.Time
	end       time.Time
}

func integrateDeadweightRecords(record, other *deadweightJob) {
	record.firstSeen = sonartime.MinTime(record.firstSeen, other.firstSeen)
	record.LastSeen = sonartime.MaxTime(record.LastSeen, other.LastSeen)
	record.start = sonartime.MinTime(record.start, other.start)
	record.end = sonartime.MaxTime(record.end, other.end)
}

func parseDeadweightRecord(r map[string]string) (*deadweightJob, bool) {
	success := true
	tag := storage.GetString(r, "tag", &success)
	// Old files used "bughunt" for the tag
	success = success && (tag == "deadweight" || tag == "bughunt")
	timestamp := storage.GetDateTime(r, "now", &success)
	id := storage.GetJobMark(r, "jobm", &success)
	user := storage.GetString(r, "user", &success)
	cmd := storage.GetString(r, "cmd", &success)
	host := storage.GetString(r, "host", &success)
	start := storage.GetDateTime(r, "start", &success)
	end := storage.GetDateTime(r, "end", &success)

	if !success {
		return nil, false
	}

	return &deadweightJob{
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
	}, true
}

// This is virtually identical to the function in mlcpuhog.go
func commandLine() (
	jsonOutput, summaryOutput bool,
	stateFilename string,
	now time.Time,
	fileOpts *util.DataFilesOptions,
	filterOpts *util.DateFilterOptions,
	err error,
) {
	opts := flag.NewFlagSet(os.Args[0]+" deadweight", flag.ContinueOnError)
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
	if fileOpts.Path == "" && fileOpts.Files == nil && dataPath != "" {
		fileOpts.Path = dataPath
	}
	err = util.RectifyDateFilterOptions(filterOpts, opts)
	if err != nil {
		return
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
		stateFilename = path.Join(fileOpts.Path, deadweightStateFilename)
	}
	if nowOpt != "" {
		var n time.Time
		n, err = time.Parse(nowOpt, sonarlog.DateTimeFormat)
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
