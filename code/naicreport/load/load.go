// load - Generate plots of machine load and uptime from Sonar data.
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
//  -report-dir directory
//  -output-path directory (obsolete name)
//    The directory in which to store the generated per-host reports.
//
//  -tag tag-name
//    A tag for the report describing what it is about, sometimes a time range.  The report will be
//    stored in a file typically called <hostname>-<tag>.json.  A typical tag is `hourly` or
//    `weekly`, tagging the report as being for hourly or weekly data.  The `naicreport hostnames`
//    command makes use of the tagging scheme when computing the set of hostnames.
//
//  -from timestamp
//    The start of the window of time we're interested in, default 1d (1 day ago).
//
//  -to timestamp
//    The end of the window of time we're interested in, default now.
//
//  -group hostname-patterns
//    This option selects a number of hostnames and merges their data to provide a summary report
//    for the cluster or subcluster.  For example, on the ML cluster we use `-group 'ml[1-3,6-9]'`
//    to produce an aggregate report for the machines with NVIDIA accelerators.
//
//    The hostname pattern is meant to match only the first element of a qualified host name, and
//    its syntax is a string prefix optionally followed by a set of numbers and numeric ranges, as
//    above, or equivalently 'ml[1,2,3,6,7,8,9]'.  Multiple patterns can be joined with commas, eg
//    'ml1,ml2,ml3,ml[6-8],ml9'.
//
//  -daily
//    Bucket data by day.  One of -daily, -hourly, or -none is required.
//
//  -hourly
//    Bucket data by hour
//
//  -none
//    Do not bucket data.
//
//  -with-downtime interval
//    Include downtime data for hosts and GPUs in the report, the `interval` is the sampling
//    interval in minutes for the cluster.
//
// Debugging / development options:
//
//  -v
//    Print various (verbose) debugging output
//
// Description:
//
// The `load` command runs sonalyze on Sonar data and produces a JSON object per host that
// represents load data, bucketed and grouped as specified by options.  The reports are written to
// the report directory in files named <hostname>-<tag>.json (if there is a tag, which is
// recommended).
//
// The data format is documented by the code below (sorry) but is engineered for compactness as an
// object carrying some metadata and a number of parallel arrays of numbers to describe the data
// values.  The data are consumed by the JavaScript code that presents the plot, see files in
// ../../dashboard.

package load

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"time"

	"go-utils/gpuset"
	"go-utils/hostglob"
	"go-utils/minmax"
	"go-utils/options"
	"go-utils/process"
	"go-utils/sonalyze"
	gut "go-utils/time"
	"naicreport/util"
)

// Command-line options
var (
	sonalyzePath   string
	configFilename string
	reportDir      string
	tag            string
	group          string
	hourly         bool
	daily          bool
	none           bool
	withDowntime   uint
	filterOpts     *util.DateFilterOptions
	filesOpts      *util.DataFilesOptions
	verbose        bool
)

func Load(progname string, args []string) error {
	err := commandLine()
	if err != nil {
		return err
	}

	// Assemble sonalyze arguments

	loadArguments := loadInitialArgs()
	downtimeArguments := downtimeInitialArgs()

	loadArguments = util.ForwardDateFilterOptions(loadArguments, filterOpts)
	downtimeArguments = util.ForwardDateFilterOptions(downtimeArguments, filterOpts)

	// This handling of bucketing isn't completely clean but it's good enough for not-insane users.
	// We can use flag.Visit() to do a better job, if we want.

	var bucketing string
	if none {
		loadArguments = append(loadArguments, "--none")
		bucketing = "none"
	} else if daily {
		loadArguments = append(loadArguments, "--daily")
		bucketing = "daily"
	} else if hourly {
		loadArguments = append(loadArguments, "--hourly")
		bucketing = "hourly"
	} else {
		return errors.New("One of --daily, --hourly, or --none is required")
	}

	if group != "" {
		if bucketing == "none" {
			return errors.New("Cannot --group together with --none")
		}
		if withDowntime > 0 {
			return errors.New("Cannot --group together with --with-downtime")
		}
		loadArguments = append(loadArguments, "--group")

		patterns, err := hostglob.SplitMultiPattern(group)
		if err != nil {
			return err
		}
		for _, p := range patterns {
			loadArguments = append(loadArguments, "--host", p)
		}
	}

	// For -- this must come last, so do standard log options last always

	loadArguments = util.ForwardDataFilesOptions(loadArguments, "--data-path", filesOpts)
	downtimeArguments = util.ForwardDataFilesOptions(downtimeArguments, "--data-path", filesOpts)

	// Obtain all the data

	if verbose {
		fmt.Printf("Sonalyze load arguments\n%v", loadArguments)
	}
	loadOutput, loadErrOutput, err := process.RunSubprocess("sonalyze", sonalyzePath, loadArguments)
	if err != nil {
		if loadErrOutput != "" {
			return errors.Join(err, fmt.Errorf("With stderr:\n%s", loadErrOutput))
		}
		return err
	}
	if loadErrOutput != "" {
		fmt.Fprintln(os.Stderr, loadErrOutput)
	}

	loadData, err := parseLoadOutputBySystem(loadOutput)
	if err != nil {
		return err
	}

	var downtimeData []*downtimeDataByHost
	if withDowntime > 0 {
		if verbose {
			fmt.Printf("Sonalyze downtime arguments\n%v", downtimeArguments)
		}
		downtimeOutput, downtimeErrOutput, err := process.RunSubprocess("sonalyze", sonalyzePath, downtimeArguments)
		if err != nil {
			if downtimeErrOutput != "" {
				return errors.Join(err, fmt.Errorf("With stderr:\n%s", downtimeErrOutput))
			}
			return err
		}
		if downtimeErrOutput != "" {
			fmt.Fprintln(os.Stderr, downtimeOutput)
		}
		downtimeData, err = parseDowntimeOutput(downtimeOutput)
		if err != nil {
			return err
		}
	}

	return writePlots(bucketing, group != "", loadData, downtimeData)
}

