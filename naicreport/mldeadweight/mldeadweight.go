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

package mldeadweight

import (
	"encoding/json"

	"fmt"
	"os"
	"time"

	"naicreport/joblog"
	"naicreport/jobstate"
	"naicreport/storage"
	"naicreport/util"
)

const (
	deadweightFilename = "deadweight-state.csv"
)

func MlDeadweight(progname string, args []string) error {
	progOpts := util.NewStandardOptions(progname + "ml-deadweight")
	jsonOutput := progOpts.Container.Bool("json", false, "Format output as JSON")
	err := progOpts.Parse(args)
	if err != nil {
		return err
	}

	if progOpts.DataFiles != nil {
		fmt.Fprintln(os.Stderr, "The -- filename ... operation is not yet implemented for ml-deadweight")
		os.Exit(1)
	}

	db, err := jobstate.ReadJobDatabaseOrEmpty(progOpts.DataPath, deadweightFilename)
	if err != nil {
		return err
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d+%d records in database\n", len(db.Active), len(db.Expired))
	}

	logs, err := joblog.ReadJoblogFiles[*deadweightJob](
		progOpts.DataPath,
		"deadweight.csv",
		progOpts.From,
		progOpts.To,
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

	now := time.Now().UTC()

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

	purgeDate := util.MinTime(progOpts.From, progOpts.To.AddDate(0, 0, -2))
	purged := jobstate.PurgeJobsBefore(db, purgeDate)
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d purged\n", purged)
	}

	if *jsonOutput {
		bytes, err := json.Marshal(createDeadweightReport(db, logs, true))
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
	} else {
		writeDeadweightReport(createDeadweightReport(db, logs, false))
	}

	return jobstate.WriteJobDatabase(progOpts.DataPath, deadweightFilename, db)
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
	perJobReports := make([]*perJobReport, 0, len(db.Active) + len(db.Expired))
	for _, jobState := range db.Active {
		if allJobs || !jobState.IsReported {
			perJobReports = append(perJobReports, makePerJobReport(jobState))
		}
	}
	for _, jobState := range db.Expired {
		if allJobs || !jobState.IsReported {
			perJobReports = append(perJobReports, makePerJobReport(jobState))
		}
	}
	return perJobReports
}

func makePerJobReport(jobState *jobstate.JobState) *perJobReport {
	job := jobState.Aux.(*deadweightJob)
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
