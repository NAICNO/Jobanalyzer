// The ml-nodes deadweight analysis runs fairly often (see next) and examines data from a larger
// time window than its running interval, and will append information about zombies, defunct
// processes and other dead weight to a daily log.  The schedule generates a fair amount of
// redundancy under normal circumstances.
//
// The analysis MUST run often enough for a job ID on a given host never to become reused between
// two consecutive analysis runs.
//
// The present component runs occasionally and filters / resolves the redundancy and creates
// formatted reports about new problems.  For this it maintains state about what it's already seen
// and reported.
//
// Requirements:
//
//  - a job that appears in the deadweight log is dead weight and should be reported
//  - the report is (for now) just textual output to be emailed
//  - we don't want to report jobs redundantly, so there will have to be persistent state
//  - we don't want the state to grow without bound

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
	deadweightFilename = "deadweight-state.csv"
)

var verbose bool

// Se comment in mlcpuhog.go re options logic for --state-file and --now

func Deadweight(progname string, args []string) error {

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
	now, from, to time.Time,
	dataFiles []string,
	err error,
) {
	opts := flag.NewFlagSet(os.Args[0]+" deadweight", flag.ContinueOnError)
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
		stateFilename = path.Join(logOpts.DataPath, deadweightFilename)
	}

	// logOpts will establish the either-or invariant here
	if logOpts.DataFiles != nil {
		dataFiles = logOpts.DataFiles
	} else {
		var files []string
		files, err = joblog.FindJoblogFiles(logOpts.DataPath, "deadweight.csv", logOpts.From, logOpts.To)
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
