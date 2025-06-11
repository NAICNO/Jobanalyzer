package cards

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"

	. "sonalyze/cmd"
	"sonalyze/data/card"
	"sonalyze/db"
	"sonalyze/db/repr"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o card-table.go cards.go

/*TABLE card

package cards

import "sonalyze/db/repr"

%%

FIELDS *repr.SysinfoCardData

 Time           string desc:"Full ISO timestamp of when the reading was taken"
 Node           string desc:"Card's node at this time"
 Index          uint64 desc:"Card's index on its node at this time"
 UUID           string desc:"Card's unique identifier (but not necessarily its only unique identifier)"
 Address        string desc:"Card's address on its node at this time"
 Manufacturer   string desc:"Card's manufacturer's name"
 Model          string desc:"Card model"
 Architecture   string desc:"Card's architecture name"
 Driver         string desc:"Card driver's version at this time"
 Firmware       string desc:"Card firmware's version at this time"
 Memory         uint64 desc:"Card's memory in KB"
 PowerLimit     uint64 desc:"Card's power limit at this time"
 MaxPowerLimit  uint64 desc:"Card's maximum power limit"
 MinPowerLimit  uint64 desc:"Card's minimum power limit"
 MaxCEClock     uint64 desc:"Card's maximum compute element clock speed"
 MaxMemoryClock uint64 desc:"Card's maximum memory clock speed"

SUMMARY CardCommand

  Print GPU card configuration data

HELP CardCommand

  Extract information about individual gpu cards on the cluster from sysinfo and present it in
  primitive form.  Output records are sorted by time and node name, note cards can be moved between
  nodes from time to time.  The default format is 'fixed'.

ALIASES

  Default  Node,Index,Manufacturer,Model,Memory
  All      Time,Node,Index,UUID,Address,Manufacturer,Model,Architecture,Driver,Firmware,Memory,\
           PowerLimit,MaxPowerLimit,MinPowerLimit,MaxCEClock,MaxMemoryClock

DEFAULTS Default

ELBAT*/

type CardCommand struct {
	HostAnalysisArgs
	FormatArgs
}

var _ = (SimpleCommand)((*CardCommand)(nil))

func (nc *CardCommand) Add(fs *CLI) {
	nc.HostAnalysisArgs.Add(fs)
	nc.FormatArgs.Add(fs)
}

func (nc *CardCommand) Validate() error {
	return errors.Join(
		nc.HostAnalysisArgs.Validate(),
		ValidateFormatArgs(
			&nc.FormatArgs, cardDefaultFields, cardFormatters, cardAliases, DefaultFixed),
	)
}

func (nc *CardCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		nc.HostAnalysisArgs.ReifyForRemote(x),
		nc.FormatArgs.ReifyForRemote(x),
	)
}

func (nc *CardCommand) Perform(_ io.Reader, stdout, stderr io.Writer) error {
	theLog, err := db.OpenReadOnlyDB(nc.ConfigFile(), nc.DataDir, db.FileListCardData, nc.LogFiles)
	if err != nil {
		return err
	}

	records, err :=
		card.Query(
			theLog,
			card.QueryFilter{
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

	records, err = ApplyQuery(nc.ParsedQuery, cardFormatters, cardPredicates, records)
	if err != nil {
		return err
	}

	// Sort by time first and node name second
	slices.SortFunc(records, func(a, b *repr.SysinfoCardData) int {
		if rel := cmp.Compare(a.Time, b.Time); rel != 0 {
			return rel
		}
		return cmp.Compare(a.Node, b.Node)
	})

	FormatData(
		stdout,
		nc.PrintFields,
		cardFormatters,
		nc.PrintOpts,
		records,
	)

	return nil
}
