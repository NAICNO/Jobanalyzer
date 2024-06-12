// This runs `sinfo` on a cluster and processes the output to generate JSON describing the cluster.
// The output is printed on stdout.
//
// Optionally, it can take its input from a file generated by `sinfo -a -o '%n/%f/%m/%X/%Y/%Z'`,
// with the -input option.
//
// Data not available from `sinfo` can optionally be provided by a file named by the -background
// parameter.  This is a JSON file containing an array of NodeConfigRecords, where each field for a
// given host supplies a default value for that field.  For hosts that are not revealed by `sinfo`,
// it supplies all fields.  It also carries metadata that makes `slurminfo` modify the data
// slightly.  See the README.md in this directory.
//
// TODO (none important):
// - compress the entries coming from background information

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-utils/config"
	. "go-utils/error"
	"go-utils/hostglob"
	"go-utils/process"
)

const (
	// First line expected from sinfo
	header = "HOSTNAMES/AVAIL_FEATURES/MEMORY/SOCKETS/CORES/THREADS"
)

func main() {
	var backgroundFilename string
	var inputFilename string

	flag.StringVar(&backgroundFilename, "background", "", "Background information in `filename`")
	flag.StringVar(&inputFilename, "input", "", "Input in `filename`, don't run sinfo")
	flag.Parse()

	// Read sinfo data.

	var stdout string
	if inputFilename != "" {
		bytes, err := os.ReadFile(inputFilename)
		Check(err, "Reading input")
		stdout = string(bytes)
	} else {
		var err error
		var stderr string
		stdout, stderr, err = process.RunSubprocess("sinfo", "sinfo", []string{"-a", "-o", "%n/%f/%m/%X/%Y/%Z"})
		Check(err, "Running 'sinfo'")
		if stderr != "" {
			fmt.Fprintln(os.Stderr, stderr)
		}
	}

	// Read background information.  `referenced` will be used to track whether an entry is used for
	// background information or entirely.

	var backgroundConfig map[string]*config.NodeConfigRecord
	var referenced = make(map[string]bool)
	if backgroundFilename != "" {
		var err error
		backgroundConfig, err = config.ReadBackgroundFile(backgroundFilename)
		Check(err, "Background")
	} else {
		backgroundConfig = make(map[string]*config.NodeConfigRecord)
	}

	// Build a map from system attributes to sets of host names.

	type sysAttrs struct {
		memMB                                   uint64
		sockets, cores, threads, gpus, gpuMemGB int
		manufacturer, gpuModel, suffix          string
		gpuMemPct, crossNode                    bool
	}

	var systems = make(map[sysAttrs][]string)

	inputLines := strings.Split(strings.TrimSpace(stdout), "\n")
	Assert(len(inputLines) > 0, "Empty input")
	Assert(inputLines[0] == header, "Bad header in input: "+inputLines[0])
	inputLines = inputLines[1:]

	for _, l := range inputLines {
		fields := strings.Split(l, "/")
		Assert(len(fields) == 6, "Bad fields in input line: "+l)
		name := fields[0]
		features := make(map[string]bool)
		for _, f := range strings.Split(fields[1], ",") {
			features[f] = true
		}
		mem, err := strconv.ParseUint(fields[2], 10, 64)
		Check(err, fields[2])
		sockets, err := strconv.ParseUint(fields[3], 10, 64)
		Check(err, fields[3])
		cores, err := strconv.ParseUint(fields[4], 10, 64)
		Check(err, fields[4])
		threads, err := strconv.ParseUint(fields[5], 10, 64)
		Check(err, fields[5])
		manufacturer := "intel"
		if _, found := features["amd"]; found {
			manufacturer = "amd"
		}
		var (
			crossNodeJobs bool
			gpus          int
			gpuMemGB      int
			gpuMemPct     bool
			gpuModel      string
			suffix        string
		)
		if background, found := backgroundConfig[name]; found {
			crossNodeJobs = background.CrossNodeJobs
			gpus = background.GpuCards
			gpuMemGB = background.GpuMemGB
			gpuMemPct = background.GpuMemPct
			if gpus > 0 {
				gpuModel = "unknown-gpu"
			}
			for _, m := range background.Metadata {
				switch m.Key {
				case "gpu":
					gpuModel = m.Value
				case "host-suffix":
					suffix = m.Value
				}
			}
			referenced[name] = true
		}
		if (strings.HasPrefix(name, "gpu") || strings.HasPrefix(name, "accel")) && gpus == 0 {
			fmt.Fprintf(os.Stderr, "WARNING: Host name '%s' suggests a GPU node but no GPU info is found in background data\n", name)
		}
		sd := sysAttrs{
			memMB:        mem,
			sockets:      int(sockets),
			cores:        int(cores),
			threads:      int(threads),
			gpus:         gpus,
			gpuMemGB:     gpuMemGB,
			manufacturer: manufacturer,
			gpuModel:     gpuModel,
			gpuMemPct:    gpuMemPct,
			crossNode:    crossNodeJobs,
			suffix:       suffix,
		}
		if bag, found := systems[sd]; found {
			systems[sd] = append(bag, name)
		} else {
			systems[sd] = []string{name}
		}
	}

	// Construct the output from the system attribute map.

	nodes := make([]*config.NodeConfigRecord, 0)
	for desc, hosts := range systems {
		names := hostglob.CompressHostnames(hosts)
		for _, h := range names {
			ht := ""
			if desc.threads > 1 {
				ht = " (hyperthreaded)"
			}
			memgb := desc.memMB / 1024
			gpu := ""
			if desc.gpus > 0 {
				gpu = fmt.Sprintf(", %dx %s @ %dGB", desc.gpus, desc.gpuModel, desc.gpuMemGB/desc.gpus)
			}
			description := fmt.Sprintf(
				"%dx%d %s%s, %dGB%s",
				desc.sockets,
				desc.cores,
				desc.manufacturer,
				ht,
				memgb,
				gpu,
			)
			nodes = append(nodes, &config.NodeConfigRecord{
				Hostname:      h + desc.suffix,
				Description:   description,
				CrossNodeJobs: desc.crossNode,
				CpuCores:      desc.sockets * desc.cores * desc.threads,
				MemGB:         int(memgb),
				GpuCards:      desc.gpus,
				GpuMemGB:      desc.gpuMemGB,
				GpuMemPct:     desc.gpuMemPct,
			})
		}
	}

	// Add all background information that was not referenced.
	//
	// We would have preferred to do this earlier so that the information could be (re)compressed, but
	// that turns out to require some reengineering I'm not prepared to do yet.

	for name, background := range backgroundConfig {
		if _, found := referenced[name]; !found {
			if background.CpuCores == 0 || background.MemGB == 0 {
				continue
			}
			for _, m := range background.Metadata {
				switch m.Key {
				case "host-suffix":
					background.Hostname += m.Value
				}
			}
			background.Metadata = nil
			nodes = append(nodes, background)
		}
	}

	results := config.NewClusterConfig(
		0,
		"",
		"",
		[]string{},
		[]string{},
		nodes,
	)

	// Save the config.  It is JSON text and is just dumped on stdout.

	config.WriteConfigTo(os.Stdout, results)
}
