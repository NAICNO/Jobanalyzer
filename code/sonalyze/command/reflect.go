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
	if opts.NoDefaults {
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
		func(fld reflect.StructField) (ok bool, name, desc string, aliases []string) {
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
// ComputedFormatSpec is used for synthesized fields: fields not in the data array.  It's good
// hygiene for them to be passed together with actual fields so that there are no duplicates in the
// name space.

type SimpleFormatSpec struct {
	Desc    string
	Aliases string
}

func ReflectFormattersFromMap(
	structTy reflect.Type,
	fields map[string]any,
) (formatters map[string]Formatter[any, PrintMods]) {
	formatters = make(map[string]Formatter[any, PrintMods])
	reflectStructFormatters(
		structTy,
		formatters,
		func(fld reflect.StructField) (ok bool, name, desc string, aliases []string) {
			name = fld.Name
			if spec, found := fields[name]; found {
				switch s := spec.(type) {
				case SimpleFormatSpec:
					desc = s.Desc
					aliases = strings.Split(s.Aliases, ",")
					ok = true
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
	admissible func(fld reflect.StructField) (ok bool, name, desc string, aliases []string),
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
			reflectStructFormatters(fldTy, subFormatters, admissible)
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
			if ok, name, desc, aliases := admissible(fld); ok {
				f := Formatter[any, PrintMods]{
					Help: desc,
					Fmt:  reflectTypeFormatter(ix, fld.Type),
				}
				formatters[name] = f
				for _, a := range aliases {
					formatters[a] = f
				}
			}
		}
	}
}

func reflectTypeFormatter(ix int, ty reflect.Type) func(any, PrintMods) string {
	switch {
	case ty == dateTimeValueTy || ty == dateTimeValueOrBlankTy:
		// Time formatters that must respect flags.
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
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
	case ty == ustrMax30Ty:
		return func(d any, ctx PrintMods) string {
			// Ustr is uint32
			val := Ustr(reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint())
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
			return func(d any, _ PrintMods) string {
				vals := reflect.Indirect(reflect.ValueOf(d)).Field(ix)
				lim := vals.Len()
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
		return func(d any, _ PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix)
			return val.MethodByName("String").Call(nil)[0].String()
		}
	default:
		// Everything else is a basic type that is handled according to kind.
		switch ty.Kind() {
		case reflect.Bool:
			return func(d any, _ PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Bool()
				// These are backwards compatible values.
				if val {
					return "yes"
				}
				return "no"
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return func(d any, _ PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
				return strconv.FormatInt(val, 10)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return func(d any, _ PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint()
				return strconv.FormatUint(val, 10)
			}
		case reflect.Float32, reflect.Float64:
			return func(d any, _ PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Float()
				return strconv.FormatFloat(val, 'g', -1, 64)
			}
		case reflect.String:
			return func(d any, _ PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).String()
				return val
			}
		default:
			panic(fmt.Sprintf("Unhandled type kind %d", ty.Kind()))
		}
	}
}
