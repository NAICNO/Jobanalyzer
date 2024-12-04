package table

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	umaps "go-utils/maps"
)

////////////////////////////////////////////////////////////////////////////////////////////////////
//
// Print modifiers.
//
// See comment block below.

type PrintMods = int

const (
	// These apply per-field according to modifiers
	PrintModSec = (1 << iota) // timestamps are printed as seconds since epoch
	PrintModIso               // timestamps are printed as Iso timestamps
	PrintModMax30

	// These are for the output format and are applied to all fields
	PrintModFixed      // fixed format
	PrintModJson       // JSON format
	PrintModCsv        // CSV format
	PrintModCsvNamed   // CSVNamed format
	PrintModAwk        // AWK format
	PrintModNoDefaults // Set (for any format) if option is set
)

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

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Formatting specs

type FormatOptions struct {
	Tag        string // if not ""
	Json       bool   // json explicitly requested
	Csv        bool   // csv or csvnamed explicitly requested
	Awk        bool   // awk explicitly requested
	Fixed      bool   // fixed output explicitly requested
	Named      bool   // csvnamed explicitly requested
	Header     bool   // true if nothing requested b/c fixed+header is default
	NoDefaults bool   // if true and the string returned is "*skip*" and the mode is csv or json then print nothing
}

func (fo *FormatOptions) IsDefaultFormat() bool {
	return !fo.Json && !fo.Csv && !fo.Awk && !fo.Fixed
}

// If fmtOpt is "" or "help" then the spec is the defaults, otherwise the spec is fmtOpt.
//
// Return a list of the known fields in the spec wrt the `formatters`, and a set of any other
// strings found in `spec`, plus also "help" if fmtOpt=="help".  Expand aliases (though not
// recursively: aliases must map to fundamental names).

type FieldSpec struct {
	Name   string
	Mod    PrintMods // /sec, /iso etc
	Header string    // name + modifier for backward compat
}

const (
	// Cheap way of preventing alias infinite expansion errors and all sorts of expansion DoS
	// attacks.  200 is probably fine: As of now, the max number of unique fields in the `jobs`
	// output is less than 50.  One could imagine various contorted but valid output specs needing
	// more than 200 by repeating fields in various ways but let's not worry.
	maxFields = 200
)

func ParseFormatSpec(
	defaults, fmtOpt string,
	formatters map[string]Formatter,
	aliases map[string][]string,
) (fields []FieldSpec, others map[string]bool, err error) {
	fields = make([]FieldSpec, 0)
	others = make(map[string]bool)
	spec := fmtOpt
	if fmtOpt == "" || fmtOpt == "help" {
		spec = defaults
	}
	if fmtOpt == "help" {
		others["help"] = true
	}
	if spec != "" {
		for _, fieldName := range strings.Split(spec, ",") {
			fields, _ = addField(fieldName, fields, others, formatters, aliases)
		}
	}
	return fields, others, nil
}

func addField(
	fieldName string,
	fields []FieldSpec,
	others map[string]bool,
	formatters map[string]Formatter,
	aliases map[string][]string,
) ([]FieldSpec, bool) {
	if newFields, found := recordField(fieldName, "", 0, fields, others, formatters, aliases); found {
		return newFields, true
	}
	if before, after, found := strings.Cut(fieldName, "/"); found {
		var m PrintMods
		switch after {
		case "m30":
			m = PrintModMax30
		case "sec":
			m = PrintModSec
		case "iso":
			m = PrintModIso
		}
		if m != 0 {
			if newFields, found := recordField(before, "/"+after, m, fields, others, formatters, aliases); found {
				return newFields, true
			}
		}
	}
	others[fieldName] = true
	return fields, false
}

func recordField(
	kwd, mod string,
	m PrintMods,
	fields []FieldSpec,
	others map[string]bool,
	formatters map[string]Formatter,
	aliases map[string][]string,
) ([]FieldSpec, bool) {
	anyAdded := false
	if _, found := formatters[kwd]; found {
		if len(fields) >= maxFields {
			// true ensures rapid unwinding
			return fields, true
		}
		fields = append(fields, FieldSpec{Name: kwd, Mod: m, Header: kwd + mod})
		anyAdded = true
	}
	if expansion, found := aliases[kwd]; found {
		for _, alias := range expansion {
			var added bool
			if len(fields) >= maxFields {
				// true ensures rapid unwinding
				return fields, true
			}
			fields, added = addField(alias, fields, others, formatters, aliases)
			anyAdded = anyAdded || added
		}
	}
	return fields, anyAdded
}

