// Print sysinfo for individual nodes.  The sysinfo is a simple config.NodeConfigRecord, without any
// surrounding context.  All fields can be printed, we just print raw data except booleans are "yes"
// or "no".
//
// If there are logfiles present in the input then we use those as a transient cluster of sysinfo.
//
// TODO: On big systems it would clearly be interesting to filter by various criteria, eg memory
// size, number of cores or cards.

package nodes

import (
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	umaps "go-utils/maps"
	uslices "go-utils/slices"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	. "sonalyze/table"
)

type NodeCommand struct {
	DevArgs
	SourceArgs
	HostArgs
	VerboseArgs
	ConfigFileArgs
	FormatArgs
	Newest bool
}

var _ = (SimpleCommand)((*NodeCommand)(nil))

//go:embed summary.txt
var summary string

func (nc *NodeCommand) Summary() string {
	return summary
}

func (nc *NodeCommand) Add(fs *CLI) {
	nc.DevArgs.Add(fs)
	nc.SourceArgs.Add(fs)
	nc.HostArgs.Add(fs)
	nc.VerboseArgs.Add(fs)
	nc.ConfigFileArgs.Add(fs)
	nc.FormatArgs.Add(fs)

	fs.Group("printing")
	fs.BoolVar(&nc.Newest, "newest", false, "Print newest record per host only")
}

