// Create a cluster config file for a cluster based on raw data collected by `sonar sysinfo` or
// extracted from the Slurm command `sinfo`.
//
// The data source is selected using either `-data-dir` (for a cluster data store containing the
// sysinfo files), `-sinfo` (to run sinfo directly) or `-sinfo-input` (to read output coming from
// sinfo from a file, that may have been altered after generation).
//
// The optional `-from` and `-to` arguments to `-data-dir` delimits the time period in which to
// scan for sysinfo records.
//
// Cluster values for the output are supplied by `-name`, `-aliases`, `-desc`, and `-exclude-users`.
//
// Default values for individual nodes are supplied by an optional file of background information,
// specified with `-background-file`.
//
// Run with -h for more information.
//
// Examples.
//
// Generating from sysinfo records from the last week, without background:
//
//   make-cluster-config \
//      -data-dir ~/sonar/data/mlx.hpc.uio.no \
//      -from 1w \
//      -name mlx.hpc.uio.no \
//      -aliases ml,mlx \
//      -desc "UiO Machine Learning nodes" \
//      -exclude-users tmux,zabbix,root,sshd
//
// Generating from current sinfo, with background:
//
//   make-cluster-config \
//      -sinfo \
//      -background-file ~/.../misc/fox.educloud.no/fox.educloud.no-background.json \
//      -name fox.educloud.no \
//      -aliases fox \
//      -desc "UiO 'Fox' supercomputer" \
//      -exclude-users tmux,zabbix,root,sshd
//
// Formats.
//
// `sonar sysinfo` will probe a node's configuration and emit a JSON object with information about
// the node.  The data format is defined by Sonar: https://github.com/NordicHPC/sonar, look in
// src/sysinfo.rs.
//
// The cluster configuration file format is the v2 format defined in ../go-utils/config/config.go.
//
// The background file format is the v1 format defined in ../go-utils/config/config.go, but fields
// may be missing, see ../go-utils/config/background.go.  Nonstandard fields may also be present,
// in particular:
//  - "gpu" may carry the GPU model, not reported by sinfo
//  - "host-suffix" may carry a suffix for the node name, not reported by sinfo

package main

import (
	"flag"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/maps"
	ut "go-utils/time"
)

// Command-line parameters
var (
	// One data source must be selected

	// Data source is a cluster data store, with optional timestamps from between which to select
	// data; the default is to select data from the last week
	dataDir       string
	from          time.Time
	to            time.Time

	// Data source is an sinfo subprocess that is run by this program
	sinfo         bool

	// Data source is a file containing the output from `sinfo -a -o '%n/%f/%m/%X/%Y/%Z'`, possibly
	// postprocessed; use - for stdin
	sinfoInput     string

	// The canonical name of the cluster (required)
	clusterName   string

	// Any aliases for the cluster (optional)
	aliases       []string

	// Human-readable description (required)
	clusterDesc   string

	// Users that should be automatically be excluded (is this used at all?) (optional)
	excludeUsers  []string

	// The background file supplies information that is not present in the command line parameters
	// or in the data (optional)
	backgroundFile string
)

func main() {
	parseFlags()
	bg := readBackground()

	// Read data and compute some format-specific values from the background data.  Common fixups
	// are applied below.
	var nodes map[string]*config.NodeConfigRecord
	if dataDir != "" {
		nodes = readNodesFromSysinfo(bg)
	} else {
		nodes = readSinfo(bg)
	}

	// For the returned nodes, apply information from the background for missing fields.
	for _, node := range nodes {
		// Apply background information.
		if bginfo := bg[node.Hostname]; bginfo != nil {
			// node.Timestamp is always valid
			// node.Hostname is always valid
			if node.Description == "" {
				node.Description = bginfo.Description
			}
			node.CrossNodeJobs = bginfo.CrossNodeJobs
			if node.CpuCores == 0 {
				node.CpuCores = bginfo.CpuCores
			}
			if node.MemGB == 0 {
				node.MemGB = bginfo.MemGB
			}
			if node.GpuCards == 0 {
				node.GpuCards = bginfo.GpuCards
			}
			if node.GpuMemGB == 0 {
				node.GpuMemGB = bginfo.GpuMemGB
			}
			if !node.GpuMemPct {
				node.GpuMemPct = bginfo.GpuMemPct
			}
		}
	}

	// Add missing nodes, these must supply all their data and we don't check.
	for bgName, bgInfo := range bg {
		if nodes[bgName] != nil {
			continue
		}
		if bgInfo.CpuCores == 0 || bgInfo.MemGB == 0 {
			continue
		}
		for _, m := range bgInfo.Metadata {
			switch m.Key {
			case "host-suffix":
				bgInfo.Hostname += m.Value
			}
		}
		bgInfo.Metadata = nil
		nodes[bgName] = bgInfo
	}

	cc := config.NewClusterConfig(
		2,
		clusterName,
		clusterDesc,
		aliases,
		excludeUsers,
		maps.Values(nodes),
	)

	err := config.WriteConfigTo(os.Stdout, cc)
	if err != nil {
		log.Fatal(err)
	}
}

func readBackground() map[string]*config.NodeConfigRecord {
	var bg map[string]*config.NodeConfigRecord
	if backgroundFile != "" {
		var err error
		bg, err = config.ReadBackgroundFile(backgroundFile)
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
	fromStr := flag.String("from", "", "Start `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	toStr := flag.String("to", "", "End `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	backgroundStr := flag.String("background-file", "", "Find background data in `filename`")
	nameStr := flag.String("name", "", "Canonical cluster `name` (required)")
	descStr := flag.String("desc", "", "Cluster `description` (required)")
	aliasStr := flag.String("aliases", "", "Cluster `alias,alias,...`")
	excludeStr := flag.String("exclude-users", "", "Exclude processes from `user,user,...`")
	sinfoFlag := flag.Bool("sinfo", false, "Run sinfo to obtain cluster information")
	sinfoInputStr := flag.String("sinfo-input", "", "Read sinfo information from `file` (- for stdin)")
	flag.Parse()

	var err error

	// Input source
	var sources int
	if *dataDirStr != "" {
		sources++
	}
	if *sinfoFlag {
		sources++
	}
	if *sinfoInputStr != "" {
		sources++
	}
	if sources != 1 {
		log.Fatal("Exactly one of -data-dir, -sinfo, and -sinfo-input must be selected")
	}
	if *dataDirStr != "" {
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
	} else {
		if *fromStr != "" || *toStr != "" {
			log.Fatal("-from and -to are only allowed with -data-dir")
		}
	}
	if *sinfoInputStr != "" {
		if *sinfoInputStr != "-" {
			sinfoInput = path.Clean(*sinfoInputStr)
		} else {
			sinfoInput = "-"
		}
	}
	sinfo = *sinfoFlag

	if *backgroundStr != "" {
		backgroundFile = path.Clean(*backgroundStr)
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
