// This runs `sinfo` on a cluster and processes the output to generate JSON describing the cluster.
// The output is printed on stdout.
//
// Optionally, it can take its input from a file generated by `sinfo -a -o '%n/%f/%m/%X/%Y/%Z'`,
// with the -input option.
//
// Data not available from `sinfo` can optionally be provided by a file named by the -aux parameter.
// This is a JSON file with the same format as the output (ie a Jobanalyzer system-config file),
// where each field for a given host supplies a default value for that field.  For hosts that are
// not revealed by `sinfo`, it supplies all fields.  It also carries metadata that makes `slurminfo`
// modify the data slightly.  See the README.md in this directory.
//
// TODO (none important):
// - compress the entries coming from background information
// - ability to mark some entries in the background information as pure background, ie,
//   not to be output in the final file, this allows the use of host number ranges that
//   cover non-existent hosts.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
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

// Sorting boilerplate.
type configSlice []*config.SystemConfig
func (a configSlice) Len() int           { return len(a) }
func (a configSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a configSlice) Less(i, j int) bool { return a[i].Hostname < a[j].Hostname }

func main() {
	var auxFilename string
	var inputFilename string

	flag.StringVar(&auxFilename, "aux", "", "Auxiliary information in `filename`")
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
		stdout, stderr, err = process.RunSubprocess("sinfo", []string{"-a", "-o", "%n/%f/%m/%X/%Y/%Z"})
		Check(err, "Running 'sinfo'")
		if stderr != "" {
			fmt.Fprintln(os.Stderr, stderr)
		}
	}

	// Read aux information.  `referenced` will be used to track whether an entry is used for
	// background information or entirely.

	var auxConfig map[string]*config.SystemConfig
	var referenced = make(map[string]bool)
	if auxFilename != "" {
		var err error
		auxConfig, err = config.ReadConfig(auxFilename, true)
		Check(err, "Reading aux")
	} else {
		auxConfig = make(map[string]*config.SystemConfig)
	}

	// Build a map from system attributes to sets of host names.

	type sysAttrs struct {
		memMB                                   uint64
		sockets, cores, threads, gpus, gpuMemGB int
		manufacturer, gpuModel, suffix          string
		gpuMemPct, slurmNode                    bool
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
		var gpus int
		var gpuMemGB int
		var gpuMemPct bool
		var gpuModel string
		var suffix string
		if background, found := auxConfig[name]; found {
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
			slurmNode:    true,
			suffix:       suffix,
		}
		if bag, found := systems[sd]; found {
			systems[sd] = append(bag, name)
		} else {
			systems[sd] = []string{name}
		}
	}

	// Construct the output from the system attribute map.

	results := make([]*config.SystemConfig, 0)
	for desc, hosts := range systems {
		names, err := hostglob.CompressHostnames(hosts)
		Check(err, fmt.Sprintf("%v", hosts))
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
			results = append(results, &config.SystemConfig{
				Hostname:    h + desc.suffix,
				Description: description,
				CpuCores:    desc.sockets * desc.cores * desc.threads,
				MemGB:       int(memgb),
				MultiNode:   desc.slurmNode,
				GpuCards:    desc.gpus,
				GpuMemGB:    desc.gpuMemGB,
				GpuMemPct:   desc.gpuMemPct,
			})
		}
	}

	// Add all background information that was not referenced.
	//
	// We would have preferred to do this earlier so that the information could be (re)compressed, but
	// that turns out to require some reengineering I'm not prepared to do yet.

	for name, aux := range auxConfig {
		if _, found := referenced[name]; !found {
			Assert(aux.CpuCores > 0 && aux.MemGB > 0, "Bad system default "+aux.Hostname)
			for _, m := range aux.Metadata {
				switch m.Key {
				case "host-suffix":
					aux.Hostname += m.Value
				}
			}
			aux.Metadata = nil
			results = append(results, aux)
		}
	}

	// We print in host name order.

	sort.Sort(configSlice(results))
	outBytes, err := json.MarshalIndent(&results, "", " ")
	Check(err, "JSON data")
	io.WriteString(os.Stdout, string(outBytes)+"\n")
}