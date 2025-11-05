package parse

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	"go-utils/config"
	uslices "go-utils/slices"
	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db"
	"sonalyze/db/repr"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o parse-table.go parse.go

/*TABLE parse

package parse

import "sonalyze/data/sample"

%%

FIELDS sample.Sample

 # TODO: IMPROVEME: The use of utc for "localtime" is a bug that comes from the Rust code.

 Version    Ustr                desc:"Semver string (MAJOR.MINOR.BUGFIX)" alias:"version,v"
 Timestamp  DateTimeValue       desc:"Timestamp of record " alias:"localtime"
 time       IsoDateTimeValue    desc:"Timestamp of record" field:"Timestamp"
 Hostname   Ustr                desc:"Host name (FQDN)" alias:"host"
 Cores      uint32              desc:"Total number of cores (including hyperthreads)" alias:"cores"
 Threads    uint32              desc:"Number of threads active" alias:"threads"
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
  all       version,localtime,host,cores,threads,memtotal,user,pid,job,cmd,cpu_pct,mem_gb,res_gb,gpus,gpu_pct,gpumem_pct,gpumem_gb,gpu_status,cputime_sec,rolledup,cpu_util_pct
  All       Version,Timestamp,Hostname,Cores,Threads,MemtotalKB,User,Pid,Ppid,Job,Cmd,CpuPct,CpuKB,RssAnonKB,Gpus,GpuPct,GpuMemPct,GpuKB,GpuFail,CpuTimeSec,Rolledup,CpuUtilPct
  roundtrip v,time,host,cores,user,job,pid,cmd,cpu%,cpukib,gpus,gpu%,gpumem%,gpukib,gpufail,cputime_sec,rolledup

DEFAULTS default

ELBAT*/

type ParseCommand struct /* implements SampleAnalysisCommand */ {
	SampleAnalysisArgs
	FormatArgs

	MergeByHostAndJob bool
	MergeByJob        bool
	Clean             bool
	LastN             uint
}

var _ = SampleAnalysisCommand((*ParseCommand)(nil))

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

	fs.Group("printing")
	fs.UintVar(&pc.LastN, "last", 0,
		"Show only the most recent `n` records for merged and cleaned jobs, 0=all")
}

func (pc *ParseCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		pc.SampleAnalysisArgs.ReifyForRemote(x),
		pc.FormatArgs.ReifyForRemote(x),
	)
	x.Bool("merge-by-host-and-job", pc.MergeByHostAndJob)
	x.Bool("merge-by-job", pc.MergeByJob)
	x.Bool("clean", pc.Clean)
	if pc.LastN > 0 {
		x.Uint("last", pc.LastN)
	}
	return e1
}

func (pc *ParseCommand) Validate() error {
	var e1 error
	if pc.LastN > 0 && (!pc.MergeByHostAndJob && !pc.MergeByJob && !pc.Clean) {
		e1 = errors.New("-last requires a merge/clean option")
	}
	return errors.Join(
		e1,
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

func (pc *ParseCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	theDb db.SampleDataProvider,
	filter sample.QueryFilter,
	hosts *Hosts,
	recordFilter *sample.SampleFilter,
) error {
	var mergedSamples []sample.SampleStream
	var samples sample.SampleStream

	if pc.Clean || pc.MergeByJob || pc.MergeByHostAndJob {
		streams, bounds, read, dropped, err :=
			sample.ReadSampleStreamsAndMaybeBounds(
				theDb,
				filter.FromDate,
				filter.ToDate,
				hosts,
				recordFilter,
				pc.MergeByJob,
				pc.Verbose,
			)
		if err != nil {
			return fmt.Errorf("Failed to read log records: %v", err)
		}
		if pc.Verbose {
			Log.Infof("%d records read + %d dropped\n", read, dropped)
			UstrStats(out, false)
		}
		if pc.Verbose {
			Log.Infof("Streams constructed by postprocessing: %d", len(streams))
			numSamples := 0
			for _, stream := range streams {
				numSamples += len(*stream)
			}
			Log.Infof("Samples retained after filtering: %d", numSamples)
		}

		switch {
		case pc.Clean:
			mergedSamples = make([]sample.SampleStream, 0, len(streams))
			for _, v := range streams {
				mergedSamples = append(mergedSamples, *v)
			}

		case pc.MergeByJob:
			mergedSamples, _ = sample.MergeByJob(streams, bounds)

		case pc.MergeByHostAndJob:
			mergedSamples = sample.MergeByHostAndJob(streams)
		}
	} else {
		// Bypass postprocessing to get the expected raw values.
		recordBlobs, dropped, err := theDb.ReadProcessSamples(pc.FromDate, pc.ToDate, hosts, pc.Verbose)
		if err != nil {
			return fmt.Errorf("Failed to read log records: %v", err)
		}
		if pc.Verbose {
			Log.Infof("%d record blobs read + %d dropped\n", len(recordBlobs), dropped)
			UstrStats(out, false)
		}

		// Simulate the normal pipeline, the recordFilter application is expected by the user.
		mapped := make([]sample.Sample, 0)
		for _, records := range recordBlobs {
			mapped = append(mapped, uslices.FilterMap(
				records,
				sample.InstantiateSampleFilter(recordFilter),
				func(r *repr.Sample) sample.Sample {
					return sample.Sample{Sample: r}
				},
			)...)
		}
		samples = sample.SampleStream(mapped)
	}

	var queryNeg func(sample.Sample) bool
	if pc.ParsedQuery != nil {
		var err error
		queryNeg, err = CompileQueryNeg(parseFormatters, parsePredicates, pc.ParsedQuery)
		if err != nil {
			return fmt.Errorf("Could not compile query: %v", err)
		}
	}

	if mergedSamples != nil {
		if pc.LastN > 0 {
			// The samples are already sorted by time, so pick the last of each run.
			for i := range mergedSamples {
				if len(mergedSamples[i]) > int(pc.LastN) {
					mergedSamples[i] = mergedSamples[i][len(mergedSamples[i])-int(pc.LastN):]
				}
			}
		}

		// All elements that are part of the InputStreamKey must be part of the sort key here.
		slices.SortStableFunc(mergedSamples, func(a, b sample.SampleStream) int {
			c := cmp.Compare(a[0].Hostname.String(), b[0].Hostname.String())
			if c == 0 {
				c = cmp.Compare(a[0].Timestamp, b[0].Timestamp)
				if c == 0 {
					c = cmp.Compare(a[0].Job, b[0].Job)
					if c == 0 {
						c = cmp.Compare(a[0].Cmd.String(), b[0].Cmd.String())
					}
				}
			}
			return c
		})
		for _, stream := range mergedSamples {
			xs := stream
			if queryNeg != nil {
				xs = slices.DeleteFunc(xs, queryNeg)
			}
			if pc.PrintOpts.Fixed || pc.PrintOpts.Separator {
				fmt.Fprintln(out, "*")
			}
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
