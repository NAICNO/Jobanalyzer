// Compute jobs aggregates from a set of log entries.

package jobs

import (
	"errors"

	. "sonalyze/cmd"
	. "sonalyze/common"
	. "sonalyze/table"
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

// MT: Constant after initialization; immutable
var uintArgs = []uintArg{
	uintArg{
		"Select only jobs with at least this many samples [default: 1]",
		1,
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
		kCpuGBAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak virtual memory use (GB)",
		0,
		"min-mem-peak",
		kCpuGBPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average virtual memory use (100=all memory)",
		0,
		"min-rmem-avg",
		kRcpuGBAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak virtual memory use (100=all memory)",
		0,
		"min-rmem-peak",
		kRcpuGBAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much average resident memory use (GB)",
		0,
		"min-res-avg",
		kRssAnonGBAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak resident memory use (GB)",
		0,
		"min-res-peak",
		kRssAnonGBPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average resident memory use (100=all memory)",
		0,
		"min-rres-avg",
		kRrssAnonGBAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak resident memory use (100=all memory)",
		0,
		"min-rres-peak",
		kRrssAnonGBAvg,
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
		kGpuGBAvg,
		false,
	},
	uintArg{
		"Select only jobs with at least this much peak GPU memory use (100=1 full GPU card)",
		0,
		"min-gpumem-peak",
		kGpuGBPeak,
		false,
	},
	uintArg{
		"Select only jobs with at least this much relative average GPU memory use (100=all cards)",
		0,
		"min-rgpumem-avg",
		kRgpuGBAvg,
		true,
	},
	uintArg{
		"Select only jobs with at least this much relative peak GPU memory use (100=all cards)",
		0,
		"min-rgpumem-peak",
		kRgpuGBPeak,
		true,
	},
}

type JobsCommand struct /* implements SampleAnalysisCommand */ {
	SharedArgs
	FormatArgs

	// Filter args
	Uints         map[string]*uint
	NoGpu         bool
	SomeGpu       bool
	Completed     bool
	Running       bool
	Zombie        bool
	MergeAll      bool
	MergeNone     bool
	MinRuntimeSec int64

	// Print args
	NumJobs uint

	// Internal / working storage
	minRuntimeStr string
}

var _ SampleAnalysisCommand = (*JobsCommand)(nil)

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

func (jc *JobsCommand) Add(fs *CLI) {
	jc.SharedArgs.Add(fs)
	jc.FormatArgs.Add(fs)

	fs.Group("job-filter")
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
	fs.StringVar(&jc.minRuntimeStr, "min-runtime", "",
		"Select only jobs with at least this much runtime, format `WwDdHhMm`, all parts\n"+
			"optional [default: 0m]")

	fs.Group("aggregation")
	fs.BoolVar(&jc.MergeAll, "merge-all", false,
		"Aggregate data across all hosts (appropriate for batch systems, but usually specified in the\n"+
			"config file, not here")
	fs.BoolVar(&jc.MergeNone, "merge-none", false,
		"Never aggregate data across hosts (appropriate for non-batch systems, but usually specified in the\n"+
			"config file, not here")
	fs.BoolVar(&jc.MergeAll, "batch", false, "Old name for -merge-all")

	fs.Group("printing")
	fs.UintVar(&jc.NumJobs, "numjobs", 0,
		"Print at most these many most recent jobs per user [default: all]")
	fs.UintVar(&jc.NumJobs, "n", 0, "Short for -numjobs n")
}

func (jc *JobsCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		jc.SharedArgs.ReifyForRemote(x),
		jc.FormatArgs.ReifyForRemote(x),
	)
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
	x.Bool("merge-none", jc.MergeNone)
	x.Bool("merge-all", jc.MergeAll)
	x.String("min-runtime", jc.minRuntimeStr)
	x.Uint("numjobs", jc.NumJobs)

	return e1
}

func (jc *JobsCommand) Validate() error {
	var e1, e2, e3 error
	e1 = errors.Join(
		jc.SharedArgs.Validate(),
		ValidateFormatArgs(
			&jc.FormatArgs, jobsDefaultFields, jobsFormatters, jobsAliases, DefaultFixed),
	)

	if jc.MergeAll && jc.MergeNone {
		e2 = errors.New("Can't both -merge-all and -merge-none")
	}

	if jc.minRuntimeStr != "" {
		jc.MinRuntimeSec, e3 = DurationToSeconds("-min-runtime", jc.minRuntimeStr)
	}

	return errors.Join(e1, e2, e3)
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
