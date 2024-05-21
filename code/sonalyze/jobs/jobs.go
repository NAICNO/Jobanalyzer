// Compute jobs aggregates from a set of log entries.

package jobs

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strconv"

	. "sonalyze/command"
)

type uintArg struct {
	text        string // help text
	initial     uint   // default value
	name        string // canonical option name
	aggregateIx int    // index in the data generated by aggregation, see perform.go
	relative    bool   // true if config-relative
}

// This is "so large that even extreme outliers resulting from bugs will not reach this value" and
// is used as the "initial" value for max attributes.  When the property has this value, no filter
// will be generated.
const bigValue = 100000000

var uintArgs = []uintArg{
	uintArg{
		"Select only jobs with at least this many samples [default: 2]",
		2,
		"min-samples",
		-1,
		false,
	},
	uintArg{
		"Select only jobs with at least this much average CPU use (100=1 full CPU)",
		0,
		"min-cpu-avg",
		kCpuPctAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak CPU use (100=1 full CPU)",
		0,
		"min-cpu-peak",
		kCpuPctPeak,
		false,
	},
	uintArg{
		"Select only jobs with at most this much average CPU use (100=1 full CPU)",
		bigValue,
		"max-cpu-avg",
		kCpuPctAvg,
		false,
	},
	uintArg{
		"Select only jobs with at most this much peak CPU use (100=1 full CPU)",
		bigValue,
		"max-cpu-peak",
		kCpuPctPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average CPU use (100=all cpus)",
		0,
		"min-rcpu-avg",
		kRcpuPctAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak CPU use (100=all cpus)",
		0,
		"min-rcpu-peak",
		kRcpuPctPeak,
		true,
	},
	uintArg{
		"Select only jobs with at most this much relative average CPU use (100=all cpus)",
		bigValue,
		"max-rcpu-avg",
		kRcpuPctAvg,
		true,
	},
	uintArg{
		"Select only jobs with at most this much relative peak CPU use (100=all cpus)",
		bigValue,
		"max-rcpu-peak",
		kRcpuPctPeak,
		true,
	},
	uintArg{
		"Select only jobs with at least this much average virtual memory use (GB)",
		0,
		"min-mem-avg",
		kCpuGibAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak virtual memory use (GB)",
		0,
		"min-mem-peak",
		kCpuGibPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average virtual memory use (100=all memory)",
		0,
		"min-rmem-avg",
		kRcpuGibAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak virtual memory use (100=all memory)",
		0,
		"min-rmem-peak",
		kRcpuGibAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much average resident memory use (GB)",
		0,
		"min-res-avg",
		kRssAnonGibAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak resident memory use (GB)",
		0,
		"min-res-peak",
		kRssAnonGibPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average resident memory use (100=all memory)",
		0,
		"min-rres-avg",
		kRrssAnonGibAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak resident memory use (100=all memory)",
		0,
		"min-rres-peak",
		kRrssAnonGibAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much average GPU use (100=1 full GPU card)",
		0,
		"min-gpu-avg",
		kGpuPctAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak GPU use (100=1 full GPU card)",
		0,
		"min-gpu-peak",
		kGpuPctPeak,
		false,
	},
	uintArg{
		"Select only jobs with at most this much average GPU use (100=1 full GPU card)",
		bigValue,
		"max-gpu-avg",
		kGpuPctAvg,
		false,
	},
	uintArg{
		"Select only jobs with at most this much peak GPU use (100=1 full GPU card)",
		bigValue,
		"max-gpu-peak",
		kGpuPctPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average GPU use (100=all cards)",
		0,
		"min-rgpu-avg",
		kRgpuPctAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak GPU use (100=all cards)",
		0,
		"min-rgpu-peak",
		kRgpuPctPeak,
		true,
	},
	uintArg{
		"Select only jobs with at most this much relative average GPU use (100=all cards)",
		bigValue,
		"max-rgpu-avg",
		kRgpuPctAvg,
		true,
	},
	uintArg{
		"Select only jobs with at most this much relative peak GPU use (100=all cards)",
		bigValue,
		"max-rgpu-peak",
		kRgpuPctPeak,
		true,
	},
	uintArg{
		"Select only jobs with at least this much average GPU memory use (100=1 full GPU card)",
		0,
		"min-gpumem-avg",
		kGpuGibAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak GPU memory use (100=1 full GPU card)",
		0,
		"min-gpumem-peak",
		kGpuGibPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average GPU memory use (100=all cards)",
		0,
		"min-rgpumem-avg",
		kRgpuGibAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak GPU memory use (100=all cards)",
		0,
		"min-rgpumem-peak",
		kRgpuGibPeak,
		true,
	},
}

type JobsCommand struct /* implements AnalysisCommand */ {
	SharedArgs
	ConfigFileArgs

	// Filter args
	Uints         map[string]*uint
	NoGpu         bool
	SomeGpu       bool
	Completed     bool
	Running       bool
	Zombie        bool
	Batch         bool
	MinRuntimeSec int64

	// Print args
	NumJobs uint
	Fmt     string

	// Synthesized and other
	printFields []string
	printOpts   *FormatOptions

	// Internal / working storage
	minRuntimeStr string
}

