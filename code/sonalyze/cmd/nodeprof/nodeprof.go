package nodeprof

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/data/nodesample"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o nodeprof-table.go nodeprof.go

/*TABLE nodeprof

package nodeprof

import "sonalyze/db/repr"

%%

FIELDS *repr.NodeSample

  Timestamp        DateTimeValue desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
  Hostname         Ustr          desc:"Name that host is known by on the cluster" alias:"host"
  UsedMemory       uint64        desc:"Amount of memory in use" alias:"usedmem"
  Load1            float64       desc:"1-minute load average" alias:"load1"
  Load5            float64       desc:"5-minute load average" alias:"load5"
  Load15           float64       desc:"15-minute load average" alias:"load15"
  RunnableEntities uint64        desc:"Number of runnable entities on system (threads)" alias:"runnable"
  ExistingEntities uint64        desc:"Number of entities on system" alias:"entitites"

SUMMARY NodeProfCommand

  Display node sample data - memory usage, load averages, and run queue length.

HELP NodeProfCommand

  Extract node profiling data from sample data and present it in primitive form.  Output
  records are sorted by node name and time.  The default format is 'fixed'.

  RunnableEntities is the length of the run queue on the node.  If the number of runnable
  entities is much higher than the available number of cores for the job then the job may
  be starved for CPU.  Note however that the run queue is per-node while the amount of CPU
  available may be per-job; the data will require careful interpretation.

ALIASES

  Default Timestamp,Hostname,UsedMemory,Load1,Load5,RunnableEntities
  All     Timestamp,Hostname,UsedMemory,Load1,Load5,Load15,RunnableEntities,ExistingEntities

DEFAULTS Default

ELBAT*/

type NodeProfCommand struct {
	HostAnalysisArgs
	FormatArgs
}

var _ = SimpleCommand((*NodeProfCommand)(nil))

func (nc *NodeProfCommand) Add(fs *CLI) {
	nc.HostAnalysisArgs.Add(fs)
	nc.FormatArgs.Add(fs)
}

func (nc *NodeProfCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		nc.HostAnalysisArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *NodeProfCommand) Validate() error {
	return errors.Join(
		nc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, nodeprofDefaultFields, nodeprofFormatters, nodeprofAliases, DefaultFixed),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Processing

func (nc *NodeProfCommand) Perform(meta special.ClusterMeta, _ io.Reader, stdout, stderr io.Writer) error {
	theLog, err := db.OpenReadOnlyDB(
		meta,
		db.FileListNodeSampleData,
	)
	if err != nil {
		return err
	}

	records, err := nodesample.Query(
		theLog,
		nodesample.QueryFilter{
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

	records, err = ApplyQuery(nc.ParsedQuery, nodeprofFormatters, nodeprofPredicates, records)
	if err != nil {
		return err
	}

	// Sort by host name first and then by ascending time
	slices.SortFunc(records, func(a, b *repr.NodeSample) int {
		if h := cmp.Compare(a.Hostname, b.Hostname); h != 0 {
			return h
		}
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		nodeprofFormatters,
		nc.PrintOpts,
		records,
	)

	return nil
}