func (nc *NodeCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	x.Bool("newest", nc.Newest)
	return errors.Join(
		nc.DevArgs.ReifyForRemote(x),
		nc.SourceArgs.ReifyForRemote(x),
		nc.HostArgs.ReifyForRemote(x),
		nc.ConfigFileArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *NodeCommand) Validate() error {
	return errors.Join(
		nc.DevArgs.Validate(),
		nc.SourceArgs.Validate(),
		nc.HostArgs.Validate(),
		nc.VerboseArgs.Validate(),
		nc.ConfigFileArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, nodesDefaultFields, nodesFormatters, nodesAliases, DefaultFixed),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Processing

func (nc *NodeCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	var theLog db.SysinfoCluster
	var err error

	cfg, err := db.MaybeGetConfig(nc.ConfigFile())
	if err != nil {
		return err
	}

	filenameGlobber, err := hostglob.NewGlobber(
		true,
		uslices.Map(nc.HostArgs.Host, func(s string) string {
			return "sysinfo-" + s
		}),
	)
	if err != nil {
		return err
	}

	if len(nc.LogFiles) > 0 {
		theLog, err = db.OpenTransientSysinfoCluster(nc.LogFiles, cfg)
	} else {
		theLog, err = db.OpenPersistentCluster(nc.DataDir, cfg)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store: %v", err)
	}

	// Read and filter the raw sysinfo data.

	recordBlobs, dropped, err := theLog.ReadSysinfoData(
		nc.FromDate,
		nc.ToDate,
		filenameGlobber,
		nc.Verbose,
	)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}
	records := uslices.Catenate(recordBlobs)
	if nc.Verbose {
		Log.Infof("%d records read + %d dropped", len(records), dropped)
		UstrStats(stderr, false)
	}

	hostGlobber, recordFilter, err := nc.buildRecordFilter(nc.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to create record filter: %v", err)
	}

	records = slices.DeleteFunc(records, func(s *config.NodeConfigRecord) bool {
		if !hostGlobber.IsEmpty() && !hostGlobber.Match(s.Hostname) {
			return true
		}
		parsed, err := time.Parse(time.RFC3339, s.Timestamp)
		if err != nil {
			return true
		}
		t := parsed.Unix()
		return !(t >= recordFilter.From && t <= recordFilter.To)
	})

	if nc.Newest {
		newr := make(map[string]*config.NodeConfigRecord)
		for _, r := range records {
			if probe := newr[r.Hostname]; probe != nil {
				if r.Timestamp > probe.Timestamp {
					newr[r.Hostname] = r
				}
			} else {
				newr[r.Hostname] = r
			}
		}
		records = umaps.Values(newr)
	}

	// Sort by host name first and then by ascending time
	slices.SortFunc(records, func(a, b *config.NodeConfigRecord) int {
		if h := cmp.Compare(a.Hostname, b.Hostname); h != 0 {
			return h
		}
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		nodesFormatters,
		nc.PrintOpts,
		uslices.Map(records, func(x *config.NodeConfigRecord) any { return x }),
	)

	return nil
}

func (nc *NodeCommand) buildRecordFilter(
	verbose bool,
) (*hostglob.HostGlobber, *db.SampleFilter, error) {
	includeHosts, err := hostglob.NewGlobber(true, nc.HostArgs.Host)
	if err != nil {
		return nil, nil, err
	}

	haveFrom := nc.SourceArgs.HaveFrom
	haveTo := nc.SourceArgs.HaveTo
	var from int64 = 0
	if haveFrom {
		from = nc.SourceArgs.FromDate.Unix()
	}
	var to int64 = math.MaxInt64
	if haveTo {
		to = nc.SourceArgs.ToDate.Unix()
	}

	var recordFilter = &db.SampleFilter{
		IncludeHosts: includeHosts,
		From:         from,
		To:           to,
	}

	return includeHosts, recordFilter, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Printing
//
// Note cross-node is a config-level attribute, it does not appear in the raw sysinfo data, and so
// it is excluded here.

func (nc *NodeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(nc.Fmt, nodesHelp, nodesFormatters, nodesAliases, nodesDefaultFields)
}

const nodesHelp = `
node
  Extract information about individual nodes on the cluster from sysinfo and present
  them in primitive form.  Output records are sorted by node name.  The default
  format is 'fixed'.
`

const v0NodesDefaultFields = "host,cores,mem,gpus,gpumem,desc"
const v1NodesDefaultFields = "Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description"
const nodesDefaultFields = v0NodesDefaultFields

// MT: Constant after initialization; immutable
var nodesAliases = map[string][]string{
	"default":   strings.Split(nodesDefaultFields, ","),
	"v0default": strings.Split(v0NodesDefaultFields, ","),
	"v1default": strings.Split(v1NodesDefaultFields, ","),
}

//go:generate sonalyze-table -o node-table.go nodes.go

//PACKAGE nodes
//IMPORT "go-utils/config"

//TABLE nodes config.NodeConfigRecord
//
// Timestamp   string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
// Hostname    string desc:"Name that host is known by on the cluster" alias:"host"
// Description string desc:"End-user description, not parseable" alias:"desc"
// CpuCores    int    desc:"Total number of cores x threads" alias:"cores"
// MemGB       int    desc:"GB of installed main RAM" alias:"mem"
// GpuCards    int    desc:"Number of installed cards" alias:"gpus"
// GpuMemGB    int    desc:"Total GPU memory across all cards" alias:"gpumem"
// GpuMemPct   bool   desc:"True if GPUs report accurate memory usage in percent" alias:"gpumempct"

// This should be generated from the above, like formatters
//
// Probably we can make to with fewer functions cf x86 fp instructions

var nodesQuery = map[string]QueryField[config.NodeConfigRecord] {
	"MemGB": {
		Convert: CvtToInt,
		FilterEqual: func(d *config.NodeConfigRecord, v any) bool {
			return d.MemGB == v.(int)
		},
		FilterLess: func(d *config.NodeConfigRecord, v any) bool {
			return d.MemGB < v.(int)
		},
		FilterLessOrEqual: func(d *config.NodeConfigRecord, v any) bool {
			return d.MemGB <= v.(int)
		},
		FilterGreater: func(d *config.NodeConfigRecord, v any) bool {
			return d.MemGB > v.(int)
		},
		FilterGreaterOrEqual: func(d *config.NodeConfigRecord, v any) bool {
			return d.MemGB >= v.(int)
		},
	},
}

func CvtToInt(s string) (any, error) {
	panic("NYI")
}