func (_ *JobsCommand) Summary() []string {
	return []string{
		"Select jobs by various criteria and present aggregate information",
		"about them.",
	}
}

func (jc *JobsCommand) lookupUint(s string) uint {
	if v, ok := jc.Uints[s]; ok {
		return *v
	}
	panic("Unknown parameter key " + s)
}

func (jc *JobsCommand) Add(fs *flag.FlagSet) {
	jc.SharedArgs.Add(fs)
	jc.ConfigFileArgs.Add(fs)

	// Filter args
	jc.Uints = make(map[string]*uint)
	for _, v := range uintArgs {
		box := new(uint)
		fs.UintVar(box, v.name, v.initial, v.text)
		jc.Uints[v.name] = box
	}
	fs.BoolVar(&jc.NoGpu, "no-gpu", false, "Select only jobs with no GPU use")
	fs.BoolVar(&jc.SomeGpu, "some-gpu", false, "Select only jobs with some GPU use")
	fs.BoolVar(&jc.Completed, "completed", false, "Select only jobs that have run to completion")
	fs.BoolVar(&jc.Running, "running", false, "Select only jobs that are still running")
	fs.BoolVar(&jc.Zombie, "zombie", false, "Select only zombie jobs (usually these are still running)")
	fs.BoolVar(&jc.Batch, "batch", false,
		"Aggregate data across all hosts (appropriate for batch systems, but usually specified in the\n"+
			"config file, not here")
	fs.StringVar(&jc.minRuntimeStr, "min-runtime", "",
		"Select only jobs with at least this much runtime, format `WwDdHhMm`, all parts\n"+
			"optional [default: 0m]")

	// Print args
	fs.UintVar(&jc.NumJobs, "numjobs", 0,
		"Print at most these many most recent jobs per user [default: all]")
	fs.UintVar(&jc.NumJobs, "n", 0, "Short for -numjobs n")
	fs.StringVar(&jc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (jc *JobsCommand) ReifyForRemote(x *Reifier) error {
	e1 := jc.SharedArgs.ReifyForRemote(x)
	e2 := jc.ConfigFileArgs.ReifyForRemote(x)

	for _, v := range uintArgs {
		box := jc.Uints[v.name]
		if *box != v.initial {
			x.UintUnchecked(v.name, *box)
		}
	}
	x.Bool("no-gpu", jc.NoGpu)
	x.Bool("some-gpu", jc.SomeGpu)
	x.Bool("completed", jc.Completed)
	x.Bool("running", jc.Running)
	x.Bool("zombie", jc.Zombie)
	x.Bool("batch", jc.Batch)
	x.String("min-runtime", jc.minRuntimeStr)
	x.Uint("numjobs", jc.NumJobs)
	x.String("fmt", jc.Fmt)

	return errors.Join(e1, e2)
}

func (jc *JobsCommand) Validate() error {
	e1 := jc.SharedArgs.Validate()
	e2 := jc.ConfigFileArgs.Validate()

	var e3 error
	if jc.minRuntimeStr != "" {
		var re *regexp.Regexp
		re, e3 = regexp.Compile(`^(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?$`)
		if e3 != nil {
			panic(e3)
		}
		if matches := re.FindStringSubmatch(jc.minRuntimeStr); matches != nil {
			var weeks, days, hours, minutes int64
			var x1, x2, x3, x4 error
			if matches[1] != "" {
				weeks, x1 = strconv.ParseInt(matches[1], 10, 64)
			}
			if matches[2] != "" {
				days, x2 = strconv.ParseInt(matches[2], 10, 64)
			}
			if matches[3] != "" {
				hours, x3 = strconv.ParseInt(matches[3], 10, 64)
			}
			if matches[4] != "" {
				minutes, x4 = strconv.ParseInt(matches[4], 10, 64)
			}
			jc.MinRuntimeSec = (((weeks*7+days)*24+hours)*60 + minutes) * 60
			e3 = errors.Join(x1, x2, x3, x4)
			if e3 != nil {
				e3 = fmt.Errorf("Invalid runtime specifier, try -h")
			}
		} else {
			e3 = errors.New("Invalid runtime specifier, try -h")
		}
	}

	var e4 error
	spec := jobsDefaultFields
	if jc.Fmt != "" {
		spec = jc.Fmt
	}
	var others map[string]bool
	jc.printFields, others, e4 = ParseFormatSpec(spec, jobsFormatters, jobsAliases)
	if e4 == nil && len(jc.printFields) == 0 {
		e4 = errors.New("No output fields were selected in format string")
	}
	jc.printOpts = StandardFormatOptions(others, DefaultFixed)

	return errors.Join(e1, e2, e3, e4)
}

func (jc *JobsCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	allUsers, skipSystemUsers, determined := jc.RecordFilterArgs.DefaultUserFilters()
	if !determined {
		// `--zombie` implies `--user=-` because the use case for `--zombie` is to hunt
		// across all users.
		allUsers, skipSystemUsers = jc.Zombie, false
	}
	excludeSystemCommands = true
	excludeHeartbeat = true
	return
}