// DO NOT EDIT.  Generated from print.go by generate-table

package uptime

import (
	"cmp"
	"fmt"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
)

// MT: Constant after initialization; immutable
var uptimeFormatters = map[string]Formatter[*UptimeLine]{
	"Device": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(d.Device, ctx)
		},
		Help: "(string) Device type: 'host' or 'gpu'",
	},
	"Hostname": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(d.Hostname, ctx)
		},
		Help: "(string) Host name for the device",
	},
	"State": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(d.State, ctx)
		},
		Help: "(string) Device state: 'up' or 'down'",
	},
	"Start": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(d.Start, ctx)
		},
		Help: "(DateTimeValue) Start time of 'up' or 'down' window",
	},
	"End": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(d.End, ctx)
		},
		Help: "(DateTimeValue) End time of 'up' or 'down' window",
	},
}

func init() {
	DefAlias(uptimeFormatters, "Device", "device")
	DefAlias(uptimeFormatters, "Hostname", "host")
	DefAlias(uptimeFormatters, "State", "state")
	DefAlias(uptimeFormatters, "Start", "start")
	DefAlias(uptimeFormatters, "End", "end")
}

type UptimeLine struct {
	Device   string
	Hostname string
	State    string
	Start    DateTimeValue
	End      DateTimeValue
}

func (c *UptimeCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `Display information about uptime and downtime of nodes and components.

The output is a timeline with uptime and downtime printed in ascending
order, hosts before devices on the host.  Periods where the node/device is
up or down are both printed, but one can select one or the other with
"-only-up" and "-only-down".

The "-interval" switch must be specified and should be the interval in
minutes for samples on the nodes in question.

A host or device is up at the start of the timeline if its first Sample is
within a small factor of the interval of the "from" time, and ditto it is
up at the end for its last Sample close to the "to" time.
`)
}

const uptimeHelp = `
uptime
  Compute the status of hosts and GPUs across time.  Default output format
  is 'fixed'.
`

func (c *UptimeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, uptimeHelp, uptimeFormatters, uptimeAliases, uptimeDefaultFields)
}

// MT: Constant after initialization; immutable
var uptimeAliases = map[string][]string{
	"default": []string{"device", "host", "state", "start", "end"},
	"Default": []string{"Device", "Hostname", "State", "Start", "End"},
	"all":     []string{"default"},
	"All":     []string{"Default"},
}

const uptimeDefaultFields = "default"
