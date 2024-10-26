// The "configs" table exposes the config file for a cluster, but only the per-node information.
// The per-cluster information in the config file is exposed by the "clusters" verb.
//
// The "configs" API is therefore very similar to the "nodes" API.  However:
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
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"

	"go-utils/config"
	"go-utils/hostglob"

	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
)

type ConfigCommand struct {
	DevArgs
	HostArgs
	RemotingArgs
	VerboseArgs
	ConfigFileArgs
	FormatArgs
}

func (cc *ConfigCommand) Summary() []string {
	return []string{
		"Extract information about nodes in the cluster configuration",
	}
}

func (cc *ConfigCommand) Add(fs *flag.FlagSet) {
	cc.DevArgs.Add(fs)
	cc.RemotingArgs.Add(fs)
	cc.HostArgs.Add(fs)
	cc.VerboseArgs.Add(fs)
	cc.ConfigFileArgs.Add(fs)
	cc.FormatArgs.Add(fs)
}

func (cc *ConfigCommand) ReifyForRemote(x *Reifier) error {
	// This is normally done by SourceArgs
	x.String("cluster", cc.RemotingArgs.Cluster)

	// As per normal, do not forward VerboseArgs.
	return errors.Join(
		cc.DevArgs.ReifyForRemote(x),
		cc.ConfigFileArgs.ReifyForRemote(x),
		cc.HostArgs.ReifyForRemote(x),
		cc.FormatArgs.ReifyForRemote(x),
	)
}

func (cc *ConfigCommand) Validate() error {
	if cc.ConfigFilename == "" {
		ApplyDefault(&cc.Remote, "data-source", "remote")
		ApplyDefault(&cc.AuthFile, "data-source", "auth-file")
		ApplyDefault(&cc.Cluster, "data-source", "cluster")
	}

	return errors.Join(
		cc.DevArgs.Validate(),
		cc.HostArgs.Validate(),
		cc.RemotingArgs.Validate(),
		cc.VerboseArgs.Validate(),
		cc.ConfigFileArgs.Validate(),
		ValidateFormatArgs(
			&cc.FormatArgs, configsDefaultFields, configsFormatters, configsAliases, DefaultFixed),
	)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Analysis

func (cc *ConfigCommand) Configs(_ io.Reader, stdout, _ io.Writer) error {
	includeHosts, err := hostglob.NewGlobber(true, cc.HostArgs.Host)
	if err != nil {
		return err
	}

	cfg, err := db.MaybeGetConfig(cc.ConfigFile())
	if err != nil {
		return err
	}
	if cfg == nil {
		return errors.New("-config-file required")
	}

	// `records` is always freshly allocated
	records := cfg.Hosts()

	if !includeHosts.IsEmpty() {
		records = slices.DeleteFunc(records, func(r *config.NodeConfigRecord) bool {
			return !includeHosts.Match(r.Hostname)
		})
	}

	slices.SortFunc(records, func(a, b *config.NodeConfigRecord) int {
		return cmp.Compare(a.Hostname, b.Hostname)
	})

	FormatData(
		stdout,
		cc.FormatArgs.PrintFields,
		configsFormatters,
		cc.FormatArgs.PrintOpts,
		records,
		false,
	)

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Printing

func (cc *ConfigCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(
		cc.Fmt, configsHelp, configsFormatters, configsAliases, configsDefaultFields)
}

const configsHelp = `
configs
  Extract information about individual nodes on the cluster from config data and
  present them in primitive form.  Output records are sorted by node name.  The
  default format is 'fixed'.
`

type configCtx = bool
type configSummary = config.NodeConfigRecord

const configsDefaultFields = "host,cores,mem,gpus,gpumem,xnode,desc"

// MT: Constant after initialization; immutable
var configsAliases = map[string][]string{
	"default":     strings.Split(configsDefaultFields, ","),
	"hostname":    []string{"host"},
	"description": []string{"desc"},
	"ram":         []string{"mem"},
}

// MT: Constant after initialization; immutable
var configsFormatters = map[string]Formatter[*configSummary, configCtx]{
	"timestamp": {
		func(i *configSummary, _ configCtx) string {
			return i.Timestamp
		},
		"Timestamp of record (UTC)",
	},
	"host": {
		func(i *configSummary, _ configCtx) string {
			return i.Hostname
		},
		"Node name",
	},
	"desc": {
		func(i *configSummary, _ configCtx) string {
			return i.Description
		},
		"Human-consumable node summary",
	},
	"xnode": {
		func(i *configSummary, _ configCtx) string {
			if i.CrossNodeJobs {
				return "yes"
			}
			return "no"
		},
		"Can node participate in cross-node jobs?",
	},
	"cores": {
		func(i *configSummary, _ configCtx) string {
			return fmt.Sprint(i.CpuCores)
		},
		"Number of cores on the node (virtual cores)",
	},
	"mem": {
		func(i *configSummary, _ configCtx) string {
			return fmt.Sprint(i.MemGB)
		},
		"GB of physical RAM on the node",
	},
	"gpus": {
		func(i *configSummary, _ configCtx) string {
			return fmt.Sprint(i.GpuCards)
		},
		"Number of installed GPU cards on the node",
	},
	"gpumem": {
		func(i *configSummary, _ configCtx) string {
			return fmt.Sprint(i.GpuMemGB)
		},
		"GB of GPU RAM on the node (across all cards)",
	},
	"gpumempct": {
		func(i *configSummary, _ configCtx) string {
			if i.GpuMemPct {
				return "yes"
			}
			return "no"
		},
		"GPUs report memory in percentage",
	},
}
