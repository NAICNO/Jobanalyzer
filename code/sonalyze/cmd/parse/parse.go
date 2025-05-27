package parse

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	"go-utils/config"
	"go-utils/maps"
	uslices "go-utils/slices"
	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o parse-table.go parse.go

/*TABLE parse

package parse

import "sonalyze/sonarlog"

%%

FIELDS sonarlog.Sample

 # TODO: IMPROVEME: The use of utc for "localtime" is a bug that comes from the Rust code.

 Version    Ustr                desc:"Semver string (MAJOR.MINOR.BUGFIX)" alias:"version,v"
 Timestamp  DateTimeValue       desc:"Timestamp of record " alias:"localtime"
 time       IsoDateTimeValue    desc:"Timestamp of record" field:"Timestamp"
 Hostname   Ustr                desc:"Host name (FQDN)" alias:"host"
 Cores      uint32              desc:"Total number of cores (including hyperthreads)" alias:"cores"
 MemtotalKB uint64              desc:"Installed main memory"
 memtotal   U64Div1M            desc:"Installed main memory (GB)" field:"MemtotalKB"
 User       Ustr                desc:"Username of process owner" alias:"user"
 Pid        uint32              desc:"Process ID" alias:"pid"
 Ppid       uint32              desc:"Process parent ID" alias:"ppid"
 Job        uint32              desc:"Job ID" alias:"job"
 Cmd        Ustr                desc:"Command name" alias:"cmd"
 CpuPct     float32             desc:"cpu% reading (CONSULT DOCUMENTATION)" alias:"cpu_pct,cpu%"
 CpuKB      uint64              desc:"Virtual memory reading" alias:"cpukib"
 mem_gb     U64Div1M            desc:"Virtual memory reading" field:"CpuKB"
 RssAnonKB  uint64              desc:"RssAnon reading"
 res_gb     U64Div1M            desc:"RssAnon reading" field:"RssAnonKB"
 Gpus       gpuset.GpuSet       desc:"GPU set (`none`,`unknown`,list)" alias:"gpus"
 GpuPct     float32             desc:"GPU utilization reading" alias:"gpu_pct,gpu%"
 GpuMemPct  float32             desc:"GPU memory percentage reading" alias:"gpumem_pct,gpumem%"
 GpuKB      uint64              desc:"GPU memory utilization reading" alias:"gpukib"
 gpumem_gb  U64Div1M            desc:"GPU memory utilization reading" field:"GpuKB"
 GpuFail    uint8               desc:"GPU status flag (0=ok, 1=error state)" alias:"gpu_status,gpufail"
 CpuTimeSec uint64              desc:"CPU time since last reading (seconds, CONSULT DOCUMENTATION)" \
                                alias:"cputime_sec"
 Rolledup   uint32              desc:"Number of rolled-up processes, minus 1" alias:"rolledup"
 Flags      uint8               desc:"Bit vector of flags, UTSL"
 CpuUtilPct float32             desc:"CPU utilization since last reading (percent, CONSULT DOCUMENTATION)" \
                                alias:"cpu_util_pct"

SUMMARY ParseCommand

Export sample data in various formats, after optional preprocessing.

This facility is mostly for debugging and experimentation, as the data
volume is typically significant and the data are not necessarily
postprocessed in a way useful to the consumer.

The -merge and -clean options perform some postprocessing, but you need to
know what you're looking at to find these useful.

HELP ParseCommand

  Read raw Sonar data and present it in whole or part.  Default output format
  is 'csv'.

ALIASES

  default   job,user,cmd
  Default   Job,User,Cmd
  all       version,localtime,host,cores,memtotal,user,pid,job,cmd,cpu_pct,mem_gb,res_gb,gpus,gpu_pct,gpumem_pct,gpumem_gb,gpu_status,cputime_sec,rolledup,cpu_util_pct
  All       Version,Timestamp,Hostname,Cores,MemtotalKB,User,Pid,Ppid,Job,Cmd,CpuPct,CpuKB,RssAnonKB,Gpus,GpuPct,GpuMemPct,GpuKB,GpuFail,CpuTimeSec,Rolledup,CpuUtilPct
  roundtrip v,time,host,cores,user,job,pid,cmd,cpu%,cpukib,gpus,gpu%,gpumem%,gpukib,gpufail,cputime_sec,rolledup

DEFAULTS default

ELBAT*/

type ParseCommand struct /* implements SampleAnalysisCommand */ {
	SampleAnalysisArgs
	FormatArgs

	MergeByHostAndJob bool
	MergeByJob        bool
	Clean             bool
}

var _ SampleAnalysisCommand = (*ParseCommand)(nil)

