// The ml-nodes deadweight analysis runs every 12h (at least - currently it runs every 2 hours, for
// testing purposes), examining data from the previous 24h, and will append information about
// zombies, defunct processes and other dead weight to a daily log.  This generates a fair amount of
// redundancy under normal circumstances.
//
// The present component runs occasionally (tbd) and filters / resolves the redundancy and creates
// formatted reports about new problems.  For this it maintains state about what it's already seen
// and reported.
//
// Requirements:
//
//  - a job that appears in the deadweight log is dead weight and should be reported
//  - the report is (for now) some textual output of the form shown below
//  - we don't want to report jobs redundantly, so there will have to be persistent state
//  - we don't want the state to grow without bound
//
// Report format:
//
//  (tbd)

package mldeadweight

import (
	"encoding/json"

	"fmt"
	"os"
	"path"
	"time"

	"naicreport/jobstate"
	"naicreport/storage"
	"naicreport/util"
)

const (
	deadweightFilename = "deadweight-state.csv"
)

type deadweightJob struct {
	id        uint32
	host      string
	user      string
	cmd       string
	firstSeen time.Time
	lastSeen  time.Time
	start     time.Time
	end       time.Time
}

func MlDeadweight(progname string, args []string) error {
	progOpts := util.NewStandardOptions(progname + "ml-deadweight")
	jsonOutput := progOpts.Container.Bool("json", false, "Format output as JSON")
	err := progOpts.Parse(args)
	if err != nil {
		return err
	}

	state, err := jobstate.ReadJobStateOrEmpty(progOpts.DataPath, deadweightFilename)
	if err != nil {
		return err
	}

	logs, err := readDeadweightLogFiles(progOpts.DataPath, progOpts.From, progOpts.To, progOpts.Verbose)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	candidates := 0
	for _, job := range logs {
		if jobstate.EnsureJob(state, job.id, job.host, job.start, now, job.lastSeen) {
			candidates++
		}
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d candidates\n", candidates)
	}

	purgeDate := util.MinTime(progOpts.From, progOpts.To.AddDate(0, 0, -2))
	purged := jobstate.PurgeJobsBefore(state, purgeDate)
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d purged\n", purged)
	}

	events := createDeadweightReport(state, logs)
	if *jsonOutput {
		bytes, err := json.Marshal(events)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
	} else {
		writeDeadweightReport(events)
	}

	return jobstate.WriteJobState(progOpts.DataPath, deadweightFilename, state)
}

type perEvent struct {
	Host              string `json:"hostname"`
	Id                uint32 `json:"id"`
	User              string `json:"user"`
	Cmd               string `json:"cmd"`
	StartedOnOrBefore string `json:"started-on-or-before"`
	FirstViolation    string `json:"first-violation"`
	LastSeen          string `json:"last-seen"`
}

func createDeadweightReport(
     state map[jobstate.JobKey]*jobstate.JobState,
     logs map[jobstate.JobKey]*deadweightJob,
) []*perEvent {
	events := make([]*perEvent, 0)
	for k, j := range state {
		if !j.IsReported {
			j.IsReported = true
			loggedJob, _ := logs[k]
			events = append(events,
				&perEvent{
					Host:              j.Host,
					Id:                j.Id,
					User:              loggedJob.user,
					Cmd:               loggedJob.cmd,
					StartedOnOrBefore: j.StartedOnOrBefore.Format(util.DateTimeFormat),
					FirstViolation:    j.FirstViolation.Format(util.DateTimeFormat),
					LastSeen:          j.LastSeen.Format(util.DateTimeFormat),
				})
		}
	}
	return events
}

func writeDeadweightReport(events []*perEvent) {
	reports := make([]*util.JobReport, 0)
	for _, e := range events {
		report := fmt.Sprintf(
			`New pointless job detected (zombie, defunct, or hung) on host "%s":
  Job#: %d
  User: %s
  Command: %s
  Started on or before: %s
  Violation first detected: %s
  Last seen: %s
`,
			e.Host,
			e.Id,
			e.User,
			e.Cmd,
			e.StartedOnOrBefore,
			e.FirstViolation,
			e.LastSeen)
		reports = append(reports, &util.JobReport{Id: e.Id, Host: e.Host, Report: report})
	}

	util.SortReports(reports)
	for _, r := range reports {
		fmt.Print(r.Report)
	}
}

func readDeadweightLogFiles(
     dataPath string,
     from, to time.Time,
     verbose bool,
) (map[jobstate.JobKey]*deadweightJob, error) {
	files, err := storage.EnumerateFiles(dataPath, from, to, "deadweight.csv")
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "%d files\n", len(files))
	}

	jobs := make(map[jobstate.JobKey]*deadweightJob)
	for _, filePath := range files {
		records, err := storage.ReadFreeCSV(path.Join(dataPath, filePath))
		if err != nil {
			continue
		}

		for _, r := range records {
			success := true
			tag := storage.GetString(r, "tag", &success)
			// Old files used "bughunt" for the tag
			success = success && (tag == "deadweight" || tag == "bughunt")
			now := storage.GetDateTime(r, "now", &success)
			id := storage.GetJobMark(r, "jobm", &success)
			user := storage.GetString(r, "user", &success)
			host := storage.GetString(r, "host", &success)
			cmd := storage.GetString(r, "cmd", &success)
			start := storage.GetDateTime(r, "start", &success)
			end := storage.GetDateTime(r, "end", &success)
			// TODO: duration

			if !success {
				continue
			}

			key := jobstate.JobKey{Id: id, Host: host}
			if r, present := jobs[key]; present {
				// id, user, and host are fixed - host b/c this is the view of a job on the ml nodes
				// TODO: cmd can change b/c of sonalyze's view on the job.
				r.firstSeen = util.MinTime(r.firstSeen, now)
				r.lastSeen = util.MaxTime(r.lastSeen, now)
				r.start = util.MinTime(r.start, start)
				r.end = util.MaxTime(r.end, end)
				// TODO: Duration
			} else {
				firstSeen := now
				lastSeen := now
				jobs[key] = &deadweightJob{
					id,
					host,
					user,
					cmd,
					firstSeen,
					lastSeen,
					start,
					end,
					// TODO: duration
				}
			}

		}
	}

	return jobs, nil
}
