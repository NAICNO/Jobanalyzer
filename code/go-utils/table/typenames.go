package table

import (
	"log"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Types we know about and information about them.
//
// The formatter for type Ty is usually FormatTy(Ty, PrintMods) -> string, but the name can be
// overridden by registering the type.
//
// The user-facing type name for type Ty is usually Ty but it can be overridden. BEWARE that this
// override has implications for the applicable query operators: if a type is "string" then string
// operators apply; if a type is "GpuSet" then some kind of set operators apply (TBD).
//
// The setComparer is func(a,b set, op int) -> bool s.t. the return value is true iff the operation
// is satisfied.  The operation is opaque to the generated code: it originates within the query
// logic, and is consumed there by the setComparer.

type TypeInfo struct {
	HelpName    string // default is the name as given
	Comparer    string // setType == false: default is cmp.Compare
	Formatter   string // default is Format<Typename>
	Parser      string // default is CvtString2<Typename>
	SetComparer string // if "", not a set; otherwise a function
}

var KnownTypes = map[string]TypeInfo{
	"bool": TypeInfo{
		Comparer: "CompareBool",
	},
	"[]string": TypeInfo{
		HelpName:    "string list",
		Formatter:   "FormatStrings",
		Parser:      "CvtString2Strings",
		SetComparer: "SetCompareStrings",
	},
	"F64Ceil": TypeInfo{
		HelpName: "int",
		Parser:   "CvtString2Float64",
	},
	"U64Div1M": TypeInfo{
		HelpName: "int",
		Parser:   "CvtString2Uint64",
	},
	"IntOrEmpty": TypeInfo{
		HelpName: "int",
		Parser:   "CvtString2Int",
	},
	"DateTimeValueOrBlank": TypeInfo{
		HelpName: "DateTimeValue",
		Parser:   "CvtString2DateTimeValue",
	},
	"IsoDateTimeOrUnknown": TypeInfo{HelpName: "IsoDateTimeValue"},
	"Ustr":                 TypeInfo{HelpName: "string"},
	"UstrMax30":            TypeInfo{HelpName: "string"},
	"gpuset.GpuSet": TypeInfo{
		HelpName:    "GpuSet",
		Formatter:   "FormatGpuSet",
		Parser:      "CvtString2GpuSet",
		SetComparer: "SetCompareGpuSets",
	},
	"*Hostnames": TypeInfo{
		HelpName:    "Hostnames",
		Formatter:   "FormatHostnames",
		Parser:      "CvtString2Hostnames",
		SetComparer: "SetCompareHostnames",
	},
}

func IsComparable(ty string) bool {
	if probe, found := KnownTypes[ty]; found {
		return probe.SetComparer == ""
	}
	return true
}

func FieldComparer(ty string) string {
	if probe, found := KnownTypes[ty]; found && probe.Comparer != "" {
		return probe.Comparer
	}
	return "cmp.Compare"
}

func SetComparer(ty string) string {
	if probe, found := KnownTypes[ty]; found && probe.SetComparer != "" {
		return probe.SetComparer
	}
	log.Fatalf("Not a set: %s", ty)
	return ""
}

func IsSetType(ty string) bool {
	if probe, found := KnownTypes[ty]; found {
		return probe.SetComparer != ""
	}
	return false
}

func FormatName(ty string) string {
	if probe := KnownTypes[ty]; probe.Formatter != "" {
		return probe.Formatter
	}
	return "Format" + Capitalize(ty)
}

func ParseName(ty string) string {
	if probe := KnownTypes[ty]; probe.Parser != "" {
		return probe.Parser
	}
	return "CvtString2" + Capitalize(ty)
}

func UserFacingTypeName(ty string) string {
	if probe := KnownTypes[ty]; probe.HelpName != "" {
		return probe.HelpName
	}
	// TODO: Strip suffix size information
	return ty
}
