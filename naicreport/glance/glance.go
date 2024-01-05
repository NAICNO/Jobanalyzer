// at-a-glance - produce cluster overview data for the main dashboard.
//
// End-user options:
//
//  -data-dir directory
//  -data-path directory (obsolete name)
//    The root directory of the Sonar data store, for a particular cluster.
//
//  -sonalyze filename
//    The `sonalyze` executable.
//
//  -config-file filename
//    The machine configuration file for the cluster.
//
//  -state-dir directory
//  -state-path directory (obsolete name)
//     The directory that holds the database files for cpuhog and deadweight reports.
//
//  -tag tag-name
//    A tag for the report describing whose data it contains, typically a human-readable string.
//    The tag is included in the report and may be displayed by the dashboard code.
//
// Description:
//
// The program examines raw Sonar data as well as the state databases for cpuhog/deadweight/... and
// produces a summary report which will be displayed by the main dashboard.  The output is always
// json, printed on stdout.
//
// The data format is defined by the code below (sorry).

package glance

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"go-utils/config"
	"go-utils/options"
	"go-utils/process"
	"go-utils/sonalyze"
	"naicreport/jobstate"
	"naicreport/util"
)

// Parameters for what are "recent" events/averages and what are events "longer ago" and even "long
// ago", these could be defaults that are overrideable but are currently hardcoded.
//
// Note that for the load report, where we have special switches for the accumulation interval, only
// some values are acceptable: 30 min => --half-hourly, 60 min => --hourly, 12 hrs => --half-daily,
// 24 hrs => --daily (and there's --none but let's not go there).

const (
	RECENT_MINS = 30
	LONGER_MINS = 12 * 60
	LONG_MINS   = 24 * 60
)

const nanosPerSec = 1000000000

var nowUTC = time.Now().UTC()
var recentCutoff = nowUTC.Add(-RECENT_MINS * 60 * nanosPerSec)
var longerCutoff = nowUTC.Add(-LONGER_MINS * 60 * nanosPerSec)
var longCutoff = nowUTC.Add(-LONG_MINS * 60 * nanosPerSec)

type optionsPkg struct {
	fileOpts   *util.DataFilesOptions
	filterOpts *util.DateFilterOptions
}

func Report(progname string, args []string) error {
	sonalyzePath, configFilename, stateDir, tag, fileOpts, filterOpts, err := commandLine()
	if err != nil {
		return err
	}

	config, err := config.ReadConfig(configFilename)
	if err != nil {
		return err
	}

	// It's possible that the code in this module would be a little less verbose if there were a
	// centralized database mapping host -> glanceRecord that is just passed around for everyone to
	// use.  But it's not obvious the code would be as easy to understand.

	logOpts := &optionsPkg{fileOpts, filterOpts}
	ujbh, err := collectUsersAndJobs(sonalyzePath, logOpts)
	if err != nil {
		return err
	}
	sdbh, err := collectStatusData(sonalyzePath, logOpts)
	if err != nil {
		return err
	}
	labh, err := collectLoadAverages(sonalyzePath, configFilename, logOpts)
	if err != nil {
		return err
	}
	hogsbh, err := collectCpuhogs(path.Join(stateDir, "cpuhog-state.csv"))
	if err != nil {
		return err
	}
	deadweightbh, err := collectDeadweight(path.Join(stateDir, "deadweight-state.csv"))
	if err != nil {
		return err
	}

	recordsByHost := make(map[string]*glanceRecord)
	for _, d := range ujbh {
		r := glanceRecordForHost(recordsByHost, d.hostname, config, tag)
		r.JobsRecent = d.jobs_recent
		r.JobsLonger = d.jobs_longer
		r.UsersRecent = d.users_recent
		r.UsersLonger = d.users_longer
	}
	for _, d := range sdbh {
		r := glanceRecordForHost(recordsByHost, d.hostname, config, tag)
		cpu_status := 0
		if d.cpu_down {
			cpu_status = 1
		}
		gpu_status := 0
		if d.gpu_down {
			gpu_status = 1
		}
		r.CpuStatus = cpu_status
		r.GpuStatus = gpu_status
	}
	for _, d := range labh {
		r := glanceRecordForHost(recordsByHost, d.hostname, config, tag)
		r.CpuRecent = d.cpu_recent
		r.CpuLonger = d.cpu_longer
		r.MemRecent = d.mem_recent
		r.MemLonger = d.mem_longer
		if cfg, found := config[d.hostname]; found && cfg.GpuCards > 0 {
			r.GpuRecent = d.gpu_recent
			r.GpuLonger = d.gpu_longer
			r.GpumemRecent = d.gpumem_recent
			r.GpumemLonger = d.gpumem_longer
		}
	}
	for _, d := range hogsbh {
		r := glanceRecordForHost(recordsByHost, d.hostname, config, tag)
		r.Violators = d.count
	}
	for _, d := range deadweightbh {
		r := glanceRecordForHost(recordsByHost, d.hostname, config, tag)
		r.Deadweights = d.count
	}

	result := make(glanceRecordSlice, 0, len(recordsByHost))
	for _, v := range recordsByHost {
		result = append(result, v)
	}
	sort.Sort(result)
	bytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("While marshaling at-a-glance data: %w", err)
	}
	os.Stdout.Write(bytes)

	return nil
}

