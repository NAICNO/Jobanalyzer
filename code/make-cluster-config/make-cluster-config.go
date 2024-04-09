// This creates a cluster config file for a cluster based on raw data collected by `sysinfo` or
// (equivalently) `sonar sysinfo`.
//
// `sysinfo` will probe a node's configuration and emit a JSON object with information about the
// node.  The data format is defined jointly by the `sysinfo` program in this repo and the `sysinfo`
// verb in Sonar.
//
// make-cluster-config collects sysinfo for nodes and will create a configuration file for the
// cluster, filling in any missing data from the `background` file, if provided.
//
// The cluster configuration file format is defined jointly by a number of programs, chiefly by
// ../rustutils/src/configs.rs and ../go-utils/config/config.go.  make-cluster-config currently only
// produces the v2 format but will eventually have to produce a new format with a per-host timeline.
//
// The background file, if present, also must currently be v2 but may eventually have to be v3
// (since the background also changes over time).
//
// Typical usage (run with -h for full argument list):
//
// make-cluster-config \
//    -data-dir ~/sonar/data/mlx.hpc.uio.no \
//    -name mlx.hpc.uio.no \
//    -desc "UiO ML nodes" \
//    -from 2023-01-01 \
//    -background ~/.../misc/mlx.hpc.uio.no/mlx.hpc.uio.no-background.json

package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/filesys"
	ut "go-utils/time"
)

// Command-line parameters
var (
	dataDir       string
	from          time.Time
	to            time.Time
	backgroundDir string
	clusterName   string
	clusterDesc   string
	excludeUsers  []string
	aliases       []string
)

func main() {
	parseFlags()

	cc := config.NewClusterConfig()
	cc.Version = 2
	cc.Name = clusterName
	cc.Description = clusterDesc
	cc.Aliases = aliases
	cc.ExcludeUser = excludeUsers

	bg := readBackground()
	info := readSysinfo()

	for _, infos := range info {
		// For the v2 format we can only have one timestamp, so take the latest always.
		var latest *config.NodeConfigRecord
		for _, info := range infos {
			if latest == nil || info.Timestamp > latest.Timestamp {
				latest = info
			}
		}

		if latest == nil {
			// Wow, weird
			continue
		}

		// Apply background information.
		if bginfo := bg[latest.Hostname]; bginfo != nil {
			// latest.Timestamp is always valid
			// latest.Hostname is always valid
			if latest.Description == "" {
				latest.Description = bginfo.Description
			}
			latest.CrossNodeJobs = bginfo.CrossNodeJobs
			if latest.CpuCores == 0 {
				latest.CpuCores = bginfo.CpuCores
			}
			if latest.MemGB == 0 {
				latest.MemGB = bginfo.MemGB
			}
			if latest.GpuCards == 0 {
				latest.GpuCards = bginfo.GpuCards
			}
			if latest.GpuMemGB == 0 {
				latest.GpuMemGB = bginfo.GpuMemGB
			}
			if !latest.GpuMemPct {
				latest.GpuMemPct = bginfo.GpuMemPct
			}
		}

		cc.Insert(latest)
	}

	// Add missing hosts
	for host, bginfo := range bg {
		if cc.LookupHost(host) == nil {
			if bginfo.CpuCores == 0 || bginfo.MemGB == 0 {
				continue
			}
			for _, m := range bginfo.Metadata {
				switch m.Key {
				case "host-suffix":
					bginfo.Hostname += m.Value
				}
			}
			bginfo.Metadata = nil
			cc.Insert(bginfo)
		}
	}

	err := config.WriteConfigTo(os.Stdout, cc)
	if err != nil {
		log.Fatal(err)
	}
}

func readSysinfo() map[string][]*config.NodeConfigRecord {
	files, err := filesys.EnumerateFiles(dataDir, from, to, "sysinfo-*.json")
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatalf("No sysinfo files found in %s", dataDir)
	}
	info := make(map[string][]*config.NodeConfigRecord)
	for _, fn := range files {
		input, err := os.Open(path.Join(dataDir, fn))
		if err != nil {
			log.Fatal(err)
		}
		// The sysinfo file has zero or more records in a row, but not wrapped in an array or
		// separated by anything more than space.  Thus, use a decoder to read the successive
		// records.
		dec := json.NewDecoder(input)
		for {
			var d config.NodeConfigRecord
			err := dec.Decode(&d)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			info[d.Hostname] = append(info[d.Hostname], &d)
		}
	}
	return info
}

func readBackground() map[string]*config.NodeConfigRecord {
	var bg map[string]*config.NodeConfigRecord
	if backgroundDir != "" {
		var err error
		bg, err = config.ReadBackgroundFile(backgroundDir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		bg = make(map[string]*config.NodeConfigRecord)
	}
	return bg
}

func parseFlags() {
	dataDirStr := flag.String("data-dir", "", "Find sysinfo files in tree below `directory` (required)")
	fromStr := flag.String("from", "1d", "Start `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	toStr := flag.String("to", "", "End `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	backgroundStr := flag.String("background", "", "Find background data in `filename`")
	nameStr := flag.String("name", "", "Canonical cluster `name` (required)")
	descStr := flag.String("desc", "", "Cluster `description` (required)")
	aliasStr := flag.String("alias", "", "Cluster `alias,alias,...`")
	excludeStr := flag.String("exclude", "", "Exclude processes from `user,user,...`")
	flag.Parse()

	var err error
	if *dataDirStr == "" {
		log.Fatal("-data-dir is required")
	}
	dataDir = path.Clean(*dataDirStr)
	if *fromStr != "" {
		from, err = ut.ParseRelativeDate(*fromStr)
		if err != nil {
			log.Fatalf("-from: %v", err)
		}
	} else {
		from = ut.PreviousDay(time.Now())
	}
	if *toStr != "" {
		to, err = ut.ParseRelativeDate(*toStr)
		if err != nil {
			log.Fatalf("-to: %v", err)
		}
	} else {
		to = ut.NextDay(time.Now())
	}
	if *backgroundStr != "" {
		backgroundDir = path.Clean(*backgroundStr)
	}
	if *nameStr == "" {
		log.Fatal("-name is required")
	}
	clusterName = *nameStr
	if *descStr == "" {
		log.Fatal("-desc is required")
	}
	clusterDesc = *descStr
	if *excludeStr != "" {
		excludeUsers = strings.Split(*excludeStr, ",")
	}
	if *aliasStr != "" {
		aliases = strings.Split(*aliasStr, ",")
	}
}
