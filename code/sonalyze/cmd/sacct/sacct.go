package sacct

import (
	"errors"
	"flag"
	"math"

	. "sonalyze/cmd"
	. "sonalyze/common"
	. "sonalyze/table"
)

const (
	defaultMinRuntime       = 600
	defaultMinReservedMem   = 10
	defaultMinReservedCores = 4
)

type SacctCommand struct /* implements AnalysisCommand */ {
	// Almost SharedArgs, but HostArgs instead of RecordFilterArgs
	DevArgs
	SourceArgs
	HostArgs
	VerboseArgs
	ConfigFileArgs
	FormatArgs

	// Selections
	// COMPLETED, TIMEOUT etc - "or" these, except "" for all
	State []string
	// User names
	User []string
	// Account names
	Account []string
	// Partition names
	Partition []string
	Job       []uint32

	All              bool
	MinRuntime       int64
	MaxRuntime       int64
	MinReservedMem   uint
	MaxReservedMem   uint
	MinReservedCores uint
	MaxReservedCores uint
	SomeGPU          bool
	NoGPU            bool
	Regular          bool
	Array            bool
	Het              bool

	// Internal
	minRuntimeStr string
	maxRuntimeStr string
}

var _ = AnalysisCommand((*SacctCommand)(nil))

func (_ *SacctCommand) Summary() []string {
	return []string{
		"Extract information from sacct data independent of sample data",
	}
}

func (sc *SacctCommand) Add(fs *flag.FlagSet) {
	sc.DevArgs.Add(fs)
	sc.SourceArgs.Add(fs)
	sc.HostArgs.Add(fs)
	sc.VerboseArgs.Add(fs)
	sc.ConfigFileArgs.Add(fs)
	sc.FormatArgs.Add(fs)
	fs.StringVar(&sc.minRuntimeStr, "min-runtime", "",
		"Select jobs with elapsed time at least this, format `WwDdHhMm`, all parts\n"+
			"optional [default: 0m]")
	fs.StringVar(&sc.maxRuntimeStr, "max-runtime", "",
		"Select jobs with elapsed time at most this, format `WwDdHhMm`, all parts\n"+
			"optional [default: 0m]")
	fs.UintVar(&sc.MinReservedMem, "min-reserved-mem", defaultMinReservedMem,
		"Select jobs with reserved memory at least this (GB)")
	fs.UintVar(&sc.MaxReservedMem, "max-reserved-mem", 0,
		"Select jobs with reserved memory at most this (GB)")
	fs.UintVar(&sc.MinReservedCores, "min-reserved-cores", defaultMinReservedCores,
		"Select jobs with reserved cores (cpus * nodes) at least this")
	fs.UintVar(&sc.MaxReservedCores, "max-reserved-cores", 0,
		"Select jobs with reserved cores (cpus * nodes) at most this")
	fs.BoolVar(&sc.All, "all", false, "Set all the -min-whatever filters to their minimum values")
	fs.BoolVar(&sc.SomeGPU, "some-gpu", false, "Select jobs that requested GPUs")
	fs.BoolVar(&sc.NoGPU, "no-gpu", false, "Select jobs that did not request GPUs")
	fs.BoolVar(&sc.Regular, "regular", false, "Show regular jobs (default)")
	fs.BoolVar(&sc.Array, "array", false, "Show array jobs")
	fs.BoolVar(&sc.Het, "het", false, "Show het jobs")
	fs.Var(NewRepeatableString(&sc.State), "state",
		"Select jobs with state `state,...`: COMPLETED, CANCELLED, DEADLINE, FAILED, OUT_OF_MEMORY, TIMEOUT")
	fs.Var(NewRepeatableString(&sc.User), "user",
		"Select jobs with user `user1,...`")
	fs.Var(NewRepeatableString(&sc.Account), "account",
		"Select jobs with account `account1,...`")
	fs.Var(NewRepeatableString(&sc.Partition), "partition",
		"Select jobs on partition `partition1,...`")
	fs.Var(NewRepeatableUint32(&sc.Job), "job",
		"Select jobs with primary job ID `job1,...`")
}

func (sc *SacctCommand) ReifyForRemote(x *ArgReifier) error {
	x.RepeatableString("state", sc.State)
	x.RepeatableString("user", sc.User)
	x.RepeatableString("account", sc.Account)
	x.RepeatableString("partition", sc.Partition)
	x.RepeatableUint32("job", sc.Job)
	x.String("min-runtime", sc.minRuntimeStr)
	x.String("max-runtime", sc.maxRuntimeStr)
	x.Uint("min-reserved-cores", sc.MinReservedCores)
	x.Uint("max-reserved-cores", sc.MaxReservedCores)
	x.Uint("min-reserved-mem", sc.MinReservedMem)
	x.Uint("max-reserved-mem", sc.MaxReservedMem)
	x.Bool("all", sc.All)
	x.Bool("some-gpu", sc.SomeGPU)
	x.Bool("no-gpu", sc.NoGPU)
	x.Bool("regular", sc.Regular)
	x.Bool("array", sc.Array)
	x.Bool("het", sc.Het)
	return errors.Join(
		sc.DevArgs.ReifyForRemote(x),
		sc.SourceArgs.ReifyForRemote(x),
		sc.HostArgs.ReifyForRemote(x),
		sc.ConfigFileArgs.ReifyForRemote(x),
		sc.FormatArgs.ReifyForRemote(x),
	)
}

func (sc *SacctCommand) Validate() error {
	var e1, e7, e8, e9 error
	e1 = errors.Join(
		sc.DevArgs.Validate(),
		sc.SourceArgs.Validate(),
		sc.HostArgs.Validate(),
		sc.VerboseArgs.Validate(),
		sc.ConfigFileArgs.Validate(),
		ValidateFormatArgs(
			&sc.FormatArgs, sacctDefaultFields, sacctFormatters, sacctAliases, DefaultFixed),
	)
	sc.MinRuntime, e7 = DurationToSeconds("-min-runtime", sc.minRuntimeStr)
	sc.MaxRuntime, e8 = DurationToSeconds("-max-runtime", sc.maxRuntimeStr)

	if sc.minRuntimeStr == "" {
		sc.MinRuntime = defaultMinRuntime
	}

	if sc.All {
		sc.MinRuntime = 0
		sc.MinReservedMem = 0
		sc.MinReservedCores = 0
		sc.MaxRuntime = 0
		sc.MaxReservedMem = 0
		sc.MaxReservedCores = 0
	}

	if sc.MaxRuntime == 0 {
		sc.MaxRuntime = math.MaxInt64
	}
	if sc.MaxReservedMem == 0 {
		sc.MaxReservedMem = math.MaxUint
	}
	if sc.MaxReservedCores == 0 {
		sc.MaxReservedCores = math.MaxUint
	}

	var types int
	if sc.Regular {
		types++
	}
	if sc.Array {
		types++
	}
	if sc.Het {
		types++
	}
	switch {
	case types == 0:
		sc.Regular = true
	case types > 1:
		e9 = errors.New("Too many output types, pick only one")
	}

	return errors.Join(e1, e7, e8, e9)
}