func writePlots(
	bucketing string,
	grouping bool,
	loadData []*loadDataBySystem,
	downtimeData []*downtimeDataByHost,
) error {
	type perSystem struct {
		Host        string `json:"hostname"`
		Description string `json:"description"`
	}

	type perHost struct {
		Date      string     `json:"date"`
		Host      string     `json:"hostname"`
		Tag       string     `json:"tag"`
		Bucketing string     `json:"bucketing"`
		Labels    []string   `json:"labels"` // formatted timestamps, for now
		Rcpu      []float64  `json:"rcpu"`
		Rmem      []float64  `json:"rmem"`
		Rres      []float64  `json:"rres"`
		Rgpu      []float64  `json:"rgpu,omitempty"`
		Rgpumem   []float64  `json:"rgpumem,omitempty"`
		DownHost  []int      `json:"downhost,omitempty"`
		DownGpu   []int      `json:"downgpu,omitempty"`
		System    *perSystem `json:"system,omitempty"`
	}

	if grouping && len(loadData) != 1 {
		return fmt.Errorf("Expected exactly one datum for grouped run, tag=%s", tag)
	}

	// The config for a host may be missing, but this should still work.
	//
	// downtimeData may be nil, in which case it should be ignored, but if not nil it must have been
	//  quantized already

	// Use the same timestamp for all records
	now := time.Now().Format(gut.CommonDateTimeFormat)

	for _, hd := range loadData {
		var basename string
		if grouping {
			basename = tag + ".json"
		} else if tag == "" {
			basename = hd.system.host + ".json"
		} else {
			basename = hd.system.host + "-" + tag + ".json"
		}
		filename := path.Join(reportDir, basename)
		output_file, err := os.CreateTemp(path.Dir(filename), "naicreport-load")
		if err != nil {
			return err
		}
		// The tempname is the full path, it's set to "" below once the file is renamed.
		tempname := output_file.Name()
		defer (func() {
			if tempname != "" {
				os.Remove(tempname)
			}
		})()

		hasGpu := hd.system != nil && hd.system.gpuCards > 0
		labels := make([]string, 0)
		rcpuData := make([]float64, 0)
		rmemData := make([]float64, 0)
		rresData := make([]float64, 0)
		var rgpuData, rgpumemData []float64
		if hasGpu {
			rgpuData = make([]float64, 0)
			rgpumemData = make([]float64, 0)
		}
		for _, d := range hd.data {
			labels = append(labels, d.datetime.Format(gut.CommonDateTimeFormat))
			rcpuData = append(rcpuData, d.rcpu)
			rmemData = append(rmemData, d.rmem)
			rresData = append(rresData, d.rres)
			if hasGpu {
				// Throw away GPU data if found to be invalid
				if math.IsNaN(d.rgpu) || math.IsNaN(d.rgpumem) {
					hasGpu = false
					rgpuData = nil
					rgpumemData = nil
				} else {
					rgpuData = append(rgpuData, d.rgpu)
					rgpumemData = append(rgpumemData, d.rgpumem)
				}
			}
		}
		downHost, downGpu := generateDowntimeData(hd, downtimeData, hasGpu)
		var system *perSystem
		if hd.system != nil {
			system = &perSystem{
				Host:        hd.system.host,
				Description: hd.system.description,
			}
		}
		data := perHost{
			Date:      now,
			Host:      hd.system.host,
			Tag:       tag,
			Bucketing: bucketing,
			Labels:    labels,
			Rcpu:      rcpuData,
			Rgpu:      rgpuData,
			Rmem:      rmemData,
			Rres:      rresData,
			Rgpumem:   rgpumemData,
			DownHost:  downHost,
			DownGpu:   downGpu,
			System:    system,
		}
		bytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("While marshaling perHost data: %v %w", data, err)
		}
		output_file.Write(bytes)

		output_file.Close()
		os.Rename(tempname, filename)
		tempname = "" // Signal that no cleanup is required
	}

	return nil
}