// Parse the non-field-name attributes as a set of formatting options.
//
// There are five format options, "fixed", "csv", "json", "awk", and default.  If none of the former
// four are requested and def is not DefaultNone then one of Fixed, Csv, Json, and Awk will be set
// according to the value of def, otherwise no format is set and the default interpretation is up to
// the formatter (see the parse command for an example of the latter).
//
// Header is set if the format (after defaulting) is "fixed" and no "noheader" attribute is present,
// or if the format is "csv" or "awk" and there is a "header" attribute.  It will never be set for
// "json".

type DefaultFormat int

const (
	DefaultNone DefaultFormat = iota
	DefaultFixed
	DefaultCsv
)

func StandardFormatOptions(others map[string]bool, def DefaultFormat) *FormatOptions {
	csvnamed := others["csvnamed"]
	csv := others["csv"] || csvnamed
	json := others["json"] && !csv
	awk := others["awk"] && !csv && !json
	fixed := others["fixed"] && !csv && !json && !awk
	nodefaults := others["nodefaults"]
	tag := ""
	for x := range others {
		if strings.HasPrefix(x, "tag:") {
			tag = x[4:]
			break
		}
	}
	if !csv && !json && !awk && !fixed {
		switch def {
		case DefaultFixed:
			fixed = true
		case DefaultCsv:
			csv = true
		case DefaultNone:
			break
		}
	}
	// json gets no header, even if one is requested
	header := (fixed && !others["noheader"]) || ((csv || awk) && others["header"])

	return &FormatOptions{
		Csv:        csv,
		Json:       json,
		Awk:        awk,
		Fixed:      fixed,
		Header:     header,
		Tag:        tag,
		Named:      csvnamed,
		NoDefaults: nodefaults,
	}
}

// FormatData defaults to fixed formatting.

func FormatData(
	out io.Writer,
	fields []FieldSpec,
	formatters map[string]Formatter,
	opts *FormatOptions,
	data []any,
) {
	ctx := ComputePrintMods(opts)

	// TODO: OPTIMIZEME: Instead of creating this huge matrix up-front and serializing field
	// formatting and output formatting, it might be better to set up some kind of
	// generator-formatter pipeline.  Allocation volume would be the same but we'd lower peak memory
	// and would take advantage of multiple cores. (Note writeStringPadded is not thread-safe, so be
	// careful.)

	// cols is a column-major representation of the output matrix, one column per field.
	cols := make([][]string, len(fields))
	for i := range fields {
		cols[i] = make([]string, len(data))
	}

	// Produce the formatted field values for all fields.
	type F struct {
		fmt func(any, PrintMods) string
		mod PrintMods
	}
	fmts := make([]F, len(fields))
	for c, f := range fields {
		fmts[c] = F{formatters[f.Name].Fmt, f.Mod}
	}
	for r, x := range data {
		for c := range fields {
			cols[c][r] = fmts[c].fmt(x, ctx|fmts[c].mod)
		}
	}

	if opts.Csv {
		formatCsv(out, fields, opts, cols)
	} else if opts.Json {
		formatJson(out, fields, opts, cols)
	} else if opts.Awk {
		formatAwk(out, fields, opts, cols)
	} else {
		formatFixed(out, fields, opts, cols)
	}
}

// The expectation here is that this is fairly low volume and that it's not worth it to try to
// optimize it to avoid allocations.
func formatFixed(unbufOut io.Writer, fields []FieldSpec, opts *FormatOptions, cols [][]string) {
	out := Buffered(unbufOut)
	defer out.Flush()

	// The column width is the max across all the entries in the column (including header,
	// if present).  If there's a tag, it is printed in the last column.
	numWidths := len(fields)
	tagCol := -1
	if opts.Tag != "" {
		tagCol = numWidths
		numWidths += 1
	}
	widths := make([]int, numWidths)

	if opts.Header {
		for col := 0; col < len(fields); col++ {
			widths[col] = max(widths[col], utf8.RuneCountInString(fields[col].Header))
		}
		if tagCol >= 0 {
			widths[tagCol] = max(widths[tagCol], len("tag"))
		}
	}

	for row := 0; row < len(cols[0]); row++ {
		for col := 0; col < len(fields); col++ {
			widths[col] = max(widths[col], utf8.RuneCountInString(cols[col][row]))
		}
		if tagCol >= 0 {
			widths[tagCol] = max(widths[tagCol], utf8.RuneCountInString(opts.Tag))
		}
	}

	var s strings.Builder

	// Header
	if opts.Header {
		s.Reset()
		for col := 0; col < len(fields); col++ {
			writeStringPadded(&s, widths[col], fields[col].Header)
		}
		if tagCol >= 0 {
			writeStringPadded(&s, widths[tagCol], "tag")
		}
		fmt.Fprintln(out, strings.TrimRight(s.String(), " "))
	}

	// Body
	for row := 0; row < len(cols[0]); row++ {
		s.Reset()
		for col := 0; col < len(fields); col++ {
			writeStringPadded(&s, widths[col], cols[col][row])
		}
		if tagCol >= 0 {
			writeStringPadded(&s, widths[tagCol], opts.Tag)
		}
		fmt.Fprintln(out, strings.TrimRight(s.String(), " "))
	}
}

