// The ml-nodes cpuhog analysis runs every 12h (at least - currently it runs every 2 hours, for
// testing purposes), examining data from the previous 24h, and will append information about CPU
// hogs to a daily log.  This generates a fair amount of redundancy under normal circumstances.
//
// The present component runs occasionally (tbd) and filters / resolves the redundancy and creates
// formatted reports about new violations.  For this it maintains state about what it's already seen
// and reported.
//
// For now this code is specific to the ML nodes, hence the "ml" in all the names.
//
// Requirements:
//
//  - a job that appears in the cpuhog log is a cpu hog and should be reported
//  - the report is (for now) some textual output of the form shown below
//  - we don't want to report jobs redundantly, so there will have to be persistent state
//  - we don't want the state to grow without bound
//
// Report format (when not JSON):
//
//     New CPU hog detected (uses a lot of CPU and no GPU) on host "XX":
//       Job#: n
//       User: username
//       Command: command name
//       Violation first detected: <date>  // this is the timestamp of the earliest record
//       Started on or before: <date>      // this is the start-time in the earliest record
//       Observed data:
//          CPU peak = n cores
//          CPU utilization avg/peak = n%, m%
//          Memory utilization avg/peak = n%, m%

package mlcpuhog

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"naicreport/jobstate"
	"naicreport/storage"
	"naicreport/util"
)

const (
	cpuhogFilename = "cpuhog-state.csv"
)

// The cpuhogState represents the view of a job across all the records read from the logs.  Here, too,
// (job#, host) identifies the job uniquely.

type cpuhogState struct {
	id        uint32        // synthesized job id
	host      string        // a single host name, since ml nodes
	user      string        // user's login name
	cmd       string        // ???
	duration  time.Duration // ???
	firstSeen time.Time     // timestamp of record in which job is first seen
	lastSeen  time.Time     // ditto the record in which the job is last seen
	start     time.Time     // the start field of the first record for the job
	end       time.Time     // the end field of the last record for the job
	cpuPeak   float64       // this and the following are the Max across all
	gpuPeak   float64       //   records seen for the job, this is necessary
	rcpuAvg   float64       //     as sonalyze will have a limited window in which
	rcpuPeak  float64       //       to gather statistics and its view will change
	rmemAvg   float64       //         over time
	rmemPeak  float64       //
}

func MlCpuhog(progname string, args []string) error {
	progOpts := util.NewStandardOptions(progname + "ml-cpuhog")
	jsonOutput := progOpts.Container.Bool("json", false, "Format output as JSON")
	err := progOpts.Parse(args)
	if err != nil {
		return err
	}

	hogState, err := jobstate.ReadJobStateOrEmpty(progOpts.DataPath, cpuhogFilename)
	if err != nil {
		return err
	}

	logs, err := readLogFiles(progOpts.DataPath, progOpts.From, progOpts.To)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	candidates := 0
	for _, job := range logs {
		if jobstate.EnsureJob(hogState, job.id, job.host, job.start, now, job.lastSeen) {
			candidates++
		}
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d candidates\n", candidates)
	}

	purged := jobstate.PurgeDeadJobs(hogState, progOpts.To)
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d purged\n", purged)
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

	return jobstate.WriteJobState(progOpts.DataPath, cpuhogFilename, hogState)
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
	hogState map[jobstate.JobKey]*jobstate.JobState,
	logs map[jobstate.JobKey]*cpuhogState) []*perEvent {

	events := make([]*perEvent, 0)
	for k, jobState := range hogState {
		if !jobState.IsReported {
			jobState.IsReported = true
			job, _ := logs[k]
			events = append(events,
				&perEvent{
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
				})
		}
	}
	return events
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

func readLogFiles(dataPath string, from, to time.Time) (map[jobstate.JobKey]*cpuhogState, error) {
	files, err := storage.EnumerateFiles(dataPath, from, to, "cpuhog.csv")
	if err != nil {
		return nil, err
	}

	jobs := make(map[jobstate.JobKey]*cpuhogState)
	for _, filePath := range files {
		records, err := storage.ReadFreeCSV(path.Join(dataPath, filePath))
		if err != nil {
			continue
		}

		for _, r := range records {
			success := true

			tag := storage.GetString(r, "tag", &success)
			success = success && tag == "cpuhog"
			now := storage.GetDateTime(r, "now", &success)
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
				continue
			}

			key := jobstate.JobKey{id, host}
			if r, present := jobs[key]; present {
				// id, user, and host are fixed - host b/c this is the view of a job on the ml nodes
				// FIXME: cmd can change b/c of sonalyze's view on the job.
				r.firstSeen = util.MinTime(r.firstSeen, now)
				r.lastSeen = util.MaxTime(r.lastSeen, now)
				r.start = util.MinTime(r.start, start)
				r.end = util.MaxTime(r.end, end)
				// FIXME: duration can change
				r.cpuPeak = math.Max(r.cpuPeak, cpuPeak)
				r.gpuPeak = math.Max(r.gpuPeak, gpuPeak)
				r.rcpuAvg = math.Max(r.rcpuAvg, rcpuAvg)
				r.rcpuPeak = math.Max(r.rcpuPeak, rcpuPeak)
				r.rmemAvg = math.Max(r.rmemAvg, rmemAvg)
				r.rmemPeak = math.Max(r.rmemPeak, rmemPeak)
			} else {
				firstSeen := now
				lastSeen := now
				duration := time.Duration(0) // FIXME
				jobs[key] = &cpuhogState{
					id,
					host,
					user,
					cmd,
					duration,
					firstSeen,
					lastSeen,
					start,
					end,
					cpuPeak,
					gpuPeak,
					rcpuAvg,
					rcpuPeak,
					rmemAvg,
					rmemPeak,
				}
			}
		}
	}

	return jobs, nil
}
