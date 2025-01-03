// OBSOLETE, WILL BE REMOVED EVENTUALLY.  CODE SHOULD USE DECLARATIVE TABLES AND generate-table TO
// IMPLEMENT TABLE LOGIC, NOT REFLECTION.
//
// Generate a table definition (currently just a map of formatters) from an annotated structure
// types using reflection.
//
// DefineTableFromTags generates the table from the tagged fields on a structure type, excluding any
// field names in an optional blocklist.
//
// DefineTableFromMap generates the table from the fields on a structure type if they appear in an
// allowlist.
//
// Both will descend into embedded fields.

package table

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"

	"go-utils/gpuset"

	. "sonalyze/common"
)

var (
	dateTimeValueTy        = reflect.TypeFor[DateTimeValue]()
	dateTimeValueOrBlankTy = reflect.TypeFor[DateTimeValueOrBlank]()
	isoDateTimeOrUnknownTy = reflect.TypeFor[IsoDateTimeOrUnknown]()
	durationValueTy        = reflect.TypeFor[DurationValue]()
	ustrMax30Ty            = reflect.TypeFor[UstrMax30]()
	gpuSetTy               = reflect.TypeFor[gpuset.GpuSet]()
	stringerTy             = reflect.TypeFor[fmt.Stringer]()
)

// Given a struct type, DefineTableFromTags constructs a map from field names to a formatter for
// each field.  Fields are excluded if they appear in isExcluded or have no `desc` annotation.
//
// A field may have an `alias` annotation in addition to its name.  The `alias` annotation is a
// comma-separated list of alias names.  Fields are excluded if any of their aliases appear in
// isExcluded.
//
// There must be no duplicates in the union of field names and aliases, or in the set of aliases.
//
// The formtting function's ctx is a bit flag vector, the flags can vary with the field b/c
// formatting specifiers like "/sec" and "/iso".

