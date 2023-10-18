// Generate data for plotting the running load of the ML systems.  The data are taken from the live
// sonar logs, by means of sonalyze.

package mlwebload

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"naicreport/storage"
	"naicreport/util"
)

func MlWebload(progname string, args []string) error {
	// Parse and sanitize options

	progOpts := util.NewStandardOptions(progname + " ml-webload")
	sonalyzePathPtr := progOpts.Container.String("sonalyze", "", "Path to sonalyze executable (required)")
	configFilenamePtr := progOpts.Container.String("config-file", "", "Path to system config file (required)")
	outputPathPtr := progOpts.Container.String("output-path", ".", "Path to output directory")
	tagPtr := progOpts.Container.String("tag", "", "Tag for output files")
	hourlyPtr := progOpts.Container.Bool("hourly", true, "Bucket data hourly")
	dailyPtr := progOpts.Container.Bool("daily", false, "Bucket data daily")
	nonePtr := progOpts.Container.Bool("none", false, "Do not bucket data")
	downtimePtr := progOpts.Container.Bool("with-downtime", false, "Include downtime data")
	err := progOpts.Parse(args)
	if err != nil {
		return err
	}
	sonalyzePath, err := util.CleanPath(*sonalyzePathPtr, "-sonalyze")
	if err != nil {
		return err
	}
	configFilename, err := util.CleanPath(*configFilenamePtr, "-config-file")
	if err != nil {
		return err
	}
	outputPath, err := util.CleanPath(*outputPathPtr, "-output-path")
	if err != nil {
		return err
	}

	// Assemble sonalyze arguments

	loadArguments := loadInitialArgs(configFilename)
	loadArguments = append(loadArguments, "--data-path", progOpts.DataPath)
	downtimeArguments := downtimeInitialArgs()
	downtimeArguments = append(downtimeArguments, "--data-path", progOpts.DataPath)
	if progOpts.HaveFrom {
		loadArguments = append(loadArguments, "--from", progOpts.FromStr)
		downtimeArguments = append(downtimeArguments, "--from", progOpts.FromStr)
	}
	if progOpts.HaveTo {
		loadArguments = append(loadArguments, "--to", progOpts.ToStr)
		downtimeArguments = append(downtimeArguments, "--to", progOpts.ToStr)
	}

	// This handling of bucketing isn't completely clean but it's good enough for not-insane users.
	// We can use flag.Visit() to do a better job, if we want.

	var bucketing string
	if *nonePtr {
		loadArguments = append(loadArguments, "--none")
		bucketing = "none"
	} else if *dailyPtr {
		loadArguments = append(loadArguments, "--daily")
		bucketing = "daily"
	} else if *hourlyPtr {
		loadArguments = append(loadArguments, "--hourly")
		bucketing = "hourly"
	} else {
		return errors.New("One of --daily, --hourly, or --none is required")
	}

	// Obtain all the data

	loadOutput, err := runSonalyze(sonalyzePath, loadArguments)
	if err != nil {
		return err
	}
	loadData, err := parseLoadOutput(loadOutput)
	if err != nil {
		return err
	}

	var downtimeData []*downtimeDataByHost
	if *downtimePtr {
		downtimeOutput, err := runSonalyze(sonalyzePath, downtimeArguments)
		if err != nil {
			return err
		}
		downtimeData, err = parseDowntimeOutput(downtimeOutput)
		if err != nil {
			return err
		}
		quantizeDowntimeData(downtimeData, loadData)
	}

	configInfo, err := storage.ReadConfig(configFilename)
	if err != nil {
		return err
	}

	// Convert selected fields to JSON

	downtimeData = downtimeData
	return writePlots(outputPath, *tagPtr, bucketing, configInfo, loadData, downtimeData)
}

func runSonalyze(sonalyzePath string, arguments []string) (string, error) {
	cmd := exec.Command(sonalyzePath, arguments...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.Join(err, errors.New(stderr.String()))
	}
	return stdout.String(), nil
}

