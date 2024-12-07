// DO NOT EDIT.  Generated from configs.go by generate-table

package configs

import (
	"go-utils/config"
	. "sonalyze/table"
)

// MT: Constant after initialization; immutable
var configFormatters = map[string]Formatter[*config.NodeConfigRecord]{
	"Timestamp": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatString(string(d.Timestamp), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Hostname": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatString(string(d.Hostname), ctx)
		},
		Help: "(string) Name that host is known by on the cluster",
	},
	"Description": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatString(string(d.Description), ctx)
		},
		Help: "(string) End-user description, not parseable",
	},
	"CrossNodeJobs": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatBool(bool(d.CrossNodeJobs), ctx)
		},
		Help: "(bool) True if jobs on this node can be multi-node",
	},
	"CpuCores": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.CpuCores), ctx)
		},
		Help: "(int) Total number of cores x threads",
	},
	"MemGB": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.MemGB), ctx)
		},
		Help: "(int) GB of installed main RAM",
	},
	"GpuCards": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.GpuCards), ctx)
		},
		Help: "(int) Number of installed cards",
	},
	"GpuMemGB": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatInt(int(d.GpuMemGB), ctx)
		},
		Help: "(int) Total GPU memory across all cards",
	},
	"GpuMemPct": {
		Fmt: func(d *config.NodeConfigRecord, ctx PrintMods) string {
			return FormatBool(bool(d.GpuMemPct), ctx)
		},
		Help: "(bool) True if GPUs report accurate memory usage in percent",
	},
}

func init() {
	DefAlias(configFormatters, "Timestamp", "timestamp")
	DefAlias(configFormatters, "Hostname", "host")
	DefAlias(configFormatters, "Description", "desc")
	DefAlias(configFormatters, "CrossNodeJobs", "xnode")
	DefAlias(configFormatters, "CpuCores", "cores")
	DefAlias(configFormatters, "MemGB", "mem")
	DefAlias(configFormatters, "GpuCards", "gpus")
	DefAlias(configFormatters, "GpuMemGB", "gpumem")
	DefAlias(configFormatters, "GpuMemPct", "gpumempct")
}

const configHelp = `
config
  Extract information about individual nodes on the cluster from config data and
  present them in primitive form.  Output records are sorted by node name.  The
  default format is 'fixed'.
`

func (c *ConfigCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, configHelp, configFormatters, configAliases, configDefaultFields)
}

// MT: Constant after initialization; immutable
var configAliases = map[string][]string{
	"default": []string{"host", "cores", "mem", "gpus", "gpumem", "desc"},
	"Default": []string{"Hostname", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "Description"},
	"all":     []string{"timestamp", "host", "desc", "xnode", "cores", "mem", "gpus", "gpumem", "gpumempct"},
	"All":     []string{"Timestamp", "Hostname", "Description", "CrossNodeJobs", "CpuCores", "MemGB", "GpuCards", "GpuMemGB", "GpuMemPct"},
}

const configDefaultFields = "default"