func DefineTableFromTags(
	structTy reflect.Type,
	isExcluded map[string]bool,
) (formatters map[string]Formatter[any]) {
	formatters = make(map[string]Formatter[any])
	reflectStructFormatters(
		structTy,
		formatters,
		func(fld reflect.StructField) (ok bool, name, desc string, aliases []string, attrs int) {
			name = fld.Name
			if isExcluded[name] {
				return
			}
			aliasStr := fld.Tag.Get("alias")
			if aliasStr != "" {
				aliases = strings.Split(aliasStr, ",")
			}
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

// DefineTableFromMap is like DefineTableFromTags, but instead of being passed a blocklist, it is
// being passed an allowlist, along with formatting information for the fields on that list.  This
// is useful when we want to decouple the specification of formatting from the structure definition,
// as when there are multiple formatters for the same data but they extract different fields and
// with different formatting rules, or when we simply do not want to specify any formatting rules in
// the structure definition because the structure definition is in a package different from the
// printing.
//
// The field values must be one of the *FormatSpec types below.
//
// SimpleFormatSpecWithAttr uses an attribute to specify a simple formatting rule that in the case
// of DefineTableFromTags would be expressed through a type.
//
// SynthesizedFormatSpecWithAttr uses an attribute to specify a simple formatting rule for a
// synthesized output field computed from a real field.

type SimpleFormatSpec struct {
	Desc    string
	Aliases string
}

// `FmtCeil` and `FmtDivideBy1M` apply simple numeric transformations.  (There could be more.)
//
// For `Fmt<Typename>` see Typename in data.go - these attributes request that values be formatted
// as for those types.

const (
	FmtCeil             = (1 << iota) // type must be floating, take ceil, print as integer
	FmtDivideBy1M                     // type must be integer, integer divide by 1024*1024
	FmtDateTimeValue                  // type must be int64
	FmtIsoDateTimeValue               // type must be int64
	FmtDurationValue                  // type must be int64
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

func DefineTableFromMap(
	structTy reflect.Type,
	fields map[string]any,
) (formatters map[string]Formatter[any]) {
	formatters = make(map[string]Formatter[any])

	// `synthesizable` is a map from a real field name to one synthesized spec that targets
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
					if s.Aliases != "" {
						aliases = strings.Split(s.Aliases, ",")
					}
					ok = true
				case SimpleFormatSpecWithAttr:
					desc = s.Desc
					if s.Aliases != "" {
						aliases = strings.Split(s.Aliases, ",")
					}
					attrs = s.Attr
					ok = true
				case SynthesizedFormatSpecWithAttr:
					panic(fmt.Sprintf("Struct field '%s' has synthesized spec", name))
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
	formatters map[string]Formatter[any],
	admissible func(fld reflect.StructField) (ok bool, name, desc string, aliases []string, attrs int),
	synthesizable func(fld reflect.StructField) (ok bool, name, desc string, attrs int),
) {
	if structTy.Kind() != reflect.Struct {
		panic("Struct type required")
	}
	for i, lim := 0, structTy.NumField(); i < lim; i++ {
		fld := structTy.Field(i)
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
			subFormatters := make(map[string]Formatter[any])
			reflectStructFormatters(fldTy, subFormatters, admissible, synthesizable)
			for name, fmt := range subFormatters {
				f := Formatter[any]{
					Help:    fmt.Help,
					AliasOf: fmt.AliasOf,
				}
				if mustTakeAddress {
					f.Fmt = func(d any, mods PrintMods) string {
						val := reflect.Indirect(reflect.ValueOf(d)).Field(i).Addr()
						return fmt.Fmt(val.Interface(), mods)
					}
				} else {
					f.Fmt = func(d any, mods PrintMods) string {
						val := reflect.Indirect(reflect.ValueOf(d)).Field(i)
						return fmt.Fmt(val.Interface(), mods)
					}
				}
				formatters[name] = f
			}
		} else {
			if ok, name, desc, aliases, attrs := admissible(fld); ok {
				f := Formatter[any]{
					Help: desc,
					Fmt:  reflectTypeFormatter(i, attrs, fld.Type),
				}
				formatters[name] = f
				if len(aliases) > 0 {
					fa := f
					fa.AliasOf = name
					for _, a := range aliases {
						formatters[a] = fa
					}
				}
			}
			if synthesizable != nil {
				if ok, name, desc, attrs := synthesizable(fld); ok {
					// synthesizable(fld) returns info for a synthesized field that addresses fld.name.
					//
					// TODO: In principle there could be multiple synthesizable fields per fld, for now
					// we'll require synthesizable (or earlier code) to panic if that case occurs.
					f := Formatter[any]{
						Help: desc,
						Fmt:  reflectTypeFormatter(i, attrs, fld.Type),
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
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
			if val == 0 && ty == dateTimeValueOrBlankTy {
				return "                "
			}
			return FormatDateTimeValue(DateTimeValue(val), ctx)
		}
	case ty == isoDateTimeOrUnknownTy:
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
			if val == 0 {
				return "Unknown"
			}
			return FormatDateTimeValue(DateTimeValue(val), ctx|PrintModIso)
		}
	case ty == durationValueTy:
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
			return FormatDurationValue(DurationValue(val), ctx)
		}
	case ty == gpuSetTy:
		return func(d any, ctx PrintMods) string {
			// GpuSet is uint32
			val := Ustr(reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint())
			return FormatGpuSet(gpuset.GpuSet(val), ctx)
		}
	case ty == ustrMax30Ty:
		return func(d any, ctx PrintMods) string {
			// Ustr is uint32
			val := Ustr(reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint())
			if (ctx&PrintModNoDefaults) != 0 && val == UstrEmpty {
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
				if (ctx&PrintModNoDefaults) != 0 && lim == 0 {
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
			panic("NYI - non-string slice")
		}
	case ty.Implements(stringerTy):
		// If it implements fmt.Stringer then use it
		return func(d any, ctx PrintMods) string {
			val := reflect.Indirect(reflect.ValueOf(d)).Field(ix)
			s := val.MethodByName("String").Call(nil)[0].String()
			return FormatString(s, ctx)
		}
	default:
		// Everything else is a basic type that is handled according to kind.
		switch ty.Kind() {
		case reflect.Bool:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Bool()
				return FormatBool(val, ctx)
			}
		case reflect.Int64:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
				// This will scale date values too, but that's considered a user error.
				if (attrs & FmtDivideBy1M) != 0 {
					val /= 1024 * 1024
				}
				if (attrs & FmtDateTimeValue) != 0 {
					return FormatDateTimeValue(DateTimeValue(val), ctx)
				}
				if (attrs & FmtIsoDateTimeValue) != 0 {
					return FormatDateTimeValue(DateTimeValue(val), ctx|PrintModIso)
				}
				if (attrs & FmtDurationValue) != 0 {
					return FormatDurationValue(DurationValue(val), ctx)
				}
				return FormatInt64(val, ctx)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Int()
				if (attrs & FmtDivideBy1M) != 0 {
					val /= 1024 * 1024
				}
				return FormatInt64(val, ctx)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Uint()
				if (attrs & FmtDivideBy1M) != 0 {
					val /= 1024 * 1024
				}
				return FormatInt64(val, ctx)
			}
		case reflect.Float32, reflect.Float64:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).Float()
				if (attrs & FmtCeil) != 0 {
					val = math.Ceil(val)
				}
				return FormatFloat(val, ty.Kind() == reflect.Float32, ctx)
			}
		case reflect.String:
			return func(d any, ctx PrintMods) string {
				val := reflect.Indirect(reflect.ValueOf(d)).Field(ix).String()
				return FormatString(val, ctx)
			}
		default:
			panic(fmt.Sprintf("Unhandled type kind %d", ty.Kind()))
		}
	}
}
