// The deadweight analysis runs fairly often (see next) and examines data from a larger time window
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
	"fmt"
	"os"
	"path"
	"time"

	"naicreport/joblog"
	"naicreport/jobstate"
	"naicreport/storage"
	"naicreport/util"
)

const (
	deadweightFilename = "deadweight-state.csv"
)

// This is new-style ingest based on running sonalyze.  The output of sonalyze is incorporated
// directly into the database, which is written here.

// --state-file becomes mandatory here

// TODO: To maintain testability then we must allow '--' to take the place of running sonalyze.
// Running sonalyze basically amounts to having a single test input.  And then '--' becomes
// exclusive of -sonalyze.  What a mess.

// Extra annoying is that progOpts does not have a way of *not* having --data-path.  So that has
// to be dealt with too.

func Ingest(progname string, args []string) error {

	// TODO: options here

	db, err := readDb(stateFilename, progOpts.Verbose)
	if err != nil {
		return error
	}

	// TODO: logic
	// Run sonalyze or read the joblog from options
	// In either case incorporate new records into database

	return jobstate.WriteJobDatabase(stateFilename, db)
}

func Report(progname string, args []string) error {

	// TODO: options here

	db, err := readDb(stateFilename, progOpts.Verbose)
	if err != nil {
		return error
	}

	// We write the DB here because the IsReported flag has changed and old records have been
	// purged.

	return writeReportsAndDbAndPurge(
		db,
		stateFilename,
		progOpts.From, progOpts.To,
		*jsonOutput, *summaryOutput, progOpts.Verbose,
	)
}

// This is old-style ingest from files stored in the log file tree.  This is obsolete and will
// be removed once the new code is stable.
//
// Se comment in mlcpuhog.go re options logic for --state-file and --now

func MlDeadweight(progname string, args []string) error {

	////////////////////////////////////////////////////////////////////////////////
	//
	// Parse options and establish inputs

	progOpts := util.NewStandardOptions(progname + "ml-deadweight")
	jsonOutput := progOpts.Container.Bool("json", false, "Format output as JSON")
	summaryOutput := progOpts.Container.Bool("summary", false, "Format output for testing")
	stateFileOpt := progOpts.Container.String("state-file", "", "Name of saved-state file (optional)")
	nowOpt := progOpts.Container.String("now", "", "ISO time to use as the present time (for testing)")
	err := progOpts.Parse(args)
	if err != nil {
		return err
	}

	var stateFilename string
	switch {
	case *stateFileOpt != "":
		stateFilename = *stateFileOpt
	case progOpts.DataPath == "":
		return errors.New("If --state-file is not present then --data-path must be")
	default:
		stateFilename = path.Join(progOpts.DataPath, deadweightFilename)
	}

	// progOpts will establish the either-or invariant here
	var dataFiles []string
	if progOpts.DataFiles != nil {
		dataFiles = progOpts.DataFiles
	} else {
		files, err := joblog.FindJoblogFiles(progOpts.DataPath, "deadweight.csv", progOpts.From, progOpts.To)
		if err != nil {
			return fmt.Errorf("Could not enumerate files: %w", err)
		}
		dataFiles = files
	}

	var now time.Time
	if *nowOpt != "" {
		n, err := time.Parse(*nowOpt, util.DateTimeFormat)
		if err != nil {
			return fmt.Errorf("Argument to --now could not be parsed: %w", err)
		}
		now = n.UTC()
	} else {
		now = time.Now().UTC()
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Read database

	db, err := readDb(stateFilename, progOpts.Verbose)
	if err != nil {
		return err
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Read logs and update the database

	logs, err := joblog.ReadJoblogFiles[*deadweightJob](
		dataFiles,
		progOpts.Verbose,
		parseDeadweightRecord,
		integrateDeadweightRecords,
	)
	if err != nil {
		return err
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d hosts in log\n", len(logs))
		for _, l := range logs {
			fmt.Fprintf(os.Stderr, " %s: %d records\n", l.Host, len(l.Jobs))
		}
	}

	new_jobs := 0
	for _, hostrec := range logs {
		for _, job := range hostrec.Jobs {
			if jobstate.EnsureJob(db, job.Id, job.Host, job.start, now, job.LastSeen, job.Expired, job) {
				new_jobs++
			}
		}
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d new jobs\n", new_jobs)
	}

	////////////////////////////////////////////////////////////////////////////////
	//
	// Generate reports, update IsReported, and save the database

	return writeReportsAndDbAndPurge(
		db,
		stateFilename,
		progOpts.From, progOpts.To,
		*jsonOutput, *summaryOutput, progOpts.Verbose)
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Common code

func readDb(stateFilename string, verbose bool) (*jobstate.JobDatabase, error) {
	db, err, _ := jobstate.ReadJobDatabaseOrEmpty(stateFilename)
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d+%d records in database\n", len(db.Active), len(db.Expired))
	}
	return db, nil
}

func writeReportsAndDbAndPurge(
	db *jobstate.JobDatabase,
	stateFilename string,
	from, to time.Time,
	jsonOutput, summaryOutput, verbose bool,
) error {
	purgeDate := util.MinTime(from, to.AddDate(0, 0, -2))
	purged := jobstate.PurgeJobsBefore(db, purgeDate)
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d purged\n", purged)
	}

	switch {
	case jsonOutput:
		bytes, err := json.Marshal(createDeadweightReport(db, true))
		if err != nil {
			return fmt.Errorf("While marshaling deadweight data: %w", err)
		}
		fmt.Println(string(bytes))
	case summaryOutput:
		report := createDeadweightReport(db, false)
		for _, r := range report {
			r.jobState.IsReported = true
			fmt.Printf("%s,%d,%s\n", r.User, r.Id, r.Host)
		}
	default:
		writeDeadweightReport(createDeadweightReport(db, false))
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
		StartedOnOrBefore: jobState.StartedOnOrBefore.Format(util.DateTimeFormat),
		FirstViolation:    jobState.FirstViolation.Format(util.DateTimeFormat),
		LastSeen:          jobState.LastSeen.Format(util.DateTimeFormat),
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
	record.firstSeen = util.MinTime(record.firstSeen, other.firstSeen)
	record.LastSeen = util.MaxTime(record.LastSeen, other.LastSeen)
	record.start = util.MinTime(record.start, other.start)
	record.end = util.MaxTime(record.end, other.end)
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