// This padder is much faster than the equivalent Sprint(), and allocates almost nothing at all.
//
// We will almost never need more spaces than initial_spaces; the padder will create more as
// necessary but not update the global string b/c that would require a lock.
const initial_spaces = "                                                                                "

func writeStringPadded(s *strings.Builder, width int, str string) {
	spaces := initial_spaces
	needed := width - utf8.RuneCountInString(str) + 2
	for len(spaces) < needed {
		spaces = spaces + spaces
	}
	s.WriteString(str)
	s.WriteString(spaces[:needed])
}

func formatCsv(out io.Writer, fields []FieldSpec, opts *FormatOptions, cols [][]string) {
	w := csv.NewWriter(out)
	defer w.Flush()

	numFields := len(fields)
	if opts.Tag != "" {
		numFields++
	}
	outFields := make([]string, numFields)

	if opts.Header {
		for i := 0; i < len(fields); i++ {
			outFields[i] = fields[i].Header
		}
		if opts.Tag != "" {
			outFields[numFields-1] = opts.Tag
		}
		w.Write(outFields)
	}

	for row := 0; row < len(cols[0]); row++ {
		outIx := 0
		for col := 0; col < len(fields); col++ {
			val := cols[col][row]
			if opts.NoDefaults && val == "*skip*" {
				// For csvnamed, we have field tags and skipped fields are not represented at all.
				// For csv, we must print empty fields, because all rows should be the same length.
				if !opts.Named {
					outFields[outIx] = ""
					outIx++
				}
			} else if opts.Named {
				outFields[outIx] = fields[col].Header + "=" + val
				outIx++
			} else {
				outFields[outIx] = val
				outIx++
			}
		}
		if opts.Tag != "" {
			if opts.Named {
				outFields[outIx] = "tag=" + opts.Tag
				outIx++
			} else {
				outFields[outIx] = opts.Tag
				outIx++
			}
		}
		w.Write(outFields[:outIx])
	}
}

// There's no natural fit for the JSON encoder here, so just do it manually.
func formatJson(unbufOut io.Writer, fields []FieldSpec, opts *FormatOptions, cols [][]string) {
	out := Buffered(unbufOut)
	defer out.Flush()

	quotedFields := make([]string, len(fields))
	for i := range fields {
		quotedFields[i] = "\"" + QuoteJson(fields[i].Header) + "\""
	}
	quotedTag := ""
	if opts.Tag != "" {
		quotedTag = "\"tag\":\"" + QuoteJson(opts.Tag) + "\""
	}

	fmt.Fprint(out, "[")
	rowSep := ""
	var s strings.Builder
	for row := range cols[0] {
		s.Reset()
		s.WriteString(rowSep)
		s.WriteRune('{')
		fieldSep := ""
		for col := range quotedFields {
			val := cols[col][row]
			if opts.NoDefaults && val == "*skip*" {
				continue
			}
			s.WriteString(fieldSep)
			s.WriteString(quotedFields[col])
			s.WriteString(":\"")
			s.WriteString(QuoteJson(val))
			s.WriteRune('"')
			fieldSep = ","
		}
		if quotedTag != "" {
			s.WriteString(fieldSep)
			s.WriteString(quotedTag)
		}
		s.WriteRune('}')
		fmt.Fprint(out, s.String())
		rowSep = ","
	}
	fmt.Fprint(out, "]")
}

// TODO: IMPROVEME: Maybe handle control characters and other gunk better?
func QuoteJson(s string) string {
	found := false
	for _, r := range s {
		if r < ' ' || r == '"' {
			found = true
			break
		}
	}
	if !found {
		return s
	}
	t := ""
	for _, r := range s {
		if r < ' ' {
			r = ' '
		} else if r == '"' {
			t += "\\"
		}
		t += string(r)
	}
	return t
}

