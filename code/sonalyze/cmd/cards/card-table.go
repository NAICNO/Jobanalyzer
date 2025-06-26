// DO NOT EDIT.  Generated from cards.go by generate-table

package cards

import "sonalyze/db/repr"

import (
	"cmp"
	"fmt"
	"go-utils/gpuset"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
)

var (
	_ = cmp.Compare(0, 0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
	_ gpuset.GpuSet
)

// MT: Constant after initialization; immutable
var cardFormatters = map[string]Formatter[*repr.SysinfoCardData]{
	"Time": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Time), ctx)
		},
		Help: "(string) Full ISO timestamp of when the reading was taken",
	},
	"Node": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Node), ctx)
		},
		Help: "(string) Card's node at this time",
	},
	"Index": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.Index), ctx)
		},
		Help: "(uint64) Card's index on its node at this time",
	},
	"UUID": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.UUID), ctx)
		},
		Help: "(string) Card's unique identifier (but not necessarily its only unique identifier)",
	},
	"Address": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Address), ctx)
		},
		Help: "(string) Card's address on its node at this time",
	},
	"Manufacturer": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Manufacturer), ctx)
		},
		Help: "(string) Card's manufacturer's name",
	},
	"Model": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Model), ctx)
		},
		Help: "(string) Card model",
	},
	"Architecture": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Architecture), ctx)
		},
		Help: "(string) Card's architecture name",
	},
	"Driver": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Driver), ctx)
		},
		Help: "(string) Card driver's version at this time",
	},
	"Firmware": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatString((d.Firmware), ctx)
		},
		Help: "(string) Card firmware's version at this time",
	},
	"Memory": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.Memory), ctx)
		},
		Help: "(uint64) Card's memory in KB",
	},
	"PowerLimit": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.PowerLimit), ctx)
		},
		Help: "(uint64) Card's power limit at this time",
	},
	"MaxPowerLimit": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.MaxPowerLimit), ctx)
		},
		Help: "(uint64) Card's maximum power limit",
	},
	"MinPowerLimit": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.MinPowerLimit), ctx)
		},
		Help: "(uint64) Card's minimum power limit",
	},
	"MaxCEClock": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.MaxCEClock), ctx)
		},
		Help: "(uint64) Card's maximum compute element clock speed",
	},
	"MaxMemoryClock": {
		Fmt: func(d *repr.SysinfoCardData, ctx PrintMods) string {
			return FormatUint64((d.MaxMemoryClock), ctx)
		},
		Help: "(uint64) Card's maximum memory clock speed",
	},
}

// MT: Constant after initialization; immutable
var cardPredicates = map[string]Predicate[*repr.SysinfoCardData]{
	"Time": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Time), v.(string))
		},
	},
	"Node": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Node), v.(string))
		},
	},
	"Index": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Index), v.(uint64))
		},
	},
	"UUID": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.UUID), v.(string))
		},
	},
	"Address": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Address), v.(string))
		},
	},
	"Manufacturer": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Manufacturer), v.(string))
		},
	},
	"Model": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Model), v.(string))
		},
	},
	"Architecture": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Architecture), v.(string))
		},
	},
	"Driver": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Driver), v.(string))
		},
	},
	"Firmware": Predicate[*repr.SysinfoCardData]{
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Firmware), v.(string))
		},
	},
	"Memory": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.Memory), v.(uint64))
		},
	},
	"PowerLimit": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.PowerLimit), v.(uint64))
		},
	},
	"MaxPowerLimit": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.MaxPowerLimit), v.(uint64))
		},
	},
	"MinPowerLimit": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.MinPowerLimit), v.(uint64))
		},
	},
	"MaxCEClock": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.MaxCEClock), v.(uint64))
		},
	},
	"MaxMemoryClock": Predicate[*repr.SysinfoCardData]{
		Convert: CvtString2Uint64,
		Compare: func(d *repr.SysinfoCardData, v any) int {
			return cmp.Compare((d.MaxMemoryClock), v.(uint64))
		},
	},
}

func (c *CardCommand) Summary(out io.Writer) {
	fmt.Fprint(out, `  Print GPU card configuration data
`)
}

const cardHelp = `
card
  Extract information about individual gpu cards on the cluster from sysinfo and present it in
  primitive form.  Output records are sorted by time and node name, note cards can be moved between
  nodes from time to time.  The default format is 'fixed'.
`

func (c *CardCommand) MaybeFormatHelp() *FormatHelp {
	return StandardFormatHelp(c.Fmt, cardHelp, cardFormatters, cardAliases, cardDefaultFields)
}

// MT: Constant after initialization; immutable
var cardAliases = map[string][]string{
	"Default": []string{"Node", "Index", "Manufacturer", "Model", "Memory"},
	"All":     []string{"Time", "Node", "Index", "UUID", "Address", "Manufacturer", "Model", "Architecture", "Driver", "Firmware", "Memory", "PowerLimit", "MaxPowerLimit", "MinPowerLimit", "MaxCEClock", "MaxMemoryClock"},
}

const cardDefaultFields = "Default"
