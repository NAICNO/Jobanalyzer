// Compute system load aggregates from a set of log entries.

package load

import (
	"errors"

	. "sonalyze/cmd"
	. "sonalyze/table"
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

type LoadCommand struct /* implements SampleAnalysisCommand */ {
	SampleAnalysisArgs
	FormatArgs

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

	// Synthesized and other
	bucketing bucketTy
}

var _ SampleAnalysisCommand = (*LoadCommand)(nil)

func (lc *LoadCommand) Add(fs *CLI) {
	lc.SampleAnalysisArgs.Add(fs)
	lc.FormatArgs.Add(fs)

	fs.Group("aggregation")
	fs.BoolVar(&lc.Hourly, "hourly", false, "Bucket and average records hourly [default]")
	fs.BoolVar(&lc.HalfHourly, "half-hourly", false, "Bucket and average records half-hourly")
	fs.BoolVar(&lc.Daily, "daily", false, "Bucket and average records daily")
	fs.BoolVar(&lc.HalfDaily, "half-daily", false, "Bucket and average records half-daily")
	fs.BoolVar(&lc.Weekly, "weekly", false, "Bucket and average records weekly")
	fs.BoolVar(&lc.None, "none", false, "Do not bucket and average records")
	fs.BoolVar(&lc.Group, "group", false, "Sum bucketed/averaged data across all the selected hosts")

	fs.Group("printing")
	fs.BoolVar(&lc.All, "all", false,
		"Print records for all times (after bucketing), cf --last [default]")
	fs.BoolVar(&lc.Last, "last", false, "Print records for the last time instant (after bucketing)")
	fs.BoolVar(&lc.Compact, "compact", false,
		"After bucketing, do not print anything for time slots that are empty")
}

func (lc *LoadCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		lc.SampleAnalysisArgs.ReifyForRemote(x),
		lc.FormatArgs.ReifyForRemote(x),
	)

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

	return e1
}

func (lc *LoadCommand) Validate() error {
	e1 := errors.Join(
		lc.SampleAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&lc.FormatArgs, loadDefaultFields, loadFormatters, loadAliases, DefaultFixed),
	)

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
		e4 = errors.New("Incoherent printing options: both -all and -last")
	}
	if !lc.All && !lc.Last {
		lc.All = true
	}

	var e5 error
	if lc.Group && lc.bucketing == bNone {
		e5 = errors.New("Grouping across hosts requires first bucketing by time")
	}

	var e6 error
	if lc.ParsedQuery != nil {
		e6 = errors.New("-q queries not supported for `profile`")
	}

	return errors.Join(e1, e3, e4, e5, e6)
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
