// Generate data for plotting the running load of a cluster.  The data are taken from the live sonar
// logs, by means of sonalyze.

package load

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"naicreport/storage"
	"naicreport/util"
)

func Load(progname string, args []string) error {
	// Parse and sanitize options

	progOpts := util.NewStandardOptions(progname + " load")
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
	downtimeArguments := downtimeInitialArgs()

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

	// For -- this must come last, so do standard options (from/to and files) last always

	loadArguments = util.AddStandardOptions(loadArguments, progOpts)
	downtimeArguments = util.AddStandardOptions(downtimeArguments, progOpts)

	// Obtain all the data

	loadOutput, err := util.RunSubprocess(sonalyzePath, loadArguments)
	if err != nil {
		return err
	}
	loadData, err := parseLoadOutput(loadOutput)
	if err != nil {
		return err
	}

	var downtimeData []*downtimeDataByHost
	if *downtimePtr {
		downtimeOutput, err := util.RunSubprocess(sonalyzePath, downtimeArguments)
		if err != nil {
			return err
		}
		downtimeData, err = parseDowntimeOutput(downtimeOutput)
		if err != nil {
			return err
		}
	}

	configInfo, err := storage.ReadConfig(configFilename)
	if err != nil {
		return err
	}

	// Convert selected fields to JSON

	downtimeData = downtimeData
	return writePlots(outputPath, *tagPtr, bucketing, configInfo, loadData, downtimeData)
}

func writePlots(
	outputPath, tag, bucketing string,
	configInfo map[string]*storage.SystemConfig,
	loadData []*loadDataByHost,
	downtimeData []*downtimeDataByHost,
) error {
	// The config for a host may be missing, but this should still work.
	//
	// downtimeData may be nil, in which case it should be ignored, but if not nil it must have been
	//  quantized already

	type perHost struct {
		Date      string                `json:"date"`
		Host      string                `json:"hostname"`
		Tag       string                `json:"tag"`
		Bucketing string                `json:"bucketing"`
		Labels    []string              `json:"labels"` // formatted timestamps, for now
		Rcpu      []float64             `json:"rcpu"`
		Rmem      []float64             `json:"rmem"`
		Rgpu      []float64             `json:"rgpu,omitempty"`
		Rgpumem   []float64             `json:"rgpumem,omitempty"`
		DownHost  []int                 `json:"downhost,omitempty"`
		DownGpu   []int                 `json:"downgpu,omitempty"`
		System    *storage.SystemConfig `json:"system,omitempty"`
	}

	// Use the same timestamp for all records
	now := time.Now().Format(util.DateTimeFormat)

	for _, hd := range loadData {
		system, _ := configInfo[hd.host]
		var basename string
		if tag == "" {
			basename = hd.host + ".json"
		} else {
			basename = hd.host + "-" + tag + ".json"
		}
		filename := path.Join(outputPath, basename)
		output_file, err := os.CreateTemp(path.Dir(filename), "naicreport-load")
		if err != nil {
			return err
		}

		hasGpu := system != nil && system.GpuCards > 0
		labels := make([]string, 0)
		rcpuData := make([]float64, 0)
		rmemData := make([]float64, 0)
		var rgpuData, rgpumemData []float64
		if hasGpu {
			rgpuData = make([]float64, 0)
			rgpumemData = make([]float64, 0)
		}
		for _, d := range hd.data {
			labels = append(labels, d.datetime.Format("01-02 15:04"))
			rcpuData = append(rcpuData, d.rcpu)
			rmemData = append(rmemData, d.rmem)
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
		data := perHost{
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
		}
		bytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("While marshaling perHost data: %v %w", data, err)
		}
		output_file.Write(bytes)

		oldname := output_file.Name()
		output_file.Close()
		os.Rename(oldname, filename)
	}

	return nil
}

func generateDowntimeData(ld *loadDataByHost, dd []*downtimeDataByHost, hasGpu bool) (downHost []int, downGpu []int) {
	if dd == nil {
		return
	}

	ddix := sort.Search(len(dd), func(i int) bool {
		return dd[i].host >= ld.host
	})
	if ddix == len(dd) {
		/* This is possible because we run sonalyze uptime with --only-down:
		   it's possible for there to be no downtime in the time window. */
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
		for ix := max(loc-1, 0); ix < len(loadData) && loadData[ix].datetime.Before(ddval.end); ix++ {
			if isDevice {
				downHost[ix] = 1
			} else if hasGpu {
				downGpu[ix] = 1
			}
		}
	}
	return
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
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
		return nil, fmt.Errorf("While unmarshaling downtime data: %w", err)
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