func writePlots(
	outputPath, tag, bucketing string,
	configInfo []*storage.SystemConfig,
	loadData []*loadDataByHost,
	downtimeData []*downtimeDataByHost,
) error {
	// configInfo may be nil and this function should still work
	// downtimeData may be nil, in which case it should be ignored, but if not nil it must have been
	//  quantized already

	type perHost struct {
		Date      string                `json:"date"`
		Host      string                `json:"hostname"`
		Tag       string                `json:"tag"`
		Bucketing string                `json:"bucketing"`
		Labels    []string              `json:"labels"`  // formatted timestamps, for now
		Rcpu      []float64             `json:"rcpu"`
		Rgpu      []float64             `json:"rgpu"`
		Rmem      []float64             `json:"rmem"`
		Rgpumem   []float64             `json:"rgpumem"`
		DownHost  []int                 `json:"downhost"`
		DownGpu   []int                 `json:"downgpu"`
		System    *storage.SystemConfig `json:"system"`
	}

	// Use the same timestamp for all records
	now := time.Now().Format(util.DateTimeFormat)

	for _, hd := range loadData {
		var basename string
		if tag == "" {
			basename = hd.host + ".json"
		} else {
			basename = hd.host + "-" + tag + ".json"
		}
		filename := path.Join(outputPath, basename)
		output_file, err := os.CreateTemp(path.Dir(filename), "naicreport-webload")
		if err != nil {
			return err
		}

		labels := make([]string, 0)
		rcpuData := make([]float64, 0)
		rgpuData := make([]float64, 0)
		rmemData := make([]float64, 0)
		rgpumemData := make([]float64, 0)
		for _, d := range hd.data {
			labels = append(labels, d.datetime.Format("01-02 15:04"))
			rcpuData = append(rcpuData, d.rcpu)
			rgpuData = append(rgpuData, d.rgpu)
			rmemData = append(rmemData, d.rmem)
			rgpumemData = append(rgpumemData, d.rgpumem)
		}
		downHost, downGpu := generateDowntimeData(hd, downtimeData)
		var system *storage.SystemConfig
		if configInfo != nil {
			for _, s := range configInfo {
				if s.Hostname == hd.host {
					system = s
					break
				}
			}
		}
		bytes, err := json.Marshal(perHost{
			Date:      now,
			Host:      hd.host,
			Tag:       tag,
			Bucketing: bucketing,
			Labels:    labels,
			Rcpu:      rcpuData,
			Rgpu:      rgpuData,
			Rmem:      rmemData,
			Rgpumem:   rgpumemData,
			DownHost:  downHost,
			DownGpu:   downGpu,
			System:    system,
		})
		if err != nil {
			return err
		}
		output_file.Write(bytes)

		oldname := output_file.Name()
		output_file.Close()
		os.Rename(oldname, filename)
	}

	return nil
}

func generateDowntimeData(ld *loadDataByHost, dd []*downtimeDataByHost) (downHost []int, downGpu []int) {
	if dd == nil {
		return
	}

	downtimeData := findDowntimeDataByHost(dd, ld.host)
	if downtimeData == nil {
		return
	}

	loadData := ld.data
	downHost = make([]int, 0, len(loadData))
	downGpu = make([]int, 0, len(loadData))

	// Walk the loadData array, it is sorted ascending by time and usually has more data points than
	// the downtimeData and is the arbiter of time.

	hostDown := 0
	gpuDown := 0
	ddix := 0
	for _, ld := range loadData {
		for ddix < len(downtimeData) && downtimeData[ddix].start.Before(ld.datetime) && downtimeData[ddix].end.Before(ld.datetime) {
			ddix++
		}
		if ddix < len(downtimeData) && downtimeData[ddix].start == ld.datetime {
			if downtimeData[ddix].device == "host" {
				hostDown = 1
			} else {
				gpuDown = 1
			}
		}
		downHost = append(downHost, hostDown)
		downGpu = append(downGpu, gpuDown)
		if ddix < len(downtimeData) && downtimeData[ddix].end == ld.datetime {
			if downtimeData[ddix].device == "host" {
				hostDown = 0
			} else {
				gpuDown = 0
			}
		}
	}

	return
}

func findDowntimeDataByHost(dd []*downtimeDataByHost, host string) []*downtimeDatum {
	loc := sort.Search(len(dd), func(i int) bool {
		return dd[i].host >= host
	})
	if loc == len(dd) {
		panic(fmt.Sprintf("Unexpected: %v\n%v", host, dd))
	}
	return dd[loc].data
}

// The downtime data must be quantized to the resolution of the load report.  This means eg that if
// the load report is hourly, and the system was down for some fraction of a particular hour, then
// the downtime is stretched to the beginning of the hour (if downtime started within the window) or
// to the end of the hour (if downtime ended within the window).  This will look slightly strange in
// the graph because both downtime and activity will be shown in the same time slot, but is
// basically the right thing.
//
// It is the load report that has the last say on what the time windows are and when the reporting
// starts and ends.
//
// If any data point in downtime falls outside the array it's OK to just discard it, at least for
// now.  (It may be that this is not the right thing for the rightmost point, where we really want
// to know if a system is currently down.)

func quantizeDowntimeData(downtimeData []*downtimeDataByHost, loadData []*loadDataByHost) {

	// The inner data of loadData (per host) are already sorted by ascending time, and it can be
	// binary-searched.
	//
	// Basic algorithm:
	//
	//  Per host
	//    For every time point t in {start,end} on the downtime data
	//      Find a window w on the load data s.t. w.start <= t < t.end
	//        If t is start, update t to be w.start
	//        else update t to be w.end (which is the start of the next w)

	for _, dh := range downtimeData {
		ld := findLoadDataByHost(loadData, dh.host)
		for _, dd := range dh.data {
			s1, _ := findLoadDataWindowByTime(ld, dd.start)
			if s1 == nil {
				dd.start = time.Unix(0, 0) // Mark as error
				dd.end = time.Unix(0, 0)
				continue
			}
			_, e2 := findLoadDataWindowByTime(ld, dd.end)
			if e2 == nil {
				dd.start = time.Unix(0, 0) // Mark as error
				dd.end = time.Unix(0, 0)
				continue
			}
			dd.start = s1.datetime
			dd.end = e2.datetime
		}
	}
}

