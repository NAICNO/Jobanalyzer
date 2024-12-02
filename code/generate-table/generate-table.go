// Usage: generate-table -o output-name input-name
//
// Output-name can be `-` for stdout but defaults to `table.out.go`; input-name can be `-` for
// stdin.
//
// See README.md for all documentation.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	inputName string
	input     *LineStream
	output    io.Writer

	outputName = flag.String("o", "table.out.go", "Name of outputFile, - for stdout")
)

func main() {
	var err error
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] inputFile\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 1 {
		inputName = flag.Args()[0]
	} else {
		flag.Usage()
		os.Exit(2)
	}

	var inputFile io.Reader
	if inputName == "-" {
		inputFile = os.Stdin
	} else {
		inputFile, err = os.Open(inputName)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *outputName == "-" {
		output = os.Stdout
	} else {
		theFile, err := os.Create(*outputName)
		if err != nil {
			log.Fatal(err)
		}
		defer theFile.Close()
		output = theFile
	}

	input = NewLineStream(inputName, inputFile)
	for block := scanBlock(); block != nil; block = scanBlock() {
		processBlock(block)
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Types we know about and information about them.  The formatter for type Ty is usually
// FormatTy(Ty, PrintMods) -> string, but the name can be overridden (for now) by registering the
// type.  Probably more attributes will appear here.

type typeInfo struct {
	formatter string
}

var knownTypes = map[string]typeInfo{
	"int":      typeInfo{formatter: "FormatInt"},
	"float":    typeInfo{formatter: "FormatFloat32"},
	"double":   typeInfo{formatter: "FormatFloat64"},
	"string":   typeInfo{formatter: "FormatString"},
	"bool":     typeInfo{formatter: "FormatBool"},
	"[]string": typeInfo{formatter: "FormatStrings"},
}

func formatName(ty string) string {
	if probe, found := knownTypes[ty]; found {
		return probe.formatter
	}
	return "Format" + ty
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Table structure.

type tableBlock struct {
	tableName string         // From the /*TABLE <table-name> line
	prefix    []Line         // Every line before %%, in order
	sections  []tableSection // Parsed sections after %%, in order
}

type tableSection struct {
	header Line   // header line, unedited
	body   []Line // every line after the header, unedited
}

var (
	fieldsLine   = regexp.MustCompile(`^FIELDS\s+(\S+)\s*$`)
	generateLine = regexp.MustCompile(`^GENERATE\s+(\S+)\s*$`)
	helpLine     = regexp.MustCompile(`^HELP(?:\s+(\S+))?\s*$`)
	aliasesLine  = regexp.MustCompile(`^ALIASES\s*$`)
	defaultsLine = regexp.MustCompile(`^DEFAULTS\s+(\S+)\s*$`)
)

type fieldSpec struct {
	name, ty string
}

func processBlock(block *tableBlock) {
	fmt.Fprintf(output, "// DO NOT EDIT.  Generated from %s by generate-table\n\n", inputName)

	for _, p := range block.prefix {
		fmt.Fprintf(output, "%s\n", p.Text)
	}

	var fieldList []fieldSpec

	si := 0
	if sect, attr, nextSi := trySect(block.sections, si, fieldsLine); sect != nil {
		si = nextSi
		if attr == "" {
			sect.header.Bad("Bogus type name for FIELDS")
		}
		fieldList = fieldSection(block, sect, attr)
	} else {
		log.Fatal("Missing required FIELDS section")
	}
	if sect, attr, nextSi := trySect(block.sections, si, generateLine); sect != nil {
		si = nextSi
		if attr == "" {
			sect.header.Bad("Bogus type name for GENERATE")
		}
		generateSection(block, fieldList, attr)
	}
	if sect, attr, nextSi := trySect(block.sections, si, helpLine); sect != nil {
		si = nextSi
		helpSection(block, sect, attr)
	}
	if sect, _, nextSi := trySect(block.sections, si, aliasesLine); sect != nil {
		si = nextSi
		aliasesSection(block, sect)
	}
	if sect, attr, nextSi := trySect(block.sections, si, defaultsLine); sect != nil {
		si = nextSi
		if attr == "" {
			sect.header.Bad("Bogus alias name for DEFAULTS")
		}
		defaultsSection(block, sect, attr)
	}

	if si < len(block.sections) {
		block.sections[si].header.Bad("Unknown or repeated section")
	}
}

func trySect(sections []tableSection, si int, re *regexp.Regexp) (*tableSection, string, int) {
	if si < len(sections) {
		sect := &sections[si]
		if m := re.FindStringSubmatch(sect.header.Text); m != nil {
			attr := ""
			if len(m) > 1 {
				attr = m[1]
			}
			return sect, attr, si + 1
		}
	}
	return nil, "", si
}

func fieldSection(block *tableBlock, sect *tableSection, recordType string) (fieldList []fieldSpec) {
	fieldList = make([]fieldSpec, 0)
	type aliasDef struct {
		field, alias string
	}
	aliasDefs := make([]aliasDef, 0)

	fmt.Fprintf(output, "// MT: Constant after initialization; immutable\n")
	fmt.Fprintf(output, "var %sFormatters = map[string]Formatter[%s]{\n", block.tableName, recordType)
	for _, bodyLine := range sect.body {
		if bodyLine.IsBlankOrComment() {
			continue
		}
		fieldName, displayType, attrs := parseBodyLine(bodyLine)
		actualFieldName := fieldName
		if fn, found := attrs["field"]; found {
			actualFieldName = fn
		} else {
			fieldList = append(fieldList, fieldSpec{fieldName, displayType})
		}
		var needsConfig bool
		if config := attrs["config"]; config != "" {
			switch config {
			case "true":
				needsConfig = true
			case "false":
				// nothing
			default:
				bodyLine.Bad("in boolean attribute " + config)
			}
		} else if strings.Index(actualFieldName, "Relative") != -1 {
			needsConfig = true
		}
		fmt.Fprintf(output, "\t\"%s\": {\n", fieldName)
		fmt.Fprintf(output, "\t\tFmt: func(d %s, ctx PrintMods) string {\n", recordType)
		if ptrName := attrs["indirect"]; ptrName != "" {
			fmt.Fprintf(output, "\t\t\tif d.%s != nil {\n", ptrName)
			fmt.Fprintf(output, "\t\t\t\treturn %s(%s(d.%s.%s), ctx)\n",
				formatName(displayType), displayType, ptrName, actualFieldName)
			fmt.Fprintf(output, "\t\t\t}\n")
			fmt.Fprintf(output, "\t\t\treturn \"?\"\n")
		} else {
			fmt.Fprintf(output, "\t\t\treturn %s(%s(d.%s), ctx)\n",
				formatName(displayType), displayType, actualFieldName)
		}
		fmt.Fprintf(output, "\t\t},\n")
		if d := attrs["desc"]; d != "" {
			fmt.Fprintf(output, "\t\tHelp: \"%s\",\n", d)
		}
		if needsConfig {
			fmt.Fprintf(output, "\t\tNeedsConfig: true,\n")
		}
		fmt.Fprintf(output, "\t},\n")
		if aliases := attrs["alias"]; aliases != "" {
			names := strings.Split(aliases, ",")
			for _, n := range names {
				aliasDefs = append(aliasDefs, aliasDef{fieldName, n})
			}
		}
	}
	fmt.Fprintf(output, "}\n\n")

	if len(aliasDefs) > 0 {
		fmt.Fprintf(output, "func init() {\n")
		for _, d := range aliasDefs {
			fmt.Fprintf(output, "\tDefAlias(%sFormatters, \"%s\", \"%s\")\n",
				block.tableName, d.field, d.alias)
		}
		fmt.Fprintf(output, "}\n\n")
	}
	return
}

func generateSection(block *tableBlock, fieldList []fieldSpec, recordName string) {
	fmt.Fprintf(output, "type %s struct {\n", recordName)
	for _, fs := range fieldList {
		fmt.Fprintf(output, "\t%s %s\n", fs.name, fs.ty)
	}
	fmt.Fprintf(output, "}\n\n")
}

func helpSection(block *tableBlock, sect *tableSection, commandName string) {
	first := 0
	last := len(sect.body) - 1
	for first < len(sect.body) && sect.body[first].IsBlankOrComment() {
		first++
	}
	for last > first && sect.body[last].IsBlankOrComment() {
		last--
	}
	fmt.Fprintf(output, "const %sHelp = `\n%s\n", block.tableName, block.tableName)
	for i := first; i <= last; i++ {
		fmt.Fprintf(output, "%s\n", sect.body[i].Text)
	}
	fmt.Fprintf(output, "`\n\n")
	if commandName != "" {
		fmt.Fprintf(output, "func (c *%s) MaybeFormatHelp() *FormatHelp {\n", commandName)
		fmt.Fprintf(
			output,
			"\treturn StandardFormatHelp(c.Fmt, %sHelp, %sFormatters, %sAliases, %sDefaultFields)\n",
			block.tableName, block.tableName, block.tableName, block.tableName)
		fmt.Fprintf(output, "}\n\n")
	}
}

func aliasesSection(block *tableBlock, sect *tableSection) {
	fmt.Fprintf(output, "// MT: Constant after initialization; immutable\n")
	fmt.Fprintf(output, "var %sAliases = map[string][]string{\n", block.tableName)
	for _, l := range sect.body {
		if l.IsBlankOrComment() {
			continue
		}
		aliasName, aliasExp := parseAliases(l)
		fmt.Fprintf(output, "\t\"%s\": []string{%s},\n",
			aliasName, "\""+strings.Join(aliasExp, "\",\"")+"\"")
	}
	fmt.Fprintf(output, "}\n\n")
}

func defaultsSection(block *tableBlock, sect *tableSection, defaultName string) {
	assertEmptyBody(sect)
	fmt.Fprintf(output, "const %sDefaultFields = \"%s\"\n\n", block.tableName, defaultName)
}

func assertEmptyBody(section *tableSection) {
	for _, l := range section.body {
		if !l.IsBlankOrComment() {
			l.Bad("section body")
		}
	}
}

var (
	tableContentLine = regexp.MustCompile(`^\s*(\S+)\s+(\S+)(?:\s+(.*))?`)
	tableFieldAttr   = regexp.MustCompile(`^\s*([a-z]+):"([^"]*)"`)
)

func parseBodyLine(l Line) (fieldName, displayType string, attrs map[string]string) {
	if m := tableContentLine.FindStringSubmatch(l.Text); m != nil {
		fieldName = m[1]
		displayType = m[2]
		attrs = parseAttrs(l, m[3])
		return
	}
	l.Bad("in FIELDS")
	return
}

var validAttr = map[string]bool{
	"desc":     true,
	"alias":    true,
	"field":    true,
	"indirect": true,
	"config":   true,
}

func parseAttrs(context Line, s string) map[string]string {
	attrs := make(map[string]string)
	for s != "" {
		m := tableFieldAttr.FindStringSubmatch(s)
		if m == nil || !validAttr[m[1]] {
			context.Bad("Invalid attribute staring at " + s)
		}
		attrs[m[1]] = m[2]
		s = s[len(m[0]):]
	}
	return attrs
}

var aliasLine = regexp.MustCompile(`^\s*(\S+)\s+(\S+)\s*$`)

func parseAliases(l Line) (aliasName string, aliasDef []string) {
	if m := aliasLine.FindStringSubmatch(l.Text); m != nil {
		aliasName = m[1]
		aliasDef = strings.Split(m[2], ",")
		return
	}
	l.Bad("in ALIASES")
	return
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Block stream, with coarse section parsing.

var (
	tablePrefix     = regexp.MustCompile(`^/\*TABLE\s+(\S+)\s*$`)
	tableSuffix     = regexp.MustCompile(`^ELBAT\*/`)
	prefixEndMarker = regexp.MustCompile(`^%%\s*$`)
	sectStart       = regexp.MustCompile(`^(?:FIELDS|GENERATE|HELP|ALIASES|DEFAULTS)`)
)

func scanBlock() *tableBlock {
	var tableName string
	for {
		if input.AtEof() {
			return nil
		}
		if m := input.Match(tablePrefix); m != nil {
			tableName = m[1]
			input.Next()
			break
		}
		input.Next()
	}

	prefix := make([]Line, 0)
	for {
		input.AssertNotEof("inside table prefix")
		if m := input.Match(tableSuffix); m != nil {
			input.MustGet().Bad("Missing %%%% inside table block")
		}
		if m := input.Match(prefixEndMarker); m != nil {
			input.Next()
			break
		}
		prefix = append(prefix, input.MustGet())
	}

	sections := make([]tableSection, 0)
	var header Line
	body := make([]Line, 0)
	for {
		input.AssertNotEof("inside table block")
		if m := input.Match(tableSuffix); m != nil {
			input.Next()
			break
		}
		if header.Text == "" {
			if probe, ok := input.Peek(); ok && probe.IsBlankOrComment() {
				// blank before the first section
				input.Next()
				continue
			}
		}
		if m := input.Match(sectStart); m != nil {
			if header.Text != "" {
				sections = append(sections, tableSection{header, body})
			}
			header = input.MustGet()
			body = make([]Line, 0)
			continue
		}
		if header.Text == "" {
			input.MustGet().Bad("Garbage before the first section")
		}
		body = append(body, input.MustGet())
	}
	if header.Text != "" {
		sections = append(sections, tableSection{header, body})
	}

	return &tableBlock{
		tableName,
		prefix,
		sections,
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Input line stream, with lookahead and some matching.  This eats I/O errors, currently.

type LineStream struct {
	scanner *bufio.Scanner
	valid   bool
	curr    Line
	eof     bool
	lineno  int
	file    string
}

type Line struct {
	Lineno int
	File   string
	Text   string
}

func (l Line) Bad(context string) {
	log.Fatalf("%s:%d: Illegal line (%s)\n%s", l.File, l.Lineno, context, l.Text)
}

var blankOrCommentLine = regexp.MustCompile(`^\s*(?:#.*)?$`)

func (l Line) IsBlankOrComment() bool {
	return blankOrCommentLine.MatchString(l.Text)
}

func NewLineStream(inputName string, inputStream io.Reader) *LineStream {
	return &LineStream{
		scanner: bufio.NewScanner(inputStream),
		file:    inputName,
	}
}

func (ls *LineStream) AssertNotEof(context string) {
	if ls.AtEof() {
		log.Fatalf("%s:%d: Unexpected EOF %s", ls.file, ls.lineno, context)
	}
}

func (ls *LineStream) AtEof() bool {
	if !ls.valid {
		ls.fill()
	}
	return ls.eof
}

func (ls *LineStream) Match(re *regexp.Regexp) []string {
	if l, ok := ls.Peek(); ok {
		return re.FindStringSubmatch(l.Text)
	}
	return nil
}

func (ls *LineStream) Peek() (Line, bool) {
	if !ls.valid {
		ls.fill()
	}
	return ls.curr, ls.valid
}

func (ls *LineStream) Get() (l Line, found bool) {
	if !ls.valid {
		ls.fill()
	}
	l, found = ls.curr, ls.valid
	ls.valid = false
	return
}

func (ls *LineStream) MustGet() (l Line) {
	if !ls.valid {
		ls.fill()
	}
	if !ls.valid {
		panic("No input line in MustGet()")
	}
	l = ls.curr
	ls.valid = false
	return
}

func (ls *LineStream) Next() {
	if !ls.valid {
		ls.fill()
	}
	ls.valid = false
}

func (ls *LineStream) fill() {
	if ls.valid {
		panic("Overwriting line in fill()")
	}
	if ls.scanner.Scan() {
		ls.lineno++
		lineno := ls.lineno
		text := ls.scanner.Text()
		for strings.HasSuffix(text, "\\") {
			text = strings.TrimSuffix(text, "\\")
			if !ls.scanner.Scan() {
				log.Fatalf("%s:%d: Line continued by backslash but no next line", ls.file, ls.lineno)
			}
			ls.lineno++
			text += strings.TrimLeft(ls.scanner.Text(), " \t")
		}
		ls.valid = true
		ls.curr = Line{Lineno: lineno, File: ls.file, Text: text}
		return
	}
	ls.eof = true
	return
}
