package sparts

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/data/slurmpart"
	"sonalyze/db/special"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o spart-table.go sparts.go

/*TABLE spart

package sparts

%%

FIELDS SpartData

 Timestamp   string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
 Partition   string desc:"Name of the partition" alias:"part"
 Nodes       []string desc:"Node list" alias:"nodes"

GENERATE SpartData

SUMMARY SpartCommand

  Print Slurm partition data

HELP SpartCommand

  Slurm partitions are named and contain a set of nodes on the machine.  Not all nodes need be in a
  partition, some nodes may be in multiple partitions at the same time, and admins can move nodes
  among partitions.  Thus the partition information (also available from the "sinfo" command) is
  time-varying.  Output records are sorted by time and partition name, the default format is
  "fixed".

ALIASES

  default  part,nodes
  Default  Partition,Nodes
  all      timestamp,part,nodes
  All      Timestamp,Partition,Nodes

DEFAULTS default

ELBAT*/

type SpartCommand struct {
	HostAnalysisArgs
	FormatArgs
}

var _ = SimpleCommand((*SpartCommand)(nil))

func (nc *SpartCommand) Add(fs *CLI) {
	nc.HostAnalysisArgs.Add(fs)
	nc.FormatArgs.Add(fs)
}

func (nc *SpartCommand) Validate() error {
	return errors.Join(
		nc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, spartDefaultFields, spartFormatters, spartAliases, DefaultFixed),
	)
}

func (nc *SpartCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		nc.HostAnalysisArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *SpartCommand) Perform(meta special.ClusterMeta, _ io.Reader, stdout, stderr io.Writer) error {
	spd, err := slurmpart.OpenSlurmPartitionDataProvider(meta)
	if err != nil {
		return err
	}

	records, err :=
		spd.Query(
			slurmpart.QueryFilter{
				HaveFrom: nc.HaveFrom,
				FromDate: nc.FromDate,
				HaveTo:   nc.HaveTo,
				ToDate:   nc.ToDate,
			},
			nc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}

	reports := make([]SpartData, 0)
	for _, r := range records {
		for _, p := range r.Partitions {
			nodes := make([]string, len(p.Nodes))
			for i, node := range p.Nodes {
				nodes[i] = string(node)
			}
			reports = append(reports, SpartData{
				Timestamp: r.Time,
				Partition: string(p.Name),
				Nodes:     nodes,
			})
		}
	}

	reports, err = ApplyQuery(nc.ParsedQuery, spartFormatters, spartPredicates, reports)
	if err != nil {
		return err
	}

	slices.SortFunc(reports, func(a, b SpartData) int {
		if rel := cmp.Compare(a.Timestamp, b.Timestamp); rel != 0 {
			return rel
		}
		return cmp.Compare(a.Partition, b.Partition)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		spartFormatters,
		nc.PrintOpts,
		reports,
	)

	return nil
}
