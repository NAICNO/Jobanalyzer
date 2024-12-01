// DO NOT EDIT.  Generated from nodes.go by sonalyze-table

package nodes
import (
    "go-utils/config"
    . "sonalyze/table"
)

// MT: Constant after initialization; immutable
var nodesFormatters = map[string]Formatter{
    "Timestamp": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatString(d.(*config.NodeConfigRecord).Timestamp, ctx)
        },
        Help: "Full ISO timestamp of when the reading was taken",
    },
    "Hostname": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatString(d.(*config.NodeConfigRecord).Hostname, ctx)
        },
        Help: "Name that host is known by on the cluster",
    },
    "Description": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatString(d.(*config.NodeConfigRecord).Description, ctx)
        },
        Help: "End-user description, not parseable",
    },
    "CpuCores": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatInt(d.(*config.NodeConfigRecord).CpuCores, ctx)
        },
        Help: "Total number of cores x threads",
    },
    "MemGB": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatInt(d.(*config.NodeConfigRecord).MemGB, ctx)
        },
        Help: "GB of installed main RAM",
    },
    "GpuCards": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatInt(d.(*config.NodeConfigRecord).GpuCards, ctx)
        },
        Help: "Number of installed cards",
    },
    "GpuMemGB": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatInt(d.(*config.NodeConfigRecord).GpuMemGB, ctx)
        },
        Help: "Total GPU memory across all cards",
    },
    "GpuMemPct": {
        Fmt: func(d any, ctx PrintMods) string {
            return FormatBool(d.(*config.NodeConfigRecord).GpuMemPct, ctx)
        },
        Help: "True if GPUs report accurate memory usage in percent",
    },
}

func init() {
    DefAlias(nodesFormatters, "Timestamp", "timestamp")
    DefAlias(nodesFormatters, "Hostname", "host")
    DefAlias(nodesFormatters, "Description", "desc")
    DefAlias(nodesFormatters, "CpuCores", "cores")
    DefAlias(nodesFormatters, "MemGB", "mem")
    DefAlias(nodesFormatters, "GpuCards", "gpus")
    DefAlias(nodesFormatters, "GpuMemGB", "gpumem")
    DefAlias(nodesFormatters, "GpuMemPct", "gpumempct")
}
