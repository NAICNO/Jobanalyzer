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
	"time"

	"naicreport/joblog"
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

	logs, err := joblog.ReadJoblogFiles[*cpuhogJob](
		progOpts.DataPath,
		"cpuhog.csv",
		progOpts.From,
		progOpts.To,
		progOpts.Verbose,
		parseCpuhogRecord,
		integrateCpuhogRecords,
	)
	if err != nil {
		return err
	}
	if progOpts.Verbose {
		fmt.Fprintf(os.Stderr, "%d hosts in log\n", len(logs))
		for _, l := range logs {
			fmt.Fprintf(os.Stderr, " %s: %d records\n", l.Host, len(l.Jobs));
		}
	}

	now := time.Now().UTC()

	new_jobs := 0
	for _, hostrec := range logs {
		for _, job := range hostrec.Jobs {
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
	logs []*joblog.JobsByHost[*cpuhogJob],
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

// cpuhogJob implements joblog.Job

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

func (s *cpuhogJob) Id() uint32 {
	return s.id
}
func (s *cpuhogJob) SetId(id uint32) {
	s.id = id
}
func (s *cpuhogJob) Host() string {
	return s.host
}
func (s *cpuhogJob) LastSeen() time.Time {
	return s.lastSeen
}
func (s *cpuhogJob) IsExpired() bool {
	return s.expired
}
func (s *cpuhogJob) SetExpired(flag bool) {
	s.expired = flag
}

func integrateCpuhogRecords(record, probe *cpuhogJob) {
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
	}, true
}
