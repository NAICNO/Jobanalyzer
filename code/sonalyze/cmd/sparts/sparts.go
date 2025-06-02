package sparts

import (
	"errors"
	"fmt"
	"io"

	. "sonalyze/cmd"
	"sonalyze/data/slurmpart"
	"sonalyze/db"
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

  TODO

ALIASES

  default  host,partition,nodes
  Default  Hostname,Partition,Nodes
  all      timestamp,host,part,nodes
  All      Timestamp,Hostname,Partition,Nodes

DEFAULTS default

ELBAT*/

type SpartCommand struct {
	HostAnalysisArgs
	FormatArgs
}

var _ = (SimpleCommand)((*SpartCommand)(nil))

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

func (nc *SpartCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	theLog, err := db.OpenReadOnlyDB(nc.ConfigFile(), nc.DataDir, db.FileListSlurmPartitionData, nc.LogFiles)
	if err != nil {
		return err
	}

	records, err :=
		slurmpart.Query(
			theLog,
			slurmpart.QueryFilter{
				HaveFromDate: nc.HaveFrom,
				FromDate:     nc.FromDate,
				HaveToDate:   nc.HaveTo,
				ToDate:       nc.ToDate,
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

	FormatData(
		stdout,
		nc.PrintFields,
		spartFormatters,
		nc.PrintOpts,
		reports,
	)

	return nil
}
