package main

import (
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


// Return a map from host to info for the host.

func readSinfo(background map[string]*config.NodeConfigRecord) map[string]*config.NodeConfigRecord {
	// Read sinfo data.

	var stdout string
	if sinfoInput != "" {
		var err error
		var bytes []byte
		if sinfoInput == "-" {
			panic("stdin not implemented")
		} else {
			bytes, err = os.ReadFile(sinfoInput)
		}
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
		if bgNode, found := background[name]; found {
			crossNodeJobs = bgNode.CrossNodeJobs
			gpus = bgNode.GpuCards
			gpuMemGB = bgNode.GpuMemGB
			gpuMemPct = bgNode.GpuMemPct
			if bgNode.GpuCards > 0 {
				gpuModel = "unknown-gpu"
			}
			for _, m := range bgNode.Metadata {
				switch m.Key {
				case "gpu":
					gpuModel = m.Value
				case "host-suffix":
					suffix = m.Value
				}
			}
			if (strings.HasPrefix(name, "gpu") || strings.HasPrefix(name, "accel")) && bgNode.GpuCards == 0 {
				fmt.Fprintf(
					os.Stderr,
					"WARNING: Host name '%s' suggests a GPU node but no GPU info is found in background data\n",
					name,
				)
			}
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

	nodes := make(map[string]*config.NodeConfigRecord, 0)
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
			nodes[h] = &config.NodeConfigRecord{
				Hostname:      h,
				Description:   description,
				CrossNodeJobs: desc.crossNode,
				CpuCores:      desc.sockets * desc.cores * desc.threads,
				MemGB:         int(memgb),
				GpuCards:      desc.gpus,
				GpuMemGB:      desc.gpuMemGB,
				GpuMemPct:     desc.gpuMemPct,
			}
		}
	}

	return nodes
}
