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
	"time"

	"go-utils/process"
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
	clusterPtr := progOpts.Container.String("cluster", "", "Cluster these host name patterns (comma-separated) (requires bucketing, too)")
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

	if *clusterPtr != "" {
		if bucketing == "none" {
			return errors.New("Cannot --cluster together with --none")
		}
		if *downtimePtr {
			return errors.New("Cannot --cluster together with --with-downtime")
		}
		loadArguments = append(loadArguments, "--cluster")

		patterns, err := storage.SplitHostnames(*clusterPtr)
		if err != nil {
			return err
		}
		for _, p := range patterns {
			loadArguments = append(loadArguments, "--host", p)
		}
	}

	// For -- this must come last, so do standard options (from/to and files) last always

	loadArguments = util.AddStandardOptions(loadArguments, progOpts)
	downtimeArguments = util.AddStandardOptions(downtimeArguments, progOpts)

	// Obtain all the data

	loadOutput, err := process.RunSubprocess(sonalyzePath, loadArguments)
	if err != nil {
		return err
	}

	loadData, err := parseLoadOutputBySystem(loadOutput)
	if err != nil {
		return err
	}

	var downtimeData []*downtimeDataByHost
	if *downtimePtr {
		downtimeOutput, err := process.RunSubprocess(sonalyzePath, downtimeArguments)
		if err != nil {
			return err
		}
		downtimeData, err = parseDowntimeOutput(downtimeOutput)
		if err != nil {
			return err
		}
	}

	return writePlots(outputPath, *tagPtr, bucketing, *clusterPtr != "", loadData, downtimeData)
}

func writePlots(
	outputPath, tag, bucketing string,
	clustering bool,
	loadData []*loadDataBySystem,
	downtimeData []*downtimeDataByHost,
) error {
	type perSystem struct {
		Host string `json:"hostname"`
		Description string `json:"description"`
	}

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
		System    *perSystem            `json:"system,omitempty"`
	}

	if clustering && len(loadData) != 1 {
		return fmt.Errorf("Expected exactly one datum for clustered run, tag=%s", tag)
	}

	// The config for a host may be missing, but this should still work.
	//
	// downtimeData may be nil, in which case it should be ignored, but if not nil it must have been
	//  quantized already

	// Use the same timestamp for all records
	now := time.Now().Format(util.DateTimeFormat)

	for _, hd := range loadData {
		var basename string
		if clustering {
			basename = tag + ".json"
		} else if tag == "" {
			basename = hd.system.host + ".json"
		} else {
			basename = hd.system.host + "-" + tag + ".json"
		}
		filename := path.Join(outputPath, basename)
		output_file, err := os.CreateTemp(path.Dir(filename), "naicreport-load")
		if err != nil {
			return err
		}

		hasGpu := hd.system != nil && hd.system.gpuCards > 0
		labels := make([]string, 0)
		rcpuData := make([]float64, 0)
		rmemData := make([]float64, 0)
		var rgpuData, rgpumemData []float64
		if hasGpu {
			rgpuData = make([]float64, 0)
			rgpumemData = make([]float64, 0)
		}
		for _, d := range hd.data {
			labels = append(labels, d.datetime.Format(util.DateTimeFormat))
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
		var system *perSystem
		if hd.system != nil {
			system = &perSystem {
				Host: hd.system.host,
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
// Handle `sonalyze load`.

// Note item higher up about "start_utc" - here we want "datetime_utc" or something like that.

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

func loadInitialArgs(configFilename string) []string {
	return []string{
		"load",
		"--config-file", configFilename,
		"--fmt=json,datetime,cpu,mem,gpu,gpumem,rcpu,rmem,rgpu,rgpumem,gpus,host",
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
		Rgpu     string `json:"rgpu"`
		Rgpumem  string `json:"rgpumem"`
		Host     string `json:"host"`
	}

	type systemDescJSON struct {
		Host string `json:"hostname"`
		Description string `json:"description"`
		GpuCards string `json:"gpucards"`
	}

	type loadDataPackageJSON struct {
		System *systemDescJSON `json:"system"`
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
				datetime: util.JsonDateTime(r.Datetime),
				cpu:      util.JsonFloat64(r.Cpu),
				mem:      util.JsonFloat64(r.Mem),
				gpu:      util.JsonFloat64(r.Gpu),
				gpumem:   util.JsonFloat64(r.Gpumem),
				gpus:     util.JsonGpulist(r.Gpus),
				rcpu:     util.JsonFloat64(r.Rcpu),
				rmem:     util.JsonFloat64(r.Rmem),
				rgpu:     util.JsonFloat64(r.Rgpu),
				rgpumem:  util.JsonFloat64(r.Rgpumem),
				host:     bySystem.System.Host,
			}
			data = append(data, newDatum)
		}
		allData = append(allData, &loadDataBySystem{
			system: &systemDesc{
				host: bySystem.System.Host,
				description: bySystem.System.Description,
				gpuCards: util.JsonInt(bySystem.System.GpuCards),
			},
			data: data,
		})
	}

	return allData, nil
}