func generateDowntimeData(ld *loadDataBySystem, dd []*downtimeDataByHost, hasGpu bool) (downHost []int, downGpu []int) {
	if dd == nil {
		return
	}

	ddix := sort.Search(len(dd), func(i int) bool {
		return dd[i].host >= ld.system.host
	})
	if ddix == len(dd) {
		/* This is possible because we run sonalyze uptime with --only-down:
		   it's possible for there to be no downtime in the time window. */
		return
	}
	if dd[ddix].host > ld.system.host {
		return
	}
	downtimeData := dd[ddix].data

	loadData := ld.data
	downHost = make([]int, len(loadData))
	if hasGpu {
		downGpu = make([]int, len(loadData))
	}

	for _, ddval := range downtimeData {
		loc := sort.Search(len(loadData), func(i int) bool {
			return loadData[i].datetime.After(ddval.start)
		})
		isDevice := ddval.device == "host"
		for ix := minmax.MaxInt(loc-1, 0); ix < len(loadData) && loadData[ix].datetime.Before(ddval.end); ix++ {
			if isDevice {
				downHost[ix] = 1
			} else if hasGpu {
				downGpu[ix] = 1
			}
		}
	}
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Handle `sonalyze uptime`.

// TODO: In sonalyze, "start_utc" or something like that should mean a timestamp, not a formatted
// date.  Then we can just slurp that in here and avoid overhead and complexity.

func downtimeInitialArgs() []string {
	return []string{
		"uptime",
		"--config-file", configFilename,
		"--interval", fmt.Sprint(withDowntime),
		"--only-down",
		"--fmt=json,device,host,start,end",
	}
}

type downtimeDatum struct {
	device string
	host   string
	start  time.Time
	end    time.Time
}

type downtimeDataByHost struct {
	host string
	data []*downtimeDatum
}

// In the returned data, the point data in the inner list are sorted by ascending time, and the
// outer list is sorted by ascending host name.

func parseDowntimeOutput(output string) ([]*downtimeDataByHost, error) {
	type downtimeRepresentation struct {
		Device string `json:"device"`
		Host   string `json:"host"`
		Start  string `json:"start"`
		End    string `json:"end"`
	}

	var rawData []*downtimeRepresentation
	err := json.Unmarshal([]byte(output), &rawData)
	if err != nil {
		return nil, fmt.Errorf("While unmarshaling downtime data: %w", err)
	}

	// The output from `sonalyze downtime` is sorted first by host, then by ascending start time.

	var outerData = make([]*downtimeDataByHost, 0)
	var innerData = make([]*downtimeDatum, 0)
	for _, repr := range rawData {
		// Convert some values
		start, startErr := time.Parse(gut.CommonDateTimeFormat, repr.Start)
		end, endErr := time.Parse(gut.CommonDateTimeFormat, repr.End)
		if startErr != nil || endErr != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Skipping bad timestamp(s) %s %s\n", repr.Start, repr.End)
			}
			continue
		}

		// Distribute into host-anchored lists
		if len(innerData) > 0 && innerData[0].host != repr.Host {
			outerData = append(outerData, &downtimeDataByHost{
				host: innerData[0].host,
				data: innerData,
			})
			innerData = make([]*downtimeDatum, 0)
		}
		innerData = append(innerData, &downtimeDatum{
			device: repr.Device,
			host:   repr.Host,
			start:  start,
			end:    end,
		})
	}
	if len(innerData) > 0 {
		outerData = append(outerData, &downtimeDataByHost{
			host: innerData[0].host,
			data: innerData,
		})
	}

	return outerData, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Handle `sonalyze load`.

// Note item higher up about "start_utc" - here we want "datetime_utc" or something like that.

type loadDatum struct {
	datetime time.Time
	cpu      float64
	mem      float64
	gpu      float64
	gpumem   float64
	gpus     gpuset.GpuSet
	rcpu     float64
	rmem     float64
	rres     float64
	rgpu     float64
	rgpumem  float64
	host     string // redundant but maybe useful
}

type systemDesc struct {
	host        string
	description string
	gpuCards    int
}

type loadDataBySystem struct {
	system *systemDesc
	data   []*loadDatum
}

// In the returned data, the point data in the inner list are sorted by ascending time, and the
// outer list is sorted by ascending host name.

