// Generate formatters from an annotated structure types using reflection.
//
// ReflectFormattersFromTags generates formatters from the tagged fields on a structure type,
// excluding any field names in an optional blocklist.
//
// ReflectFormattersFromMap generates formatters from the fields on a structure type if they
// appear in an allowlist.
//
// Both will descend into embedded fields.
//
// TODO: At the moment they do not handle circular structures, but they could (and should).

package command

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"go-utils/gpuset"

	. "sonalyze/common"
)

////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Annotation types
//
// These types hold an int64 unix timestamp and indicate a particular kind of formatting.
//
// TODO: DateTimeValue and DateTimeValueOrBlank will eventually honor /iso and /sec modifiers.
// TODO: Possibly IsoDateTimeOrUnknown should honor /sec at least.

type DateTimeValue int64        // yyyy-mm-dd hh:mm
type DateTimeValueOrBlank int64 // yyyy-mm-dd hh:mm or 16 blanks
type IsoDateTimeOrUnknown int64 // yyyy-mm-ddThh:mmZhh:mm
type DateValue int64            // yyyy-mm-dd
type TimeValue int64            // hh:mm

func (val DateValue) String() string {
	return time.Unix(int64(val), 0).UTC().Format("2006-01-02")
}

func (val TimeValue) String() string {
	return time.Unix(int64(val), 0).UTC().Format("15:04")
}

func (val IsoDateTimeOrUnknown) String() string {
	if val == 0 {
		return "Unknown"
	}
	return time.Unix(int64(val), 0).UTC().Format(time.RFC3339)
}

// More of the same

type IntOrEmpty int // the int value, but "" if zero
type UstrMax30 Ustr // the string value but only max 30 first chars in fixed mode

func (val IntOrEmpty) String() string {
	if val == 0 {
		return ""
	}
	return strconv.FormatInt(int64(val), 10)
}

// Type representations of some of those, for when we need the types.

var (
	// TODO: When we can adopt Go 1.22: valTy := reflect.TypeFor[TypeName]()
	dummyDateTimeValue DateTimeValue
	dateTimeValueTy    = reflect.TypeOf(dummyDateTimeValue)

	dummyDateTimeValueOrBlank DateTimeValueOrBlank
	dateTimeValueOrBlankTy    = reflect.TypeOf(dummyDateTimeValueOrBlank)

	dummyUstrMax30Value UstrMax30
	ustrMax30Ty         = reflect.TypeOf(dummyUstrMax30Value)

	dummyGpuSetValue gpuset.GpuSet
	gpuSetTy         = reflect.TypeOf(dummyGpuSetValue)

	// See the Example for reflect.TypeOf in the Go documentation.
	stringerTy = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
)

////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Print modifiers.
//
// See comment block below.

type PrintMods = int

const (
	// These apply per-field according to modifiers
	// TODO: These are currently unimplemented.
	PrintModSec = (1 << iota) // timestamps are printed as seconds since epoch
	PrintModIso               // timestamps are printed as Iso timestamps

	// These are for the output format and are applied to all fields
	PrintModFixed      // fixed format
	PrintModJson       // JSON format
	PrintModCsv        // CSV format
	PrintModCsvNamed   // CSVNamed format
	PrintModAwk        // AWK format
	PrintModNoDefaults // Set (for any format) if option is set
)

// This is a temporary solution.  Currently the callers of FormatData must pass this when they use
// reflected formatters, but once all modules have reflected formatters this will instead be
// computed by FormatData.  At that point, the callers of FormatData will no longer pass their last
// argument.

func ComputePrintMods(opts *FormatOptions) PrintMods {
	var x PrintMods
	switch {
	case opts.Csv:
		if opts.Named {
			x = PrintModCsvNamed
		} else {
			x = PrintModCsv
		}
	case opts.Json:
		x = PrintModJson
	case opts.Awk:
		x = PrintModAwk
	case opts.Fixed:
		x = PrintModFixed
	}
	if opts.NoDefaults && (opts.Csv || opts.Json) {
		x |= PrintModNoDefaults
	}
	return x
}

