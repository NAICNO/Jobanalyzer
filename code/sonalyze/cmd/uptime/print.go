package uptime

import (
	"cmp"
	"io"
	"slices"

	. "sonalyze/table"
)

// (TODO: Some of these comments may be obsolete?)
//
// The time ranges are not clearly inclusive or exclusive, it depends on the data available.  We use
// the inclusive end points of the total date range when we're at the ends of that range, and the
// dates from a boundary Sample otherwise.  Adjacent reports will tend to overlap, by design, as the
// date used for the end of one report is the same as the one for the beginning of the next.
//
// The start and end times are represented as strings to be compatible with the Rust code, which
// rounds down the time to minute precision.  This matters only when there are some closely related
// records in the data stream, which normally does not happen but can happen, and since the printing
// code only prints times to minute precision the data would appear unsorted on output if we were to
// use second precision internally.
//
// TODO: IMPROVEME: It's a bug in both the Rust code and this code that the folding of the timestamp
// to string (and the rounding off to minute precision) happens *after* we determine the timeline.
// In some cases, this will result in duplicate records in the output, where the data stream we
// operated on had sub-minute precision and we constructed records according to that.  Mostly this
// will happen when the data are somewhat wonky, but it's still wrong.

//go:generate ../../../generate-table/generate-table -o uptime-table.go print.go

/*TABLE uptime

package uptime

%%

FIELDS *UptimeLine

 Device   string        alias:"device" desc:"Device type: 'host' or 'gpu'"
 Hostname string        alias:"host"   desc:"Host name for the device"
 State    string        alias:"state"  desc:"Device state: 'up' or 'down'"
 Start    DateTimeValue alias:"start"  desc:"Start time of 'up' or 'down' window"
 End      DateTimeValue alias:"end"    desc:"End time of 'up' or 'down' window"

GENERATE UptimeLine

SUMMARY UptimeCommand

Display information about uptime and downtime of nodes and components.

The output is a timeline with uptime and downtime printed in ascending
order, hosts before devices on the host.  Periods where the node/device is
up or down are both printed, but one can select one or the other with
"-only-up" and "-only-down".

The "-interval" switch must be specified and should be the interval in
minutes for samples on the nodes in question.

A host or device is up at the start of the timeline if its first Sample is
within a small factor of the interval of the "from" time, and ditto it is
up at the end for its last Sample close to the "to" time.

HELP UptimeCommand

  Compute the status of hosts and GPUs across time.  Default output format
  is 'fixed'.

ALIASES

 default device,host,state,start,end
 Default Device,Hostname,State,Start,End
 all     default
 All     Default

DEFAULTS default

ELBAT*/

func (uc *UptimeCommand) printReports(out io.Writer, reports []*UptimeLine) error {
	reports, err := ApplyQuery(uc.ParsedQuery, uptimeFormatters, uptimePredicates, reports)
	if err != nil {
		return err
	}

	slices.SortFunc(reports, func(a, b *UptimeLine) int {
		c := cmp.Compare(a.Hostname, b.Hostname)
		if c == 0 {
			c = cmp.Compare(a.Start, b.Start)
			if c == 0 {
				if a.Device != b.Device {
					if a.Device == "host" {
						c = -1
					} else {
						c = 1
					}
				}
				if c == 0 {
					c = cmp.Compare(a.End, b.End)
					if c == 0 {
						c = cmp.Compare(a.State, b.State)
					}
				}
			}
		}
		return c
	})
	FormatData(
		out,
		uc.PrintFields,
		uptimeFormatters,
		uc.PrintOpts,
		reports,
	)
	return nil
}