func glanceRecordForHost(
	recordsByHost map[string]*glanceRecord,
	hostname string,
	config map[string]*config.SystemConfig,
	tag string,
) *glanceRecord {
	if r, found := recordsByHost[hostname]; found {
		return r
	}
	machine := ""
	if cfg, found := config[hostname]; found {
		machine = cfg.Description
	}
	r := &glanceRecord{
		Host:    hostname,
		Tag:     tag,
		Machine: machine,
		Recent:  RECENT_MINS,
		Longer:  LONGER_MINS,
		Long:    LONG_MINS,
	}
	recordsByHost[hostname] = r
	return r
}

type glanceRecord struct {
	Host         string  `json:"hostname"`
	Tag          string  `json:"tag,omitempty"`
	Machine      string  `json:"machine,omitempty"`
	Recent       int     `json:"recent"`
	Longer       int     `json:"longer"`
	Long         int     `json:"long"`
	CpuStatus    int     `json:"cpu_status"`
	GpuStatus    int     `json:"gpu_status"`
	JobsRecent   int     `json:"jobs_recent"`
	JobsLonger   int     `json:"jobs_longer"`
	UsersRecent  int     `json:"users_recent"`
	UsersLonger  int     `json:"users_longer"`
	CpuRecent    float64 `json:"cpu_recent"`
	CpuLonger    float64 `json:"cpu_longer"`
	MemRecent    float64 `json:"mem_recent"`
	MemLonger    float64 `json:"mem_longer"`
	GpuRecent    float64 `json:"gpu_recent,omitempty"`
	GpuLonger    float64 `json:"gpu_longer,omitempty"`
	GpumemRecent float64 `json:"gpumem_recent,omitempty"`
	GpumemLonger float64 `json:"gpumem_longer,omitempty"`
	Violators    int     `json:"violators_long"`
	Deadweights  int     `json:"zombies_long"`
}

type glanceRecordSlice []*glanceRecord

func (x glanceRecordSlice) Len() int           { return len(x) }
func (x glanceRecordSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x glanceRecordSlice) Less(i, j int) bool { return x[i].Host < x[j].Host }

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Users and Jobs

type usersAndJobsByHost struct {
	hostname                   string
	jobs_recent, jobs_longer   int
	users_recent, users_longer int
}