// Given a struct type, ReflectFormattersFromTags constructs a map from field names to a formatter
// for each field.  Fields are excluded if they appear in isExcluded or have no `desc` annotation.
//
// A field may have an `alias` annotation in addition to its name.  The alias is treated just as the
// name.  Aliases are a consequence of older code using "convenient" names for fields while we want
// to move to a world where fields are named in a transparent and uniform way.  Clients can ask for
// the real name or the alias.  Default lists in client code can refer to whatever fields they want.
// The `alias` annotation is a comma-separated list of alias names.  Fields are excluded if any of
// their aliases appear in isExcluded.
//
// There must be no duplicates in the union of field names and aliases, or in the set of aliases.
//
// The formtting function's ctx is a bit flag vector, the flags can vary with the field b/c
// formatting specifiers like "/sec".  This is not yet well developed but will work OK.

func ReflectFormattersFromTags(
	structTy reflect.Type,
	isExcluded map[string]bool,
) (formatters map[string]Formatter[any, PrintMods]) {
	formatters = make(map[string]Formatter[any, PrintMods])
	reflectStructFormatters(
		structTy,
		formatters,
		func(fld reflect.StructField) (ok bool, name, desc string, aliases []string, attrs int) {
			name = fld.Name
			if isExcluded[name] {
				return
			}
			aliases = strings.Split(fld.Tag.Get("alias"), ",")
			for _, a := range aliases {
				if isExcluded[a] {
					return
				}
			}
			desc = fld.Tag.Get("desc")
			if desc == "" {
				return
			}
			ok = true
			return
		},
		nil,
	)
	return
}

// ReflectFormattersFromMap is like ReflectFormattersFromTags, but instead of being passed a
// blocklist, it is being passed an allowlist, along with formatting information for the fields on
// that list.  This is useful when we want to decouple the specification of formatting from the
// structure definition, as when there are multiple formatters for the same data but they extract
// different fields and with different formatting rules, or when we simply do not want to specify
// any formatting rules in the structure definition because the structure definition is in a package
// different from the printing.
//
// The field values must be one of the *FormatSpec types below.
//
// SimpleFormatSpecWithAttr uses an attribute to specify a simple formatting rule, that in the case
// of ReflectFormattersFromTags can be expressed through a type.
//
// SynthesizedFormatSpecWithAttr uses an attribute to specify a simple formatting rule for a
// synthesized output field.
//
// (Unused and maybe unneeded) ComputedFormatSpec is used for synthesized fields: fields not in the data array.

type SimpleFormatSpec struct {
	Desc    string
	Aliases string
}

// `FmtDefaultable` indicates that the field has a default value to skip if nodefaults is set.
// Numbers, Ustr, string, and GpuSet are defaultable.
//
// For Fmt<Typename> see typename at top of file

const (
	FmtDefaultable      = (1 << iota)
	FmtDateTimeValue    // type must be int64
	FmtIsoDateTimeValue // type must be int64
	FmtDivideBy1M       // type must be integer
	// There could be more, just as there are various types to express the same thing
)

type SimpleFormatSpecWithAttr struct {
	Desc    string
	Aliases string
	Attr    int
}

type SynthesizedFormatSpecWithAttr struct {
	Desc     string
	RealName string
	Attr     int
}

type ComputedFormatSpec struct {
	RealName string // field name this is derived from
	Desc     string
	Fmt      func(d any) string // d will be a pointer value
}

