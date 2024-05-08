// Compute system load aggregates from a set of log entries.

package load

import (
	"errors"
	"flag"

	. "sonalyze/command"
)

type bucketTy int

const (
	bNone bucketTy = iota
	bHalfHourly
	bHourly
	bHalfDaily
	bDaily
	bWeekly
)

const (
	loadDefaultFields = "date,time,cpu,mem,gpu,gpumem,gpumask"
)

type LoadCommand struct /* implements Command */ {
	SharedArgs
	ConfigFileArgs

	// Filtering and aggregation args
	Hourly     bool
	HalfHourly bool
	Daily      bool
	HalfDaily  bool
	Weekly     bool
	None       bool
	Group      bool

	// Print args
	All     bool
	Last    bool
	Compact bool
	Fmt     string

	// Synthesized and other
	bucketing   bucketTy
	printFields []string
	printOpts   *FormatOptions
}

func (lc *LoadCommand) Add(fs *flag.FlagSet) {
	lc.SharedArgs.Add(fs)
	lc.ConfigFileArgs.Add(fs)

	fs.BoolVar(&lc.Hourly, "hourly", false, "Bucket and average records hourly [default]")
	fs.BoolVar(&lc.HalfHourly, "half-hourly", false, "Bucket and average records half-hourly")
	fs.BoolVar(&lc.Daily, "daily", false, "Bucket and average records daily")
	fs.BoolVar(&lc.HalfDaily, "half-daily", false, "Bucket and average records half-daily")
	fs.BoolVar(&lc.Weekly, "weekly", false, "Bucket and average records weekly")
	fs.BoolVar(&lc.None, "none", false, "Do not bucket and average records")
	fs.BoolVar(&lc.Group, "group", false, "Sum bucketed/averaged data across all the selected hosts")

	fs.BoolVar(&lc.All, "all", false,
		"Print records for all times (after bucketing), cf --last [default]")
	fs.BoolVar(&lc.Last, "last", false, "Print records for the last time instant (after bucketing)")
	fs.BoolVar(&lc.Compact, "compact", false,
		"After bucketing, do not print anything for time slots that are empty")
	fs.StringVar(&lc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (lc *LoadCommand) ReifyForRemote(x *Reifier) error {
	e1 := lc.SharedArgs.ReifyForRemote(x)
	e2 := lc.ConfigFileArgs.ReifyForRemote(x)

	x.Bool("hourly", lc.Hourly)
	x.Bool("half-hourly", lc.HalfHourly)
	x.Bool("daily", lc.Daily)
	x.Bool("half-daily", lc.HalfDaily)
	x.Bool("weekly", lc.Weekly)
	x.Bool("none", lc.None)
	x.Bool("group", lc.Group)

	x.Bool("all", lc.All)
	x.Bool("last", lc.Last)
	x.Bool("compact", lc.Compact)
	x.String("fmt", lc.Fmt)

	return errors.Join(e1, e2)
}

func (lc *LoadCommand) Validate() error {
	e1 := lc.SharedArgs.Validate()
	e2 := lc.ConfigFileArgs.Validate()

	var e3 error
	n := 0
	if lc.Hourly {
		n++
	}
	if lc.HalfHourly {
		n++
	}
	if lc.Daily {
		n++
	}
	if lc.HalfDaily {
		n++
	}
	if lc.Weekly {
		n++
	}
	if lc.None {
		n++
	}
	if n > 1 {
		e3 = errors.New("Too many bucketing options")
	}
	switch {
	case lc.None:
		lc.bucketing = bNone
	case lc.HalfHourly:
		lc.bucketing = bHalfHourly
	case lc.Hourly:
		lc.bucketing = bHourly
	case lc.HalfDaily:
		lc.bucketing = bHalfDaily
	case lc.Daily:
		lc.bucketing = bDaily
	case lc.Weekly:
		lc.bucketing = bWeekly
	default:
		lc.bucketing = bHourly
	}

	var e4 error
	if lc.All && lc.Last {
		e4 = errors.New("Incoherent printing options")
	}
	if !lc.All && !lc.Last {
		lc.All = true
	}

	var e5 error
	if lc.Group && lc.bucketing == bNone {
		e5 = errors.New("Grouping across hosts requires first bucketing by time")
	}

	var e6 error
	spec := loadDefaultFields
	if lc.Fmt != "" {
		spec = lc.Fmt
	}
	var others map[string]bool
	lc.printFields, others, e6 = ParseFormatSpec(spec, loadFormatters, loadAliases)
	if e6 == nil && len(lc.printFields) == 0 {
		e6 = errors.New("No output fields were selected in format string")
	}
	lc.printOpts = StandardFormatOptions(others)

	return errors.Join(e1, e2, e3, e4, e5, e6)
}

func (lc *LoadCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	// `load` implies `--user=-` b/c we're interested in system effects.
	allUsers = true
	skipSystemUsers = false
	excludeSystemCommands = true
	excludeHeartbeat = false
	return
}