func (pc *ParseCommand) Add(fs *CLI) {
	pc.SampleAnalysisArgs.Add(fs)
	pc.FormatArgs.Add(fs)

	fs.Group("aggregation")
	fs.BoolVar(&pc.MergeByHostAndJob, "merge-by-host-and-job", false,
		"Merge streams that have the same host and job ID")
	fs.BoolVar(&pc.MergeByJob, "merge-by-job", false,
		"Merge streams that have the same job ID, across hosts")
	fs.BoolVar(&pc.Clean, "clean", false,
		"Clean the job but perform no merging")
}

func (pc *ParseCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		pc.SampleAnalysisArgs.ReifyForRemote(x),
		pc.FormatArgs.ReifyForRemote(x),
	)
	x.Bool("merge-by-host-and-job", pc.MergeByHostAndJob)
	x.Bool("merge-by-job", pc.MergeByJob)
	x.Bool("clean", pc.Clean)
	return e1
}

func (pc *ParseCommand) Validate() error {
	return errors.Join(
		pc.SampleAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&pc.FormatArgs, parseDefaultFields, parseFormatters, parseAliases, DefaultCsv),
	)
}

func (pc *ParseCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	// `parse` implies `--user=-` b/c we're interested in raw data.
	allUsers = true
	skipSystemUsers = false
	excludeSystemCommands = false
	excludeHeartbeat = false
	return
}

func (pc *ParseCommand) ConfigFile() string {
	// `sonalyze parse` accepts no config file
	return ""
}

func (pc *ParseCommand) NeedsBounds() bool {
	return pc.MergeByJob
}

func (pc *ParseCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	cluster db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds, // for pc.MergeByJob only
	hostGlobber *Hosts,
	recordFilter *db.SampleFilter,
) error {
	var mergedSamples []*sonarlog.SampleStream
	var samples sonarlog.SampleStream
	switch {
	default:
		// Reread the data, bypassing all postprocessing, to get the expected raw values.  If it's
		// expensive then so be it - this is special-case code usually used for limited testing, not
		// something you'd use for analysis.
		recordBlobs, _, err := cluster.ReadSamples(pc.FromDate, pc.ToDate, hostGlobber, pc.Verbose)
		if err != nil {
			return err
		}

		// Simulate the normal pipeline, the recordFilter application is expected by the user.
		mapped := make([]sonarlog.Sample, 0)
		for _, records := range recordBlobs {
			mapped = append(mapped, uslices.FilterMap(
				records,
				db.InstantiateSampleFilter(recordFilter),
				func(r *db.Sample) sonarlog.Sample {
					return sonarlog.Sample{Sample: r}
				},
			)...)
		}
		samples = sonarlog.SampleStream(mapped)

	case pc.Clean:
		mergedSamples = maps.Values(streams)

	case pc.MergeByJob:
		mergedSamples, _ = sonarlog.MergeByJob(streams, bounds)

	case pc.MergeByHostAndJob:
		mergedSamples = sonarlog.MergeByHostAndJob(streams)
	}

	var queryNeg func(sonarlog.Sample) bool
	if pc.ParsedQuery != nil {
		var err error
		queryNeg, err = CompileQueryNeg(parseFormatters, parsePredicates, pc.ParsedQuery)
		if err != nil {
			return fmt.Errorf("Could not compile query: %v", err)
		}
	}

	if mergedSamples != nil {
		// All elements that are part of the InputStreamKey must be part of the sort key here.
		slices.SortStableFunc(mergedSamples, func(a, b *sonarlog.SampleStream) int {
			c := cmp.Compare((*a)[0].Hostname.String(), (*b)[0].Hostname.String())
			if c == 0 {
				c = cmp.Compare((*a)[0].Timestamp, (*b)[0].Timestamp)
				if c == 0 {
					c = cmp.Compare((*a)[0].Job, (*b)[0].Job)
					if c == 0 {
						c = cmp.Compare((*a)[0].Cmd.String(), (*b)[0].Cmd.String())
					}
				}
			}
			return c
		})
		for _, stream := range mergedSamples {
			xs := *stream
			if queryNeg != nil {
				xs = slices.DeleteFunc(xs, queryNeg)
			}
			fmt.Fprintln(out, "*")
			FormatData(
				out,
				pc.PrintFields,
				parseFormatters,
				pc.PrintOpts,
				xs,
			)
		}
	} else {
		if queryNeg != nil {
			samples = slices.DeleteFunc(samples, queryNeg)
		}
		FormatData(
			out,
			pc.PrintFields,
			parseFormatters,
			pc.PrintOpts,
			samples,
		)
	}
	return nil
}
