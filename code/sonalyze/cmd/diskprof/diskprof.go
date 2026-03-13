package diskprof

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/data/disksample"
	"sonalyze/db/repr"
	"sonalyze/db/types"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o diskprof-table.go diskprof.go

/*TABLE diskprof

package diskprof

import "sonalyze/db/repr"

%%

FIELDS *repr.DiskSample

  Timestamp        DateTimeValue desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
  Hostname         Ustr          desc:"Name that host is known by on the cluster" alias:"host"
  Name             Ustr          desc:"Name of disk" alias:"name"
  Major            uint64        desc:"Major device number" alias:"major"
  Minor            uint64        desc:"Minor device number" alias:"minor"
  MsReading        uint64        desc:"ms spent reading" alias:"ms-reading"
  MsWriting        uint64        desc:"ms spent writing" alias:"ms-writing"

SUMMARY DiskProfCommand

  Display disk sample data

HELP DiskProfCommand

  Extract disk profiling data from sample data and present it in primitive form.  Output
  records are sorted by node name and time.  The default format is 'fixed'.

ALIASES

  Default Timestamp,Hostname,Name,MsReading,MsWriting
  All     Timestamp,Hostname,Name,Major,Minor,MsReading,MsWriting

DEFAULTS Default

ELBAT*/

type DiskProfCommand struct {
	HostAnalysisArgs
	FormatArgs
}

var _ = SimpleCommand((*DiskProfCommand)(nil))

func (nc *DiskProfCommand) Add(fs *CLI) {
	nc.HostAnalysisArgs.Add(fs)
	nc.FormatArgs.Add(fs)
}

func (nc *DiskProfCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		nc.HostAnalysisArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *DiskProfCommand) Validate() error {
	return errors.Join(
		nc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, diskprofDefaultFields, diskprofFormatters, diskprofAliases, DefaultFixed),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Processing

func (nc *DiskProfCommand) Perform(meta types.Context, _ io.Reader, stdout, stderr io.Writer) error {
	dsd, err := disksample.OpenDiskSampleDataProvider(meta)
	if err != nil {
		return err
	}

	records, err := dsd.Query(
		disksample.QueryFilter{
			HaveFrom: nc.HaveFrom,
			FromDate: nc.FromDate,
			HaveTo:   nc.HaveTo,
			ToDate:   nc.ToDate,
			Host:     nc.HostArgs.Host,
		},
		nc.Verbose,
	)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}

	records, err = ApplyQuery(nc.ParsedQuery, diskprofFormatters, diskprofPredicates, records)
	if err != nil {
		return err
	}

	// Sort by host name first and then by ascending time
	slices.SortFunc(records, func(a, b *repr.DiskSample) int {
		if h := cmp.Compare(a.Hostname, b.Hostname); h != 0 {
			return h
		}
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		diskprofFormatters,
		nc.PrintOpts,
		records,
	)

	return nil
}
