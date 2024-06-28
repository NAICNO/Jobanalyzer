// Produce a per-node timeline of core busyness.
//
// The fixed output format is like this:
//
// NodeName
//   Time ...oO..O...o...o...
//   Time ...oO..O...O...o...
//   Time ...oO..O.......O...
//   ...
//
// where the characters represent CPUs in order and there is a mark indicating how busy the cpu was
// during the previous time interval.  The mark is "." for "not busy" (< 10% utilization), "o" for
// somewhat busy (< 25%), and "O" for busy.
//
// The output is sorted by node name, then by time.
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
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"

	"go-utils/hostglob"
	umaps "go-utils/maps"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

type TopCommand struct /* implements AnalysisCommand */ {
	// Almost SharedArgs, but HostArgs instead of RecordFilterArgs
	DevArgs
	SourceArgs
	HostArgs
	VerboseArgs
	ConfigFileArgs
}

var _ = AnalysisCommand((*TopCommand)(nil))

func (tc *TopCommand) Summary() []string {
	return []string{
		"Print per-cpu load across time for one or more nodes",
	}
}

func (tc *TopCommand) Add(fs *flag.FlagSet) {
	tc.DevArgs.Add(fs)
	tc.SourceArgs.Add(fs)
	tc.HostArgs.Add(fs)
	tc.VerboseArgs.Add(fs)
	tc.ConfigFileArgs.Add(fs)
}

func (tc *TopCommand) Validate() error {
	var e1, e2, e3, e4, e5 error
	e1 = tc.DevArgs.Validate()
	e2 = tc.SourceArgs.Validate()
	e3 = tc.HostArgs.Validate()
	e4 = tc.VerboseArgs.Validate()
	e5 = tc.ConfigFileArgs.Validate()
	return errors.Join(e1, e2, e3, e4, e5)
}

func (tc *TopCommand) ReifyForRemote(x *Reifier) error {
	// tc.Verbose is not reified, as for SharedArgs.
	return errors.Join(
		tc.DevArgs.ReifyForRemote(x),
		tc.SourceArgs.ReifyForRemote(x),
		tc.HostArgs.ReifyForRemote(x),
		tc.ConfigFileArgs.ReifyForRemote(x),
	)
}

func (tc *TopCommand) MaybeFormatHelp() *FormatHelp {
	// FIXME, but currently no format options at all
	return nil
}

func (tc *TopCommand) Top(stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, err := MaybeGetConfig(tc.ConfigFile())
	if err != nil {
		return err
	}

	hostGlobber, err := hostglob.NewGlobber(true, tc.Host)
	if err != nil {
		return err
	}

	var theLog db.SampleCluster
	if len(tc.LogFiles) > 0 {
		theLog, err = db.OpenTransientSampleCluster(tc.LogFiles, cfg)
	} else {
		theLog, err = db.OpenPersistentCluster(tc.DataDir, cfg)
	}
	if err != nil {
		return fmt.Errorf("Failed to open log store\n%w", err)
	}

	streams, _, read, dropped, err :=
		sonarlog.ReadLoadDataStreams(
			theLog,
			tc.FromDate,
			tc.ToDate,
			hostGlobber,
			tc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records\n%w", err)
	}
	if tc.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	hostStreams := umaps.Values(streams)
	slices.SortFunc(hostStreams, func(a, b *sonarlog.LoadData) int {
		if a.Host.String() < b.Host.String() {
			return -1
		}
		if a.Host.String() > b.Host.String() {
			return 1
		}
		return 0
	})

	// Ad-hoc fixed-format output for now

	buf := bufio.NewWriter(stdout)
	defer buf.Flush()
	for _, v := range hostStreams {
		if len(v.Data) > 0 {
			buf.WriteString("HOST: ")
			buf.WriteString(v.Host.String())
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
