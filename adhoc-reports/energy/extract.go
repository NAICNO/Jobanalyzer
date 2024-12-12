// Extract and join sacct and sinfo data to produce a log of completed jobs for energy monitoring
// simulation purposes.  Somewhat specific to Fox, but will generalize with a little work.

package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	dullJobs = "extern|batch|interactive"
	onlyGpu  = flag.Bool("gpu", false, "Only nodes matching /gpu-.*/")
	dull     = flag.Bool("dull", false, "Include dull jobs (" + dullJobs + ")")
	from     = flag.String("from", "now-2days", "Scan window start time (sacct -S parameter)")
	to       = flag.String("to", "now", "Scan window end time (sacct -E parameter)")
	nomig    = flag.Bool("nomig", false, "Exclude the 'mig' partition")
)

const sacctFields = "JobID,State,Partition,Priority,Submit,Start,End,TimelimitRaw,NodeList,AllocTRES"

var (
	fieldNames = strings.Split(sacctFields, ",")
	nAlloc     = slices.Index(fieldNames, "AllocTRES")
	nNodelist  = slices.Index(fieldNames, "NodeList")
	nPartition = slices.Index(fieldNames, "Partition")
	nJobID     = slices.Index(fieldNames, "JobID")
	nState     = slices.Index(fieldNames, "State")

	// Attributes in AllocTRES field
	tresRe = regexp.MustCompile(`([a-z][a-z0-9/:]*)=(.*)`)

	// Boring job IDs
	dullRe = regexp.MustCompile(`\d+\.(` + dullJobs + `)`)
)

func main() {
	flag.Parse()

	gpus := sinfoInfo()
	// Hack!!  Sinfo on Fox does not reveal this for some reason.
	if _, found := gpus["gpu-10"]; !found {
		gpus["gpu-10"] = gpu{"h100", 2}
	}

	data := sacctInfo()
	for r := range data {
		record := data[r]

		// Create uniform partition names
		if strings.Index(record[nPartition], "accel") != -1 {
			record[nPartition] = "accel"
		} else if strings.Index(record[nPartition], "bigmem") != -1 {
			record[nPartition] = "bigmem"
		} else if record[nPartition] != "normal" && record[nPartition] != "mig" {
			record[nPartition] = "other"
		}

		// Change "CANCELLED by ..." into CANCELLED
		if strings.HasPrefix(record[nState], "CANCELLED") {
			record[nState] = "CANCELLED"
		}

		// CPU count, GPU count, GPU type
		// The AllocTRES is a comma-separated list of name=value
		// We want cpu=n
		// We want gres/gpu:type=m or barring that, gres/gpu=m, in the latter
		//   case we must infer the gpu type from the sinfo
		var gpuType, gpuCount, cpuCount string
		for _, f := range strings.Split(record[nAlloc], ",") {
			if m := tresRe.FindStringSubmatch(f); m != nil {
				switch {
				case m[1] == "cpu":
					cpuCount = m[2]
				case strings.HasPrefix(m[1], "gres/gpu:"):
					// Prefer the first type we see, there really should only be one anyway
					if gpuType == "" {
						gpuType = strings.TrimPrefix(m[1], "gres/gpu:")
						gpuCount = m[2]
					}
				case m[1] == "gres/gpu":
					gpuCount = m[2]
				}
			}
		}
		if gpuType == "" && gpuCount != "" {
			node := record[nNodelist]
			// Reject node lists for GPU nodes for now, but easy to fix
			if strings.IndexAny(node, ",[") != -1 {
				panic("Too many nodes: " + node)
			}
			g := gpus[node]
			if g.kind == "" {
				for k, v := range gpus {
					fmt.Fprintf(os.Stderr, "%s -> %s %d\n", k, v.kind, v.count)
				}
				panic("No gpu for " + node)
			}
			gpuType = g.kind
		}

		// Replace AllocTRES by CPUS and then attach GPUS and GPUType
		record[nAlloc] = cpuCount
		record = append(record, gpuCount, gpuType)

		data[r] = record
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header, same fields are replaced/added as above
	fieldNames[nAlloc] = "CPUS"
	fieldNames = append(fieldNames, "GPUS", "GPUType")
	w.Write(fieldNames)

	for _, r := range data {
		if !*dull && dullRe.MatchString(r[nJobID]) {
			continue
		}
		if *onlyGpu && !strings.HasPrefix(r[nNodelist], "gpu-") {
			continue
		}
		if *nomig && r[nPartition] == "mig" {
			continue
		}
		w.Write(r)
	}
}

func sacctInfo() [][]string {
	sacct := exec.Command(
		"sacct",
		"--noheader",
		"-aP",
		"-s",
		"CANCELLED,COMPLETED,DEADLINE,FAILED,OUT_OF_MEMORY,TIMEOUT",
		"-o",
		sacctFields,
		"-S",
		*from,
		"-E",
		*to,
	)
	var out strings.Builder
	sacct.Stdout = &out
	err := sacct.Run()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	result := make([][]string, 0)
	for scanner.Scan() {
		result = append(result, strings.Split(scanner.Text(), "|"))
	}
	return result
}

type gpu struct {
	kind  string
	count int
}

func sinfoInfo() map[string]gpu {
	sinfo := exec.Command("sinfo", "-o", "%n|%G", "--noheader")
	var out strings.Builder
	sinfo.Stdout = &out
	err := sinfo.Run()
	if err != nil {
		log.Fatal(err)
	}
	info := make(map[string]gpu)
	infoRe := regexp.MustCompile(`gpu:([^:]*):(\d+)`)
	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "|")
		if m := infoRe.FindStringSubmatch(fields[1]); m != nil {
			var g gpu
			g.kind = m[1]
			g.count, _ = strconv.Atoi(m[2])
			info[fields[0]] = g
		}
	}
	return info
}
