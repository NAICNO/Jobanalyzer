package snodes

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/data/slurmnode"
	"sonalyze/db"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o snode-table.go snodes.go

/*TABLE snode

package snodes

%%

FIELDS SnodeData

 # Note the CrossNodeJobs field is a config-level attribute, it does not appear in the raw sysinfo
 # data, and so it is not included here.

 Timestamp   string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
 Nodes       []string desc:"Node list" alias:"nodes"
 States      []string desc:"State list" alias:"states"

GENERATE SnodeData

SUMMARY SnodeCommand

  Print Slurm node data

HELP SnodeCommand

  Nodes managed by Slurm can be in various states and belong to various partitions.  A node may also
  be managed by Slurm at some points in time, and be unmanaged at other points, and can be moved
  among partitions.  Output records are sorted by time.  The default format is 'fixed'.

ALIASES

  default  nodes,states
  Default  Nodes,States
  all      timestamp,nodes,states
  All      Timestamp,Nodes,States

DEFAULTS default

ELBAT*/

type SnodeCommand struct {
	HostAnalysisArgs
	FormatArgs
}

var _ = (SimpleCommand)((*SnodeCommand)(nil))

func (nc *SnodeCommand) Add(fs *CLI) {
	nc.HostAnalysisArgs.Add(fs)
	nc.FormatArgs.Add(fs)
}

func (nc *SnodeCommand) Validate() error {
	return errors.Join(
		nc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, snodeDefaultFields, snodeFormatters, snodeAliases, DefaultFixed),
	)
}

func (nc *SnodeCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		nc.HostAnalysisArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *SnodeCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	theLog, err := db.OpenReadOnlyDB(nc.ConfigFile(), nc.DataDir, db.FileListSlurmNodeData, nc.LogFiles)
	if err != nil {
		return err
	}

	records, err :=
		slurmnode.Query(
			theLog,
			slurmnode.QueryFilter{
				FromDate: nc.FromDate,
				ToDate:   nc.ToDate,
			},
			nc.Verbose,
		)
	if err != nil {
		return fmt.Errorf("Failed to read log records: %v", err)
	}

	reports := make([]SnodeData, 0)
	for _, r := range records {
		for _, n := range r.Nodes {
			names := make([]string, len(n.Names))
			for i, name := range n.Names {
				names[i] = string(name)
			}
			reports = append(reports, SnodeData{
				Timestamp: r.Time,
				Nodes:     names,
				States:    slices.Clone(n.States),
			})
		}
	}

	reports, err = ApplyQuery(nc.ParsedQuery, snodeFormatters, snodePredicates, reports)
	if err != nil {
		return err
	}

	slices.SortFunc(reports, func(a, b SnodeData) int {
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		snodeFormatters,
		nc.PrintOpts,
		reports,
	)

	return nil
}