func ReflectFormattersFromMap(
	structTy reflect.Type,
	fields map[string]any,
) (formatters map[string]Formatter[any, PrintMods]) {
	formatters = make(map[string]Formatter[any, PrintMods])

	// `synthesizable` is a map from a real field name to one synthesized/computed spec that targets
	// that real field name.
	type SynthSpec struct {
		SynthesizedName string
		spec            any
	}
	synthesizable := make(map[string]SynthSpec)
	for k, v := range fields {
		var realname string
		switch s := v.(type) {
		case SynthesizedFormatSpecWithAttr:
			realname = s.RealName
		case ComputedFormatSpec:
			realname = s.RealName
		default:
			continue
		}
		if _, found := synthesizable[realname]; found {
			// We can lift this restriction if we have to
			panic(fmt.Sprintf("Multiple synthesized fields targeting '%s'", realname))
		}
		synthesizable[realname] = SynthSpec{k, v}
	}

	reflectStructFormatters(
		structTy,
		formatters,
		func(fld reflect.StructField) (ok bool, name, desc string, aliases []string, attrs int) {
			name = fld.Name
			if spec, found := fields[name]; found {
				switch s := spec.(type) {
				case SimpleFormatSpec:
					desc = s.Desc
					aliases = strings.Split(s.Aliases, ",")
					ok = true
				case SimpleFormatSpecWithAttr:
					desc = s.Desc
					aliases = strings.Split(s.Aliases, ",")
					attrs = s.Attr
					ok = true
				case SynthesizedFormatSpecWithAttr:
					panic(fmt.Sprintf("Struct field '%s' has synthesized spec", name))
				case ComputedFormatSpec:
					panic(fmt.Sprintf("Struct field '%s' has computed spec", name))
				default:
					panic("Invalid FormatSpec")
				}
			}
			return
		},
		func(fld reflect.StructField) (ok bool, name, desc string, attrs int) {
			if spec, found := synthesizable[fld.Name]; found {
				switch info := spec.spec.(type) {
				case SynthesizedFormatSpecWithAttr:
					name = spec.SynthesizedName
					desc = info.Desc
					attrs = info.Attr
					ok = true
				case ComputedFormatSpec:
					panic("NYI")
				default:
					panic("Invalid FormatSpec")
				}
			}
			return
		},
	)
	return
}

func reflectStructFormatters(
	structTy reflect.Type,
	formatters map[string]Formatter[any, PrintMods],
	admissible func(fld reflect.StructField) (ok bool, name, desc string, aliases []string, attrs int),
	synthesizable func(fld reflect.StructField) (ok bool, name, desc string, attrs int),
) {
	if structTy.Kind() != reflect.Struct {
		panic("Struct type required")
	}
	for i, lim := 0, structTy.NumField(); i < lim; i++ {
		// TODO: once we move to Go 1.22: no temp binding
		ix := i
		fld := structTy.Field(ix)
		if fld.Anonymous {
			// Trace through embedded field.  The formatting function will receive the outer
			// structure (or pointer to it), but the formatter generator code operates on the inner
			// struct.  We could change everything to use FieldByName but don't want to do that.  So
			// we wrap each returned formatting function in a function that obtains the field value
			// if it is a pointer and otherwise construct a pointer to the field, and then pass that
			// pointer to the generated formatter.
			fldTy := fld.Type
			mustTakeAddress := false
			if fldTy.Kind() == reflect.Struct {
				mustTakeAddress = true
			} else if fldTy.Kind() == reflect.Pointer && fldTy.Elem().Kind() == reflect.Struct {
				fldTy = fldTy.Elem()
			} else {
				continue
			}
			subFormatters := make(map[string]Formatter[any, PrintMods])
			reflectStructFormatters(fldTy, subFormatters, admissible, synthesizable)
			for name, fmt := range subFormatters {
				// TODO: once we move to Go 1.22: no temp binding
				theFmt := fmt.Fmt
				f := Formatter[any, PrintMods]{
					Help: fmt.Help,
				}
				if mustTakeAddress {
					f.Fmt = func(d any, mods PrintMods) string {
						val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Addr()
						return theFmt(val.Interface(), mods)
					}
				} else {
					f.Fmt = func(d any, mods PrintMods) string {
						val := reflect.Indirect(reflect.ValueOf(d)).Field(ix)
						return theFmt(val.Interface(), mods)
					}
				}
				formatters[name] = f
			}
		} else {
			if ok, name, desc, aliases, attrs := admissible(fld); ok {
				f := Formatter[any, PrintMods]{
					Help: desc,
					Fmt:  reflectTypeFormatter(ix, attrs, fld.Type),
				}
				formatters[name] = f
				for _, a := range aliases {
					formatters[a] = f
				}
			}

			if synthesizable != nil {
				if ok, name, desc, attrs := synthesizable(fld); ok {
					// synthesizable(fld) returns info for a synthesized field that addresses fld.name.
					//
					// TODO: In principle there could be multiple synthesizable fields per fld, for now
					// we'll require synthesizable (or earlier code) to panic if that case occurs.
					f := Formatter[any, PrintMods]{
						Help: desc,
						Fmt:  reflectTypeFormatter(ix, attrs, fld.Type),
					}
					formatters[name] = f
				}
			}
		}
	}
}

