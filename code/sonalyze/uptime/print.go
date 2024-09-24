package uptime

import (
	"io"
	"sort"

	. "sonalyze/command"
)

func (uc *UptimeCommand) printReports(out io.Writer, reports []report) {
	sort.Sort(sortableReports(reports))
	FormatData(
		out,
		uc.PrintFields,
		uptimeFormatters,
		uc.PrintOpts,
		reports,
		uptimeCtx(false),
	)
}

type sortableReports []report

func (sr sortableReports) Len() int {
	return len(sr)
}

func (sr sortableReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func (sr sortableReports) Less(i, j int) bool {
	// Sort the reports by hostname, start time, device type, end time, and state
	if sr[i].host == sr[j].host {
		if sr[i].start == sr[j].start {
			if sr[i].device == sr[j].device {
				if sr[i].end == sr[j].end {
					return sr[i].state < sr[j].state
				}
				return sr[i].end < sr[j].end
			}
			return sr[i].device == "host"
		}
		return sr[i].start < sr[j].start
	}
	return sr[i].host < sr[j].host
}

func (uc *UptimeCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(uc.Fmt, printHelp, uptimeFormatters, uptimeAliases, uptimeDefaultFields)
}

const printHelp = `
print
  Compute the status of hosts and GPUs across time.  Default output format
  is 'fixed'.
`

const uptimeDefaultFields = "device,host,state,start,end"

// MT: Constant after initialization; immutable
var uptimeAliases = map[string][]string{
	"all": []string{
		"device",
		"host",
		"state",
		"start",
		"end",
	},
}

type uptimeCtx bool

// MT: Constant after initialization; immutable
var uptimeFormatters = map[string]Formatter[report, uptimeCtx]{
	"device": {
		func(d report, _ uptimeCtx) string {
			return d.device
		},
		"Device type: 'host' or 'gpu'",
	},
	"host": {
		func(d report, _ uptimeCtx) string {
			return d.host
		},
		"Host name for the device",
	},
	"state": {
		func(d report, _ uptimeCtx) string {
			return d.state
		},
		"Device state: 'up' or 'down'",
	},
	"start": {
		func(d report, _ uptimeCtx) string {
			return d.start
		},
		"Start time of 'up' or 'down' window (yyyy-mm-dd hh:mm)",
	},
	"end": {
		func(d report, _ uptimeCtx) string {
			return d.end
		},
		"End time of 'up' or 'down' window (yyyy-mm-dd hh:mm)",
	},
}
