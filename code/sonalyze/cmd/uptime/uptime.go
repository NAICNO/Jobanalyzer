// Compute uptime for a host or a host's GPUs.
//
// Given a list of Samples, including heartbeat records, the uptime for each host can be computed
// by looking at gaps in the timeline of observations for the host.  If a gap exceeds the threshold
// for the gap, we assume the system was down.

package uptime

import (
	_ "embed"
	"errors"
	"fmt"
	"io"

	. "sonalyze/cmd"
	. "sonalyze/table"
)

type UptimeCommand struct /* implements SampleAnalysisCommand */ {
	SharedArgs
	FormatArgs

	Interval uint
	OnlyUp   bool
	OnlyDown bool
}

var _ SampleAnalysisCommand = (*UptimeCommand)(nil)

//go:embed summary.txt
var summary string

func (_ *UptimeCommand) Summary(out io.Writer) {
	fmt.Fprintf(out, summary)
}

func (uc *UptimeCommand) Add(fs *CLI) {
	uc.SharedArgs.Add(fs)
	uc.FormatArgs.Add(fs)

	fs.Group("application-control")
	fs.UintVar(&uc.Interval, "interval", 0,
		"The maximum sampling `interval` in minutes (before any randomization) seen in the data (required)")

	fs.Group("printing")
	fs.BoolVar(&uc.OnlyUp, "only-up", false, "Show only times when systems are up")
	fs.BoolVar(&uc.OnlyDown, "only-down", false, "Show only times when systems are down")
}

func (uc *UptimeCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		uc.SharedArgs.ReifyForRemote(x),
		uc.FormatArgs.ReifyForRemote(x),
	)
	x.Uint("interval", uc.Interval)
	x.Bool("only-up", uc.OnlyUp)
	x.Bool("only-down", uc.OnlyDown)
	return e1
}

func (uc *UptimeCommand) Validate() error {
	var e1, e3, e4, e5 error
	e1 = uc.SharedArgs.Validate()
	if uc.Interval == 0 {
		e3 = errors.New("-interval is required")
	}
	if uc.OnlyUp && uc.OnlyDown {
		e4 = errors.New("Nonsensical -only-up AND -only-down")
	}
	e5 = ValidateFormatArgs(
		&uc.FormatArgs, uptimeDefaultFields, uptimeFormatters, uptimeAliases, DefaultFixed)
	return errors.Join(e1, e3, e4, e5)
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