func reflectTypeFormatter(ix int, attrs int, ty reflect.Type) func(any, PrintMods) string {
	switch {
	case ty == dateTimeValueTy || ty == dateTimeValueOrBlankTy:
		// Time formatters that must respect flags.
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
			if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == 0 {
				return "*skip*"
			}
			switch {
			case val == 0 && ty == dateTimeValueOrBlankTy:
				return "                "
			case (ctx & PrintModSec) != 0:
				return strconv.FormatInt(val, 10)
			case (ctx & PrintModIso) != 0:
				return time.Unix(val, 0).UTC().Format(time.RFC3339)
			default:
				return FormatYyyyMmDdHhMmUtc(val)
			}
		}
	case ty == gpuSetTy:
		return func(d any, ctx PrintMods) string {
			// GpuSet is uint32
			val := Ustr(reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint())
			set := gpuset.GpuSet(val)
			if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && set.IsEmpty() {
				return "*skip*"
			}
			return set.String()
		}
	case ty == ustrMax30Ty:
		return func(d any, ctx PrintMods) string {
			// Ustr is uint32
			val := Ustr(reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint())
			if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == UstrEmpty {
				return "*skip*"
			}
			s := val.String()
			if (ctx & PrintModFixed) != 0 {
				// TODO: really the rune length, no?
				if len(s) > 30 {
					return s[:30]
				}
			}
			return s
		}
	case ty.Kind() == reflect.Slice:
		// Slices are a little weird now, only []string.  We sort them before printing, this is
		// generally the right thing (and gives us stable output).  There's a risk that there are
		// slices that should not be sorted b/c other elements have indices into them, but there are
		// none of that kind right now.
		switch ty.Elem().Kind() {
		case reflect.String:
			return func(d any, ctx PrintMods) string {
				vals := reflect.Indirect(reflect.ValueOf(d)).Field(ix)
				lim := vals.Len()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && lim == 0 {
					return "*skip*"
				}
				ss := make([]string, lim)
				for j := 0; j < lim; j++ {
					ss[j] = vals.Index(j).String()
				}
				slices.Sort(ss)
				return strings.Join(ss, ",")
			}
		default:
			panic("NYI")
		}
	case ty.Implements(stringerTy):
		// If it implements fmt.Stringer then use it
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix)
			s := val.MethodByName("String").Call(nil)[0].String()
			if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && s == "" {
				return "*skip*"
			}
			return s
		}
	default:
		// Everything else is a basic type that is handled according to kind.
		switch ty.Kind() {
		case reflect.Bool:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Bool()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && !val {
					return "*skip*"
				}
				// These are backwards compatible values.
				if val {
					return "yes"
				}
				return "no"
			}
		case reflect.Int64:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == 0 {
					return "*skip*"
				}
				if (attrs & FmtDateTimeValue) != 0 {
					return FormatYyyyMmDdHhMmUtc(val)
				}
				if (attrs & FmtIsoDateTimeValue) != 0 {
					return time.Unix(val, 0).UTC().Format(time.RFC3339)
				}
				if (attrs & FmtDivideBy1M) != 0 {
					val /= 1024 * 1024
				}
				return strconv.FormatInt(val, 10)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == 0 {
					return "*skip*"
				}
				if (attrs & FmtDivideBy1M) != 0 {
					val /= 1024 * 1024
				}
				return strconv.FormatInt(val, 10)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == 0 {
					return "*skip*"
				}
				if (attrs & FmtDivideBy1M) != 0 {
					val /= 1024 * 1024
				}
				return strconv.FormatUint(val, 10)
			}
		case reflect.Float32, reflect.Float64:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Float()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == 0 {
					return "*skip*"
				}
				prec := 64
				if ty.Kind() == reflect.Float32 {
					prec = 32
				}
				return strconv.FormatFloat(val, 'g', -1, prec)
			}
		case reflect.String:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).String()
				if (attrs&FmtDefaultable) != 0 && (ctx&PrintModNoDefaults) != 0 && val == "" {
					return "*skip*"
				}
				return val
			}
		default:
			panic(fmt.Sprintf("Unhandled type kind %d", ty.Kind()))
		}
	}
}