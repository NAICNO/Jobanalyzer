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

//go:generate ../../../generate-table/generate-table -o node-table.go nodes.go

/*TABLE node

package nodes

import (
	"go-utils/config"
	. "sonalyze/table"
)

%%

FIELDS *config.NodeConfigRecord

 # Note the CrossNodeJobs field is a config-level attribute, it does not appear in the raw sysinfo
 # data, and so it is not included here.

 Timestamp   string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
 Hostname    string desc:"Name that host is known by on the cluster" alias:"host"
 Description string desc:"End-user description, not parseable" alias:"desc"
 CpuCores    int    desc:"Total number of cores x threads" alias:"cores"
 MemGB       int    desc:"GB of installed main RAM" alias:"mem"
 GpuCards    int    desc:"Number of installed cards" alias:"gpus"
 GpuMemGB    int    desc:"Total GPU memory across all cards" alias:"gpumem"
 GpuMemPct   bool   desc:"True if GPUs report accurate memory usage in percent" alias:"gpumempct"

HELP NodeCommand

  Extract information about individual nodes on the cluster from sysinfo and present
  them in primitive form.  Output records are sorted by node name.  The default
  format is 'fixed'.

ALIASES

  default  host,cores,mem,gpus,gpumem,desc
  Default  Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description
  all      timestamp,host,desc,cores,mem,gpus,gpumem,gpumempct
  All      Timestamp,Hostname,Description,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct

DEFAULTS default

ELBAT*/

type NodeCommand struct {
	DevArgs
	SourceArgs
	QueryArgs
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
	nc.QueryArgs.Add(fs)
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
		nc.QueryArgs.ReifyForRemote(x),
		nc.HostArgs.ReifyForRemote(x),
		nc.ConfigFileArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *NodeCommand) Validate() error {
	return errors.Join(
		nc.DevArgs.Validate(),
		nc.SourceArgs.Validate(),
		nc.QueryArgs.Validate(),
		nc.HostArgs.Validate(),
		nc.VerboseArgs.Validate(),
		nc.ConfigFileArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, nodeDefaultFields, nodeFormatters, nodeAliases, DefaultFixed),
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

	hostGlobber, recordFilter, query, err := nc.buildRecordFilter(nc.Verbose)
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
		if !(t >= recordFilter.From && t <= recordFilter.To) {
			return true
		}
		if query != nil && !query(s) {
			return true
		}
		return false
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
		nodeFormatters,
		nc.PrintOpts,
		records,
	)

	return nil
}

func (nc *NodeCommand) buildRecordFilter(
	verbose bool,
) (*hostglob.HostGlobber, *db.SampleFilter, func(*config.NodeConfigRecord) bool, error) {
	includeHosts, err := hostglob.NewGlobber(true, nc.HostArgs.Host)
	if err != nil {
		return nil, nil, nil, err
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

	var query func(*config.NodeConfigRecord) bool
	if nc.ParsedQuery != nil {
		c, err := CompileQuery(nodePredicates, nc.ParsedQuery)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Could not compile query: %v", err)
		}
		query = c
	}

	return includeHosts, recordFilter, query, nil
}

// Here:
//
// * If Convert is nil then type must be string and we just use the input string.
// * Compare must not be nil, it extracts the field and then does a straight value
//   comparison

var nodePredicates = map[string]Predicate[*config.NodeConfigRecord]{
	"Timestamp": Predicate[*config.NodeConfigRecord]{
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.Timestamp, v.(string))
		},
	},
	"Hostname": Predicate[*config.NodeConfigRecord]{
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.Hostname, v.(string))
		},
	},
	"Description": Predicate[*config.NodeConfigRecord]{
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.Description, v.(string))
		},
	},
	"CpuCores": Predicate[*config.NodeConfigRecord]{
		Convert: CvtString2Int,
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.CpuCores, v.(int))
		},
	},
	"MemGB": Predicate[*config.NodeConfigRecord]{
		Convert: CvtString2Int,
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.MemGB, v.(int))
		},
	},
	"GpuCards": Predicate[*config.NodeConfigRecord]{
		Convert: CvtString2Int,
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.GpuCards, v.(int))
		},
	},
	"GpuMemGB": Predicate[*config.NodeConfigRecord]{
		Convert: CvtString2Int,
		Compare: func(d *config.NodeConfigRecord, v any) int {
			return cmp.Compare(d.GpuMemGB, v.(int))
		},
	},
	"GpuMemPct": Predicate[*config.NodeConfigRecord]{
		Convert: CvtString2Bool,
		Compare: func(d *config.NodeConfigRecord, v any) int {
			var x, y int
			if d.GpuMemPct {
				x = 1
			}
			if v.(bool) {
				y = 1
			}
			return x - y
		},
	},
}
