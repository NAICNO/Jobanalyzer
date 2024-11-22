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
	"io"
	"reflect"
	"slices"
	"strings"

	"go-utils/config"
	"go-utils/hostglob"
	uslices "go-utils/slices"

	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	. "sonalyze/table"
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

func (cc *ConfigCommand) Add(fs *CLI) {
	cc.DevArgs.Add(fs)
	cc.RemotingArgs.Add(fs)
	cc.HostArgs.Add(fs)
	cc.VerboseArgs.Add(fs)
	cc.ConfigFileArgs.Add(fs)
	cc.FormatArgs.Add(fs)
}

func (cc *ConfigCommand) ReifyForRemote(x *ArgReifier) error {
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

func (cc *ConfigCommand) Perform(_ io.Reader, stdout, _ io.Writer) error {
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
		cc.PrintFields,
		configsFormatters,
		cc.PrintOpts,
		uslices.Map(records, func(x *config.NodeConfigRecord) any { return x }),
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
config
  Extract information about individual nodes on the cluster from config data and
  present them in primitive form.  Output records are sorted by node name.  The
  default format is 'fixed'.
`

const (
	v0ConfigsDefaultFields = "host,cores,mem,gpus,gpumem,xnode,desc"
	v1ConfigsDefaultFields = "Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,CrossNodeJobs,Description"
	configsDefaultFields   = v0ConfigsDefaultFields
)

// MT: Constant after initialization; immutable
var configsAliases = map[string][]string{
	"default":   strings.Split(configsDefaultFields, ","),
	"v0default": strings.Split(v0ConfigsDefaultFields, ","),
	"v1default": strings.Split(v1ConfigsDefaultFields, ","),
}

type SFS = SimpleFormatSpec

// MT: Constant after initialization; immutable
var configsFormatters = DefineTableFromMap(
	reflect.TypeFor[config.NodeConfigRecord](),
	map[string]any{
		"Timestamp":     SFS{"Full ISO timestamp of when the reading was taken", "timestamp"},
		"Hostname":      SFS{"Name that host is known by on the cluster", "host"},
		"Description":   SFS{"End-user description, not parseable", "desc"},
		"CrossNodeJobs": SFS{"True if jobs on this node can be multi-node", "xnode"},
		"CpuCores":      SFS{"Total number of cores x threads", "cores"},
		"MemGB":         SFS{"GB of installed main RAM", "mem"},
		"GpuCards":      SFS{"Number of installed cards", "gpus"},
		"GpuMemGB":      SFS{"Total GPU memory across all cards", "gpumem"},
		"GpuMemPct":     SFS{"True if GPUs report accurate memory usage in percent", "gpumempct"},
	},
)
