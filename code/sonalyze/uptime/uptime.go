// Compute uptime for a host or a host's GPUs.
//
// Given a list of Samples, including heartbeat records, the uptime for each host can be computed
// by looking at gaps in the timeline of observations for the host.  If a gap exceeds the threshold
// for the gap, we assume the system was down.

package uptime

import (
	"errors"
	"flag"

	. "sonalyze/command"
)

type UptimeCommand struct /* implements AnalysisCommand */ {
	SharedArgs
	ConfigFileArgs

	Interval uint
	OnlyUp   bool
	OnlyDown bool
	Fmt      string

	// Synthesized and other
	printFields []string
	printOpts   *FormatOptions
}

func (_ *UptimeCommand) Summary() []string {
	return []string{
		"Compute and print information about uptime and downtime of nodes",
		"and components.",
	}
}

func (uc *UptimeCommand) Add(fs *flag.FlagSet) {
	uc.SharedArgs.Add(fs)
	uc.ConfigFileArgs.Add(fs)

	fs.UintVar(&uc.Interval, "interval", 0,
		"The maximum sampling `interval` in minutes (before any randomization) seen in the data")
	fs.BoolVar(&uc.OnlyUp, "only-up", false, "Show only times when systems are up")
	fs.BoolVar(&uc.OnlyDown, "only-down", false, "Show only times when systems are down")
	fs.StringVar(&uc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (uc *UptimeCommand) ReifyForRemote(x *Reifier) error {
	e1 := uc.SharedArgs.ReifyForRemote(x)
	e2 := uc.ConfigFileArgs.ReifyForRemote(x)
	x.Uint("interval", uc.Interval)
	x.Bool("only-up", uc.OnlyUp)
	x.Bool("only-down", uc.OnlyDown)
	x.String("fmt", uc.Fmt)
	return errors.Join(e1, e2)
}

func (uc *UptimeCommand) Validate() error {
	var e1, e2, e3, e4, e5 error
	e1 = uc.SharedArgs.Validate()
	e2 = uc.ConfigFileArgs.Validate()
	if uc.Interval == 0 {
		e3 = errors.New("-interval is required")
	}
	if uc.OnlyUp && uc.OnlyDown {
		e4 = errors.New("Nonsensical -only-up AND -only-down")
	}
	var others map[string]bool
	uc.printFields, others, e5 = ParseFormatSpec(uptimeDefaultFields, uc.Fmt, uptimeFormatters, uptimeAliases)
	if e5 == nil && len(uc.printFields) == 0 {
		e5 = errors.New("No output fields were selected in format string")
	}
	uc.printOpts = StandardFormatOptions(others, DefaultFixed)
	return errors.Join(e1, e2, e3, e4, e5)
}

func (uc *UptimeCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	allUsers = true
	skipSystemUsers = false
	excludeSystemCommands = false
	excludeHeartbeat = false
	return
}