func collectUsersAndJobs(
	sonalyzePath string,
	progOpts *optionsPkg,
) ([]*usersAndJobsByHost, error) {

	// First get the raw data about recent jobs across all hosts

	type sonalyzeJobsData struct {
		Hostname       string `json:"host"`
		User           string `json:"user"`
		Classification string `json:"classification"` // 0xSOMETHING
		EndUTC         string `json:"end"`            // YYYY-MM-DD HH:MM
	}

	var rawData []*sonalyzeJobsData
	err := runAndUnmarshal(
		sonalyzePath,
		[]string{"jobs", "-u-", "--fmt=json,host,user,classification,end"},
		progOpts,
		&rawData,
	)
	if err != nil {
		return nil, err
	}

	// Then process those data to count users and jobs
	//
	// A job is running "recently" if it's still running or its ending time is after the time that
	// starts the "recent" period.  It is running "longer" ago (but still interesting to us) if is
	// recent or its ending time is after the time that starts the "longer" period.

	type accum struct {
		users_recent, users_longer map[string]bool
		jobs_recent, jobs_longer   int
	}

	host_data := make(map[string]*accum, len(rawData))
	for _, repr := range rawData {
		var hostrec *accum
		if r, found := host_data[repr.Hostname]; found {
			hostrec = r
		} else {
			hostrec = &accum{
				users_recent: make(map[string]bool),
				users_longer: make(map[string]bool),
			}
			host_data[repr.Hostname] = hostrec
		}

		classification := sonalyze.JsonInt(repr.Classification)
		end := sonalyze.JsonDateTime(repr.EndUTC)

		// Note is_recent is included in is_longer all the way

		is_recent := (classification&sonalyze.LIVE_AT_END) != 0 || recentCutoff.Before(end)
		if is_recent {
			hostrec.jobs_recent++
			if _, found := hostrec.users_recent[repr.User]; !found {
				hostrec.users_recent[repr.User] = true
			}
		}

		is_longer := is_recent || longerCutoff.Before(end)
		if is_longer {
			hostrec.jobs_longer++
			if _, found := hostrec.users_longer[repr.User]; !found {
				hostrec.users_longer[repr.User] = true
			}
		}
	}

	// Then construct and return the result

	result := make([]*usersAndJobsByHost, 0)
	for k, v := range host_data {
		result = append(result, &usersAndJobsByHost{
			hostname:     k,
			jobs_recent:  v.jobs_recent,
			jobs_longer:  v.jobs_longer,
			users_recent: len(v.users_recent),
			users_longer: len(v.users_longer),
		})
	}

	return result, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// System status

type systemStatusByHost struct {
	hostname string
	cpu_down bool
	gpu_down bool
}

func collectStatusData(sonalyzePath string, progOpts *optionsPkg) ([]*systemStatusByHost, error) {

	// First run sonalyze to collect information about system status

	type sonalyzeUptimeData struct {
		Host   string `json:"host"`
		Device string `json:"device"`
		State  string `json:"state"`
	}

	var rawData []*sonalyzeUptimeData
	err := runAndUnmarshal(
		sonalyzePath,
		[]string{"uptime", "--interval", "4", "--fmt=json,host,device,state"},
		progOpts,
		&rawData)
	if err != nil {
		return nil, err
	}

	// Then process those data to count users and jobs.
	//
	// A system or gpu is down "now" if it is currently down so we really only care about the last
	// record per host/device combination, which will always end at the current time UTC so don't
	// worry about checking that.  Scanning from the end of the input and encountering a new host:
	//
	//  - if the state is "up"
	//    - if the device is "cpu" then they are both up (they were both down but are now up)
	//    - otherwise the device is "gpu", and they are both up (the gpu was down but is now up)
	//  - else
	//    - if the device is "gpu" then the cpu is up but the gpu is down
	//    - otherwise the device is "cpu" and they are both down

	hosts := make(map[string]*systemStatusByHost)
	for i := len(rawData) - 1; i >= 0; i-- {
		d := rawData[i]
		if _, found := hosts[d.Host]; !found {
			hosts[d.Host] = &systemStatusByHost{
				hostname: d.Host,
				cpu_down: d.State == "down" && d.Device == "cpu",
				gpu_down: d.State == "down",
			}
		}
	}

	// Construct and return result

	result := make([]*systemStatusByHost, 0, len(hosts))
	for _, v := range hosts {
		result = append(result, v)
	}

	return result, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Load averages

type loadAveragesByHost struct {
	hostname                                                                                             string
	cpu_recent, cpu_longer, mem_recent, mem_longer, gpu_recent, gpu_longer, gpumem_recent, gpumem_longer float64
}

func collectLoadAverages(
	sonalyzePath string,
	configFilename string,
	progOpts *optionsPkg,
) ([]*loadAveragesByHost, error) {

	if RECENT_MINS != 30 {
		return nil, errors.New("UNIMPLEMENTED: Only half-hourly 'recent' interval")
	}
	recentData, err := collectLoadAveragesOnce(
		sonalyzePath,
		configFilename,
		progOpts,
		"--half-hourly")
	if err != nil {
		return nil, err
	}

	if LONGER_MINS != 60*12 {
		return nil, errors.New("UNIMPLEMENTED: Only half-daily 'longer' interval")
	}
	longerData, err := collectLoadAveragesOnce(
		sonalyzePath,
		configFilename,
		progOpts,
		"--half-daily")
	if err != nil {
		return nil, err
	}

	// Join

	all := make(map[string]*loadAveragesByHost)
	for k, _ := range recentData {
		all[k] = &loadAveragesByHost{hostname: k}
	}
	for k, _ := range longerData {
		all[k] = &loadAveragesByHost{hostname: k}
	}
	for k, v := range recentData {
		obj := all[k]
		obj.cpu_recent = v.rcpu
		obj.mem_recent = v.rmem
		obj.gpu_recent = v.rgpu
		obj.gpumem_recent = v.rgpumem
	}
	for k, v := range longerData {
		obj := all[k]
		obj.cpu_longer = v.rcpu
		obj.mem_longer = v.rmem
		obj.gpu_longer = v.rgpu
		obj.gpumem_longer = v.rgpumem
	}

	result := make([]*loadAveragesByHost, 0, len(all))
	for _, v := range all {
		result = append(result, v)
	}

	return result, nil
}

type sonalyzeLoadData struct {
	host    string
	rcpu    float64
	rmem    float64
	rgpu    float64
	rgpumem float64
}

func collectLoadAveragesOnce(
	sonalyzePath, configFilename string,
	progOpts *optionsPkg,
	bucketArg string,
) (map[string]*sonalyzeLoadData, error) {

	type loadDatumJSON struct {
		Host    string `json:"host"`
		Rcpu    string `json:"rcpu"`
		Rmem    string `json:"rmem"`
		Rgpu    string `json:"rgpu"`
		Rgpumem string `json:"rgpumem"`
	}

	type systemDescJSON struct {
		Host        string `json:"hostname"`
		Description string `json:"description"`
		GpuCards    string `json:"gpucards"`
	}

	type loadDataPackageJSON struct {
		System  *systemDescJSON  `json:"system"`
		Records []*loadDatumJSON `json:"records"`
	}

	type loadDataWithSystemJSON []*loadDataPackageJSON

	var rawData loadDataWithSystemJSON
	err := runAndUnmarshal(
		sonalyzePath,
		[]string{
			"load",
			bucketArg,
			"--fmt=json,host,rcpu,rmem,rgpu,rgpumem",
			"--config-file", configFilename},
		progOpts,
		&rawData)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]*sonalyzeLoadData)
	for _, ds := range rawData {
		rs := ds.Records
		// All the hosts in ds are the same, and we only care about the last record for each host.
		if len(rs) > 0 {
			d := rs[len(rs)-1]
			hosts[d.Host] = &sonalyzeLoadData{
				host:    d.Host,
				rcpu:    sonalyze.JsonFloat64(d.Rcpu),
				rmem:    sonalyze.JsonFloat64(d.Rmem),
				rgpu:    sonalyze.JsonFloat64(d.Rgpu),
				rgpumem: sonalyze.JsonFloat64(d.Rgpumem),
			}
		}
	}

	return hosts, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Hogs and deadweights

// We want to count the hogs or deadweights that started within some time window.  This is basically
// a function of the cpuhog state file: firstViolation will have to be within the window.  So read
// the state file and then count, by host.

// Hogs are specific to the ML nodes; there is a more general notion of "policy violators" that
// is not yet well developed.

type badJobsByHost struct {
	hostname string
	count    int
}

func collectCpuhogs(stateFilename string) ([]*badJobsByHost, error) {
	return countDatabaseEntries(stateFilename)
}

func collectDeadweight(stateFilename string) ([]*badJobsByHost, error) {
	return countDatabaseEntries(stateFilename)
}

func countDatabaseEntries(stateFilename string) ([]*badJobsByHost, error) {
	db, err, _ := jobstate.ReadJobDatabaseOrEmpty(stateFilename)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]*badJobsByHost)
	for _, job := range db.Active {
		if longCutoff.Before(job.FirstViolation) {
			if h, found := hosts[job.Host]; found {
				h.count++
			} else {
				hosts[job.Host] = &badJobsByHost{hostname: job.Host, count: 1}
			}
		}
	}
	for _, job := range db.Expired {
		if longCutoff.Before(job.FirstViolation) {
			if h, found := hosts[job.Host]; found {
				h.count++
			} else {
				hosts[job.Host] = &badJobsByHost{hostname: job.Host, count: 1}
			}
		}
	}

	results := make([]*badJobsByHost, 0, len(hosts))
	for _, v := range hosts {
		results = append(results, v)
	}

	return results, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Utilities

