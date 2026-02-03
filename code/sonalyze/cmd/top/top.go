// Produce a per-node timeline of core busyness.
//
// See summary.txt for info.
//
// TODO:
//
// CSV, awk, json not implemented but coming.  In this case it will be a record with
// nodename,time,string where the string is as above.
//
// In addition to the "text" format there will probably be a "log" format (values 0, 1, 2, 3
// corresponding to 0-12.5%, 12.5%-25%, 25%-50%, 50%-100%), a true-percent format, an absolute-value
// format, and a diff-since-previous format.
//
// Long-term I see the possibility of asking slurm about node allocation for a job and being able to
// filter these data by job#.

package top

import (
	"bufio"
	"cmp"
	_ "embed"
	"fmt"
	"io"
	"slices"

	umaps "go-utils/maps"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/cpusample"
	"sonalyze/db/types"
	. "sonalyze/table"
)

type TopCommand struct /* implements AnalysisCommand */ {
	HostAnalysisArgs
}

var _ = AnalysisCommand((*TopCommand)(nil))
var _ = SimpleCommand((*TopCommand)(nil))

//go:embed summary.txt
var summary string

func (tc *TopCommand) Summary(out io.Writer) {
	fmt.Fprint(out, summary)
}

func (tc *TopCommand) Add(fs *CLI) {
	tc.HostAnalysisArgs.Add(fs)
}

func (tc *TopCommand) Validate() error {
	return tc.HostAnalysisArgs.Validate()
}

func (tc *TopCommand) ReifyForRemote(x *ArgReifier) error {
	return tc.HostAnalysisArgs.ReifyForRemote(x)
}

func (tc *TopCommand) MaybeFormatHelp() *FormatHelp {
	// FIXME, but currently no format options at all
	return nil
}

func (tc *TopCommand) Perform(meta types.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	cdp, err := cpusample.OpenCpuSampleDataProvider(meta)

	// TODO: Use standard query interface with standard QueryFilter here.

	hostGlobber, err := NewHosts(true, tc.Host)
	if err != nil {
		return err
	}
	streams, _, read, dropped, err :=
		cdp.Query(
			tc.FromDate,
			tc.ToDate,
			hostGlobber,
			tc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}
	if tc.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	hostStreams := umaps.Values(streams)
	slices.SortFunc(hostStreams, func(a, b *cpusample.CpuSamplesByHost) int {
		return cmp.Compare(a.Hostname.String(), b.Hostname.String())
	})

	// Ad-hoc fixed-format output for now

	buf := bufio.NewWriter(stdout)
	defer buf.Flush()
	for _, v := range hostStreams {
		if len(v.Data) > 0 {
			buf.WriteString("HOST: ")
			buf.WriteString(v.Hostname.String())
			buf.WriteByte('\n')

			for i := 1; i < len(v.Data); i++ {
				tdiff := float64(v.Data[i].Time - v.Data[i-1].Time)
				buf.WriteString("  ")
				buf.WriteString(FormatYyyyMmDdHhMmUtc(v.Data[i].Time))
				buf.WriteByte(' ')
				for j := range v.Data[i].Decoded {
					n := float64(v.Data[i].Decoded[j]-v.Data[i-1].Decoded[j]) / tdiff
					if n >= 0.25 {
						buf.WriteByte('O')
					} else if n >= 0.10 {
						buf.WriteByte('o')
					} else {
						buf.WriteByte('.')
					}
				}
				buf.WriteByte('\n')
			}
		}
	}

	return nil
}
