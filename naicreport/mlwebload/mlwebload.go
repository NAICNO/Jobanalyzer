// Generate data for plotting the running load of the ML systems.  The data are taken from the live
// sonar logs, by means of sonalyze.

package mlwebload

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"naicreport/storage"
	"naicreport/util"
)

type systemConfig struct {
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
	CpuCores    int    `json:"cpu_cores"`
	MemGB       int    `json:"mem_gb"`
	GpuCards    int    `json:"gpu_cards"`
	GpuMemGB    int    `json:"gpumem_gb"`
}

func MlWebload(progname string, args []string) error {
	// Parse and sanitize options

	progOpts := util.NewStandardOptions(progname + " ml-webload")
	sonalyzePathPtr := progOpts.Container.String("sonalyze", "", "Path to sonalyze executable (required)")
	configFilenamePtr := progOpts.Container.String("config-file", "", "Path to system config file (required)")
	outputPathPtr := progOpts.Container.String("output-path", ".", "Path to output directory")
	tagPtr := progOpts.Container.String("tag", "", "Tag for output files")
	hourlyPtr := progOpts.Container.Bool("hourly", true, "Bucket data hourly")
	dailyPtr := progOpts.Container.Bool("daily", false, "Bucket data daily")
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

	// Assemble sonalyze arguments and run it, collecting its output

	arguments := []string{
		"load",
		"--data-path", progOpts.DataPath,
		"--config-file", configFilename,
		"--fmt=csvnamed," + sonalyzeFormat,
	}
	if progOpts.HaveFrom {
		arguments = append(arguments, "--from", progOpts.FromStr)
	}
	if progOpts.HaveTo {
		arguments = append(arguments, "--to", progOpts.ToStr)
	}
	// This isn't completely clean but it's good enough for not-insane users.
	// We can use flag.Visit() to do a better job.  This is true in general.
	var bucketing string
	if *dailyPtr {
		arguments = append(arguments, "--daily")
		bucketing = "daily"
	} else if *hourlyPtr {
		arguments = append(arguments, "--hourly")
		bucketing = "hourly"
	} else {
		return errors.New("One of --daily or --hourly is required")
	}

	cmd := exec.Command(sonalyzePath, arguments...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return errors.Join(err, errors.New(stderr.String()))
	}

	// Interpret the output from sonalyze

	output, err := parseOutput(stdout.String())
	if err != nil {
		return err
	}

	// Get the system config if possible

	var configInfo []*systemConfig
	configFile, err := os.Open(configFilename)
	if err == nil {
		bytes, err := io.ReadAll(configFile)
		if err == nil {
			err := json.Unmarshal(bytes, &configInfo)
			if err != nil {
				configInfo = nil
			}
		}
	}

	// Convert selected fields to JSON

	return writePlots(outputPath, *tagPtr, bucketing, configInfo, output)
}

func writePlots(outputPath, tag, bucketing string, configInfo []*systemConfig, output []*hostData) error {
	// configInfo may be nil

	type perPoint struct {
		X string  `json:"x"`
		Y float64 `json:"y"`
	}

	type perHost struct {
		Date      string        `json:"date"`
		Hostname  string        `json:"hostname"`
		Tag       string        `json:"tag"`
		Bucketing string        `json:"bucketing"`
		Rcpu      []perPoint    `json:"rcpu"`
		Rgpu      []perPoint    `json:"rgpu"`
		Rmem      []perPoint    `json:"rmem"`
		Rgpumem   []perPoint    `json:"rgpumem"`
		System    *systemConfig `json:"system"`
	}

	// Use the same timestamp for all records
	now := time.Now().Format(util.DateTimeFormat)

	for _, hd := range output {
		var basename string
		if tag == "" {
			basename = hd.hostname + ".json"
		} else {
			basename = hd.hostname + "-" + tag + ".json"
		}
		filename := path.Join(outputPath, basename)
		output_file, err := os.CreateTemp(path.Dir(filename), "naicreport-webload")
		if err != nil {
			return err
		}

		rcpuData := make([]perPoint, 0)
		rgpuData := make([]perPoint, 0)
		rmemData := make([]perPoint, 0)
		rgpumemData := make([]perPoint, 0)
		for _, d := range hd.data {
			ts := d.datetime.Format("01-02 15:04")
			rcpuData = append(rcpuData, perPoint{ts, d.rcpu})
			rgpuData = append(rgpuData, perPoint{ts, d.rgpu})
			rmemData = append(rmemData, perPoint{ts, d.rmem})
			rgpumemData = append(rgpumemData, perPoint{ts, d.rgpumem})
		}
		var system *systemConfig
		if configInfo != nil {
			for _, s := range configInfo {
				if s.Hostname == hd.hostname {
					system = s
					break
				}
			}
		}
		bytes, err := json.Marshal(perHost{
			Date:      now,
			Hostname:  hd.hostname,
			Tag:       tag,
			Bucketing: bucketing,
			Rcpu:      rcpuData,
			Rgpu:      rgpuData,
			Rmem:      rmemData,
			Rgpumem:   rgpumemData,
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

const (
	sonalyzeFormat = "datetime,cpu,mem,gpu,gpumem,rcpu,rmem,rgpu,rgpumem,gpus,host"
)

type datum struct {
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
	hostname string // redundant but maybe useful
}

type hostData struct {
	hostname string
	data     []*datum
}

// The output from sonalyze is sorted first by host, then by increasing time.  Thus it's fine to
// read record-by-record, bucket by host easily, and then assume that data are sorted within host.

func parseOutput(output string) ([]*hostData, error) {
	rows, err := storage.ParseFreeCSV(strings.NewReader(output))
	if err != nil {
		return nil, err
	}

	allData := make([]*hostData, 0)

	var curData []*datum
	curHost := ""
	for _, row := range rows {
		success := true
		newHost := storage.GetString(row, "host", &success)
		if !success {
			continue
		}
		if newHost != curHost {
			if curData != nil {
				allData = append(allData, &hostData{hostname: curHost, data: curData})
			}
			curData = make([]*datum, 0)
			curHost = newHost
		}
		newDatum := &datum{
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
			hostname: newHost,
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
		allData = append(allData, &hostData{hostname: curHost, data: curData})
	}

	return allData, nil
}