func runAndUnmarshal(sonalyzePath string, arguments []string, progOpts *optionsPkg, rawData any) error {
	arguments = util.ForwardDateFilterOptions(arguments, progOpts.filterOpts)
	arguments = util.ForwardDataFilesOptions(arguments, "--data-path", progOpts.fileOpts)
	sonalyzeOutput, err := process.RunSubprocess(sonalyzePath, arguments)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(sonalyzeOutput), rawData)
	if err != nil {
		var extraErr error
		if sonalyzeOutput == "" {
			extraErr = errors.New("Empty output")
		}
		return errors.Join(
			fmt.Errorf("While unmarshaling output of %s %v", sonalyzePath, arguments),
			extraErr,
			err)
	}
	return nil
}

func commandLine() (
	sonalyzePath string,
	configFilename string,
	stateDir string,
	tag string,
	fileOpts *util.DataFilesOptions,
	filterOpts *util.DateFilterOptions,
	err error,
) {
	opts := flag.NewFlagSet(os.Args[0]+" at-a-glance", flag.ContinueOnError)
	fileOpts = util.AddDataFilesOptions(opts, "data-dir", "Root `directory` of data store")
	filterOpts = util.AddDateFilterOptions(opts)
	opts.StringVar(&sonalyzePath, "sonalyze", "", "Sonalyze executable `filename` (required)")
	opts.StringVar(&configFilename, "config-file", "", "Read cluster configuration from `filename` (required)")
	opts.StringVar(&stateDir, "state-dir", "", "Directory holding databases for system state (required)")
	opts.StringVar(&tag, "tag", "", "Annotate output with `cluster-name` (optional)")
	var dataPath string
	opts.StringVar(&dataPath, "data-path", "", "Obsolete name for -data-dir")
	var statePath string
	opts.StringVar(&statePath, "state-path", "", "Obsolete name for -state-dir")
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
	err1 := util.RectifyDataFilesOptions(fileOpts, opts)
	err5 := util.RectifyDateFilterOptions(filterOpts, opts)
	sonalyzePath, err2 := options.RequireCleanPath(sonalyzePath, "-sonalyze")
	configFilename, err3 := options.RequireCleanPath(configFilename, "-config-file")
	if stateDir == "" && statePath != "" {
		stateDir = statePath
	}
	stateDir, err4 := options.RequireCleanPath(stateDir, "-state-dir")
	err = errors.Join(err1, err2, err3, err4, err5)
	return
}
