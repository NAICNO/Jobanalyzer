// DO NOT EDIT.  Generated from print.go by generate-table

package uptime

import (
	. "sonalyze/table"
)

// MT: Constant after initialization; immutable
var uptimeFormatters = map[string]Formatter[*UptimeLine]{
	"Device": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(string(d.Device), ctx)
		},
		Help: "Device type: 'host' or 'gpu'",
	},
	"Hostname": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(string(d.Hostname), ctx)
		},
		Help: "Host name for the device",
	},
	"State": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatString(string(d.State), ctx)
		},
		Help: "Device state: 'up' or 'down'",
	},
	"Start": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.Start), ctx)
		},
		Help: "Start time of 'up' or 'down' window",
	},
	"End": {
		Fmt: func(d *UptimeLine, ctx PrintMods) string {
			return FormatDateTimeValue(DateTimeValue(d.End), ctx)
		},
		Help: "End time of 'up' or 'down' window",
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
