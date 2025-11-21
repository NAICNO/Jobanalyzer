// The "configs" API is very similar to the "nodes" API.  However:
//
// - The configs API relies only on the config file, it accepts no data store or file list
// - The command line API:
//   - remotely we have -remote / -cluster / -auth-file as normal
//   - locally we have -config-file but not -data-dir
//   - there is therefore special sauce in the daemon code to deal with the translation
//
// TODO: On big systems it would clearly be interesting to filter by various criteria, eg memory
// size, number of cores or cards.

package configs

import (
	"cmp"
	"errors"
	"io"
	"slices"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/data/common"
	"sonalyze/data/config"
	"sonalyze/db/repr"
	"sonalyze/db/types"
	. "sonalyze/table"
)

//go:generate ../../../generate-table/generate-table -o config-table.go configs.go

/*TABLE config

package configs

import "sonalyze/db/repr"

%%

FIELDS *repr.NodeSummary

 Timestamp     string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
 Hostname      string desc:"Name that host is known by on the cluster" alias:"host"
 Description   string desc:"End-user description, not parseable" alias:"desc"
 CpuCores      int    desc:"Total number of cores x threads" alias:"cores"
 MemGB         int    desc:"GB of installed main RAM" alias:"mem"
 GpuCards      int    desc:"Number of installed cards" alias:"gpus"
 GpuMemGB      int    desc:"Total GPU memory across all cards" alias:"gpumem"
 GpuMemPct     bool   desc:"True if GPUs report accurate memory usage in percent" alias:"gpumempct"

SUMMARY ConfigCommand

Display information about nodes in a cluster configuration.

The node configuration is time-dependent and is computed from data reported
by the node and additional background information.

For overall cluster data, use "cluster".  Also see "node" for closely
related data.

HELP ConfigCommand

  Extract information about individual nodes on the cluster from config data and
  present them in primitive form.  Output records are sorted by node name.  The
  default format is 'fixed'.

ALIASES

  default  host,cores,mem,gpus,gpumem,desc
  Default  Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description
  all      timestamp,host,desc,cores,mem,gpus,gpumem,gpumempct
  All      Timestamp,Hostname,Description,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct

DEFAULTS default

ELBAT*/

type ConfigCommand struct {
	DevArgs
	SourceArgs
	QueryArgs
	HostArgs
	VerboseArgs
	FormatArgs
}

var _ = SimpleCommand((*ConfigCommand)(nil))

func (cc *ConfigCommand) Add(fs *CLI) {
	cc.DevArgs.Add(fs)
	cc.SourceArgs.Add(fs, 14)
	cc.QueryArgs.Add(fs)
	cc.HostArgs.Add(fs)
	cc.VerboseArgs.Add(fs)
	cc.FormatArgs.Add(fs)
}

func (cc *ConfigCommand) ReifyForRemote(x *ArgReifier) error {
	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		cc.DevArgs.ReifyForRemote(x),
		cc.SourceArgs.ReifyForRemote(x),
		cc.QueryArgs.ReifyForRemote(x),
		cc.HostArgs.ReifyForRemote(x),
		cc.FormatArgs.ReifyForRemote(x),
	)
}

func (cc *ConfigCommand) Validate() error {
	return errors.Join(
		cc.DevArgs.Validate(),
		cc.SourceArgs.Validate(),
		cc.QueryArgs.Validate(),
		cc.HostArgs.Validate(),
		cc.VerboseArgs.Validate(),
		ValidateFormatArgs(
			&cc.FormatArgs, configDefaultFields, configFormatters, configAliases, DefaultFixed),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Analysis

func (cc *ConfigCommand) Perform(meta types.Context, _ io.Reader, stdout, _ io.Writer) error {
	hosts, err := NewHosts(true, cc.HostArgs.Host)
	if err != nil {
		return err
	}
	includeHosts := hosts.HostnameGlobber()

	cdp, err := config.OpenConfigDataProvider(meta)
	if err != nil {
		return err
	}
	ns, err := cdp.Query(
		config.QueryArgs{
			QueryFilter: common.QueryFilter{
				HaveFrom: cc.HaveFrom,
				FromDate: cc.FromDate,
				HaveTo:   cc.HaveTo,
				ToDate:   cc.ToDate,
			},
			Verbose: cc.Verbose,
			Newest:  true,
		},
	)
	if err != nil {
		return err
	}
	records := make([]*repr.NodeSummary, 0, len(ns))
	for _, n := range ns {
		records = append(records, &n.NodeSummary)
	}
	// TODO: Why would this not be included in the query?  It would make the most sense.  Probably
	// old code.
	if !includeHosts.IsEmpty() {
		records = slices.DeleteFunc(records, func(r *repr.NodeSummary) bool {
			return !includeHosts.Match(r.Hostname)
		})
	}
	records, err = ApplyQuery(cc.ParsedQuery, configFormatters, configPredicates, records)
	if err != nil {
		return err
	}

	slices.SortFunc(records, func(a, b *repr.NodeSummary) int {
		return cmp.Compare(a.Hostname, b.Hostname)
	})

	FormatData(
		stdout,
		cc.PrintFields,
		configFormatters,
		cc.PrintOpts,
		records,
	)

	return nil
}