func findLoadDataByHost(ld []*loadDataByHost, host string) []*loadDatum {
	loc := sort.Search(len(ld), func(i int) bool {
		return ld[i].host >= host
	})
	if loc == len(ld) {
		for _, x := range ld {
			fmt.Printf("%s\n", x.host)
		}
		panic(fmt.Sprintf("Unexpected: %s\n%v", host, ld))
	}
	return ld[loc].data
}

func findLoadDataWindowByTime(ld []*loadDatum, t time.Time) (*loadDatum, *loadDatum) {
	loc := sort.Search(len(ld), func(i int) bool {
		return !t.After(ld[i].datetime)
	})
	// TODO: An argument could be made about the right end point of the timeline here that we should
	// be able to include it somehow.
	if loc >= len(ld)-1 {
		return nil, nil
	}
	return ld[loc], ld[loc+1]
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Handle `sonalyze uptime`.

// TODO: In sonalyze, "start_utc" or something like that should mean a timestamp, not a formatted
// date.  Then we can just slurp that in here and avoid overhead and complexity.

func downtimeInitialArgs() []string {
	return []string{
		"uptime",
		"--interval", "4",
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
		return nil, err
	}

	// The output from `sonalyze downtime` is sorted first by host, then by ascending start time.

	var outerData = make([]*downtimeDataByHost, 0)
	var innerData = make([]*downtimeDatum, 0)
	for _, repr := range rawData {
		// Convert some values
		start, startErr := time.Parse(util.DateTimeFormat, repr.Start)
		end, endErr := time.Parse(util.DateTimeFormat, repr.End)
		if startErr != nil || endErr != nil {
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
// Handle `sonalyze load`.  Currently this uses csv format but that's not necessary - it's a
// holdover from an older design.

// TODO: Switch to JSON, probably, and note item higher up about "start_utc" - here we want
// "datetime_utc" or something like that.

func loadInitialArgs(configFilename string) []string {
	return []string{
		"load",
		"--config-file", configFilename,
		"--fmt=csvnamed,datetime,cpu,mem,gpu,gpumem,rcpu,rmem,rgpu,rgpumem,gpus,host",
	}
}

type loadDatum struct {
	datetime time.Time
	cpu      float64
	mem      float64
	gpu      float64
	gpumem   float64
	gpus     []uint32 // nil for "unknown"
	rcpu     float64
	rmem     float64
	rgpu     float64
	rgpumem  float64
	host     string // redundant but maybe useful
}

type loadDataByHost struct {
	host string
	data []*loadDatum
}

// In the returned data, the point data in the inner list are sorted by ascending time, and the
// outer list is sorted by ascending host name.

func parseLoadOutput(output string) ([]*loadDataByHost, error) {
	rows, err := storage.ParseFreeCSV(strings.NewReader(output))
	if err != nil {
		return nil, err
	}

	// The output from sonalyze is sorted first by host, then by increasing time.  Thus it's fine to
	// read record-by-record, bucket by host easily, and then assume that data are sorted within host.

	allData := make([]*loadDataByHost, 0)

	var curData []*loadDatum
	curHost := ""
	for _, row := range rows {
		success := true
		newHost := storage.GetString(row, "host", &success)
		if !success {
			continue
		}
		if newHost != curHost {
			if curData != nil {
				allData = append(allData, &loadDataByHost{host: curHost, data: curData})
			}
			curData = make([]*loadDatum, 0)
			curHost = newHost
		}
		newDatum := &loadDatum{
			datetime: storage.GetDateTime(row, "datetime", &success),
			cpu:      storage.GetFloat64(row, "cpu", &success),
			mem:      storage.GetFloat64(row, "mem", &success),
			gpu:      storage.GetFloat64(row, "gpu", &success),
			gpumem:   storage.GetFloat64(row, "gpumem", &success),
			gpus:     nil,
			rcpu:     storage.GetFloat64(row, "rcpu", &success),
			rmem:     storage.GetFloat64(row, "rmem", &success),
			rgpu:     storage.GetFloat64(row, "rgpu", &success),
			rgpumem:  storage.GetFloat64(row, "rgpumem", &success),
			host:     newHost,
		}
		gpuRepr := storage.GetString(row, "gpus", &success)
		var gpuData []uint32 // Unknown set
		if gpuRepr != "unknown" {
			gpuData = make([]uint32, 0) // Empty set
			if gpuRepr != "none" {
				for _, t := range strings.Split(gpuRepr, ",") {
					n, err := strconv.ParseUint(t, 10, 32)
					if err == nil {
						gpuData = append(gpuData, uint32(n))
					}
				}
			}
		}
		newDatum.gpus = gpuData
		if !success {
			continue
		}
		curData = append(curData, newDatum)
	}
	if curData != nil {
		allData = append(allData, &loadDataByHost{host: curHost, data: curData})
	}

	return allData, nil
}
