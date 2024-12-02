// DO NOT EDIT.  Generated from print.go by generate-table

package uptime

import (
	. "sonalyze/table"
)

import (
	"fmt"
	"io"
)

var (
	_ fmt.Formatter
	_ = io.SeekStart
)

// MT: Constant after initialization; immutable
var uptimeFormatters = map[string]Formatter[*UptimeLine]{
	"Device": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(string(d.Device), ctx)
		},
		Help: "(string) Device type: 'host' or 'gpu'",
	},
	"Hostname": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(string(d.Hostname), ctx)
		},
		Help: "(string) Host name for the device",
	},
	"State": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(string(d.State), ctx)
		},
		Help: "(string) Device state: 'up' or 'down'",
	},
	"Start": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Start), ctx)
		},
		Help: "(DateTimeValue) Start time of 'up' or 'down' window",
	},
	"End": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.End), ctx)
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