// awk output: fields are space-separated and spaces are not allowed within fields, they
// are replaced by `_`.  For good perf we count on ReplaceAll returning the input string if
// there are no replacements (current Go libraries do this correctly).
func formatAwk(unbufOut io.Writer, fields []FieldSpec, opts *FormatOptions, cols [][]string) {
	out := Buffered(unbufOut)
	defer out.Flush()

	var line strings.Builder
	for row := range cols[0] {
		line.Reset()
		sep := ""
		for col := range fields {
			val := cols[col][row]
			line.WriteString(sep)
			line.WriteString(strings.ReplaceAll(val, " ", "_"))
			sep = " "
		}
		if opts.Tag != "" {
			line.WriteString(sep)
			line.WriteString(opts.Tag)
			sep = " "
		}
		fmt.Fprintln(out, line.String())
	}
}

func FormatRawRowmajorAwk(unbufOut io.Writer, header []string, matrix [][]string) {
	out := Buffered(unbufOut)
	defer out.Flush()

	var line strings.Builder
	if header != nil {
		line.Reset()
		sep := ""
		for _, val := range header {
			line.WriteString(sep)
			line.WriteString(strings.ReplaceAll(val, " ", "_"))
			sep = " "
		}
		fmt.Fprintln(out, line.String())
	}

	for _, r := range matrix {
		line.Reset()
		sep := ""
		for _, val := range r {
			line.WriteString(sep)
			if val == "" {
				val = "."
			}
			line.WriteString(strings.ReplaceAll(val, " ", "_"))
			sep = " "
		}
		fmt.Fprintln(out, line.String())
	}
}

func FormatRawRowmajorCsv(out io.Writer, header []string, matrix [][]string) {
	w := csv.NewWriter(out)
	defer w.Flush()

	if header != nil {
		w.Write(header)
	}
	for _, r := range matrix {
		w.Write(r)
	}
}

func Buffered(unbufOut io.Writer) *bufio.Writer {
	if b, ok := unbufOut.(*bufio.Writer); ok {
		return b
	}
	return bufio.NewWriter(unbufOut)
}

type FormatHelp struct {
	Text     string
	Fields   []string
	Helps    map[string]string
	Aliases  map[string][]string
	Defaults string
}

// AliasOf is nonempty for a field with its own formatter that is an alias of a canonical field and
// whose name is used as the field name, not the canonical field name.  The value is the name of the
// canonical field.

type Formatter struct {
	Fmt     func(data any, ctx PrintMods) string
	Help    string
	AliasOf string
}

func DefAlias(formatters map[string]Formatter, canonical, alias string) {
	f, found := formatters[canonical]
	if !found {
		panic(fmt.Sprintf("Formatter not found: %s", canonical))
	}
	f.AliasOf = canonical
	formatters[alias] = f
}

func StandardFormatHelp(
	fmt string,
	helpText string,
	formatters map[string]Formatter,
	aliases map[string][]string,
	defaultFields string,
) *FormatHelp {
	if fmt == "help" {
		fields := make([]string, 0, len(formatters))
		helps := make(map[string]string, len(formatters))
		newAliases := maps.Clone(aliases)
		for k, v := range formatters {
			if v.AliasOf != "" {
				newAliases[k] = []string{v.AliasOf}
			} else {
				fields = append(fields, k)
				helps[k] = v.Help
			}
		}
		return &FormatHelp{
			Text:     helpText,
			Fields:   fields,
			Helps:    helps,
			Aliases:  newAliases,
			Defaults: defaultFields,
		}
	}
	return nil
}

func PrintFormatHelp(out io.Writer, h *FormatHelp) {
	if h != nil {
		fmt.Fprintln(out, h.Text)
		fmt.Fprintln(out, "Syntax:\n  -fmt=(field|alias|control),...")
		fmt.Fprintln(out, "\nFields:")
		fields := slices.Clone(h.Fields)
		sort.Sort(sort.StringSlice(fields))
		for _, f := range fields {
			fmt.Fprintf(out, "  %s - %s\n", f, h.Helps[f])
		}
		if len(h.Aliases) > 0 {
			fmt.Fprintln(out, "\nAliases:")
			aliases := umaps.Keys(h.Aliases)
			sort.Sort(sort.StringSlice(aliases))
			for _, k := range aliases {
				// Do not sort the names in the expansion because the order matters
				fmt.Fprintf(out, "  %s --> %s\n", k, strings.Join(h.Aliases[k], ","))
			}
		}
		fmt.Fprintf(
			out, `
Defaults:
  %s

For control operations and data field modifiers, try 'sonalyze help printing'.

`,
			h.Defaults,
		)
	}
}