func loadInitialArgs() []string {
	return []string{
		"load",
		"--config-file", configFilename,
		"--fmt=json,datetime,cpu,mem,gpu,gpumem,rcpu,rmem,rres,rgpu,rgpumem,gpus,host",
	}
}

func parseLoadOutputBySystem(output string) ([]*loadDataBySystem, error) {
	// It's a recorded bug that all JSON primitive data are represented as strings, necessitating
	// extra code here.

	type loadDatumJSON struct {
		Datetime string `json:"datetime"`
		Cpu      string `json:"cpu"`
		Mem      string `json:"mem"`
		Gpu      string `json:"gpu"`
		Gpumem   string `json:"gpumem"`
		Gpus     string `json:"gpus"`
		Rcpu     string `json:"rcpu"`
		Rmem     string `json:"rmem"`
		Rres     string `json:"rres"`
		Rgpu     string `json:"rgpu"`
		Rgpumem  string `json:"rgpumem"`
		Host     string `json:"host"`
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
	err := json.Unmarshal([]byte(output), &rawData)
	if err != nil {
		return nil, fmt.Errorf("While unmarshaling data from `sonalyze load`: %w", err)
	}

	allData := make([]*loadDataBySystem, 0)
	for _, bySystem := range rawData {
		data := make([]*loadDatum, 0)
		for _, r := range bySystem.Records {
			newDatum := &loadDatum{
				datetime: sonalyze.JsonDateTime(r.Datetime),
				cpu:      sonalyze.JsonFloat64(r.Cpu),
				mem:      sonalyze.JsonFloat64(r.Mem),
				gpu:      sonalyze.JsonFloat64(r.Gpu),
				gpumem:   sonalyze.JsonFloat64(r.Gpumem),
				gpus:     sonalyze.JsonGpuSet(r.Gpus),
				rcpu:     sonalyze.JsonFloat64(r.Rcpu),
				rmem:     sonalyze.JsonFloat64(r.Rmem),
				rres:     sonalyze.JsonFloat64(r.Rres),
				rgpu:     sonalyze.JsonFloat64(r.Rgpu),
				rgpumem:  sonalyze.JsonFloat64(r.Rgpumem),
				host:     bySystem.System.Host,
			}
			data = append(data, newDatum)
		}
		allData = append(allData, &loadDataBySystem{
			system: &systemDesc{
				host:        bySystem.System.Host,
				description: bySystem.System.Description,
				gpuCards:    sonalyze.JsonInt(bySystem.System.GpuCards),
			},
			data: data,
		})
	}

	return allData, nil
}

func commandLine() error {
	opts := flag.NewFlagSet(os.Args[0]+" load", flag.ContinueOnError)
	filesOpts = util.AddDataFilesOptions(opts, "data-dir", "Root `directory` of data store")
	opts.StringVar(&sonalyzePath, "sonalyze", "", "Sonalyze executable `filename` (required)")
	opts.StringVar(&configFilename, "config-file", "", "Read cluster configuration from `filename` (required)")
	filterOpts = util.AddDateFilterOptions(opts)
	const defaultReportDir = "."
	opts.StringVar(&reportDir, "report-dir", defaultReportDir, "Store reports in `directory`")
	opts.StringVar(&tag, "tag", "", "Tag report file names with `tag-name` (optional)")
	opts.StringVar(&group, "group", "", "Group these `host name patterns` (comma-separated) (requires bucketing, too)")
	opts.BoolVar(&hourly, "hourly", true, "Bucket data hourly")
	opts.BoolVar(&daily, "daily", false, "Bucket data daily")
	opts.BoolVar(&none, "none", false, "Do not bucket data")
	opts.UintVar(&withDowntime, "with-downtime", 0, "Include downtime data for this sampling `interval` (minutes)")
	opts.BoolVar(&verbose, "v", false, "Verbose (debugging) output")
	var dataPath string
	opts.StringVar(&dataPath, "data-path", "", "Obsolete name for -data-dir")
	var outputPath string
	opts.StringVar(&outputPath, "output-path", "", "Obsolete name for -report-dir")
	err := opts.Parse(os.Args[2:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return err
	}
	err5 := util.RectifyDateFilterOptions(filterOpts, opts)
	if filesOpts.Path == "" && filesOpts.Files == nil && dataPath != "" {
		filesOpts.Path = dataPath
	}
	var err2, err3, err4 error
	err1 := util.RectifyDataFilesOptions(filesOpts, opts)
	sonalyzePath, err2 = options.RequireCleanPath(sonalyzePath, "-sonalyze")
	configFilename, err3 = options.RequireCleanPath(configFilename, "-config-file")
	if reportDir == defaultReportDir && outputPath != "" {
		reportDir = outputPath
	}
	reportDir, err4 = options.RequireCleanPath(reportDir, "-report-dir")
	return errors.Join(err1, err2, err3, err4, err5)
}
