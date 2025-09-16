// Produce a per-node timeline of gpu data.

package gpus

import (
	"errors"
	"fmt"
	"io"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/gpusample"
	"sonalyze/db"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o gpus-table.go gpus.go

/*TABLE gpu

package gpus

%%

FIELDS *ReportLine

 Timestamp   DateTimeValue desc:"Timestamp of when the reading was taken"
 Hostname    Ustr          desc:"Name that host is known by on the cluster"
 Index       uint64        desc:"Card index on the host"
 Fan         uint64        desc:"Fan speed in percent of max"
 Memory      uint64        desc:"Amount of memory in use"
 Temperature int64         desc:"Card temperature in degrees C"
 Power       uint64        desc:"Current power draw in Watts"
 PowerLimit  uint64        desc:"Current power limit in Watts"
 CEClock     uint64        desc:"Current compute element clock in MHz"
 MemoryClock uint64        desc:"Current memory clock in MHz"

SUMMARY GpuCommand

Experimental: Print per-gpu data across time for one or more cards on one or more nodes.

HELP GpuCommand

  Extract information about individual gpus on the cluster from sample data.  The default
  format is 'fixed'.

ALIASES

  default   Hostname,Gpu,Timestamp,Memory,PowerDraw
  Default   Hostname,Gpu,Timestamp,Memory,PowerDraw
  All       Timestamp,Hostname,Index,Fan,Memory,Temperature,PowerDraw,\
            PowerLimit,CEClock,MemoryClock

DEFAULTS default

ELBAT*/

type GpuCommand struct /* implements AnalysisCommand */ {
	HostAnalysisArgs
	FormatArgs

	Gpu int
}

var _ = AnalysisCommand((*GpuCommand)(nil))

func (gc *GpuCommand) Add(fs *CLI) {
	gc.HostAnalysisArgs.Add(fs)
	gc.FormatArgs.Add(fs)
	fs.Group("record-filter")
	fs.IntVar(&gc.Gpu, "gpu", -1, "Select single GPU")
}

func (gc *GpuCommand) Validate() error {
	return errors.Join(
		gc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&gc.FormatArgs, gpuDefaultFields, gpuFormatters, gpuAliases, DefaultFixed),
	)
}

func (gc *GpuCommand) ReifyForRemote(x *ArgReifier) error {
	if gc.Gpu != -1 {
		x.IntUnchecked("gpu", gc.Gpu)
	}
	return errors.Join(
		gc.HostAnalysisArgs.ReifyForRemote(x),
		gc.FormatArgs.ReifyForRemote(x),
	)
}

type ReportLine struct {
	Timestamp DateTimeValue
	Hostname  Ustr
	Gpu       int
	*gpusample.PerGpuSample
}

func (gc *GpuCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	theLog, err := db.OpenReadOnlyDB(gc.ConfigFile(), gc.DataDir, db.FileListGpuSampleData, gc.LogFiles)
	if err != nil {
		return err
	}
	hostGlobber, err := NewHosts(true, gc.Host)
	if err != nil {
		return err
	}

	streams, _, read, dropped, err :=
		gpusample.ReadGpuSamplesByHost(
			theLog,
			gc.FromDate,
			gc.ToDate,
			hostGlobber,
			gc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}
	if gc.Verbose {
		Log.Infof("%d records read + %d dropped\n", read, dropped)
		UstrStats(stderr, false)
	}

	reports := make([]*ReportLine, 0)
	for _, s := range streams {
		for _, d := range s.Data {
			for i, gpu := range d.Decoded {
				if gc.Gpu == -1 || i == gc.Gpu {
					var r ReportLine
					r.Timestamp = DateTimeValue(d.Time)
					r.Hostname = s.Hostname
					r.Gpu = i
					r.PerGpuSample = &gpu
					reports = append(reports, &r)
				}
			}
		}
	}

	reports, err = ApplyQuery(gc.ParsedQuery, gpuFormatters, gpuPredicates, reports)
	if err != nil {
		return err
	}

	FormatData(
		stdout,
		gc.PrintFields,
		gpuFormatters,
		gc.PrintOpts,
		reports,
	)

	return nil
}
