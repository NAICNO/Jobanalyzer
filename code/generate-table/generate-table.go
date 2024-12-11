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

	"generate-table/parser"
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
// Types we know about and information about them.
//
// The formatter for type Ty is usually FormatTy(Ty, PrintMods) -> string, but the name can be
// overridden by registering the type.
//
// The user-facing type name for type Ty is usually Ty but it can be overridden. BEWARE that this
// override has implications for the applicable query operators: if a type is "string" then string
// operators apply; if a type is "GpuSet" then some kind of set operators apply (TBD).

type typeInfo struct {
	formatter string
	helpName  string
}

var knownTypes = map[string]typeInfo{
	"int":                  typeInfo{formatter: "FormatInt"},
	"float":                typeInfo{formatter: "FormatFloat32"},
	"double":               typeInfo{formatter: "FormatFloat64"},
	"string":               typeInfo{formatter: "FormatString"},
	"bool":                 typeInfo{formatter: "FormatBool"},
	"[]string":             typeInfo{formatter: "FormatStrings", helpName: "string list"},
	"IntCeil":              typeInfo{helpName: "int"},
	"IntDiv1M":             typeInfo{helpName: "int"},
	"IntOrEmpty":           typeInfo{helpName: "int"},
	"DateTimeValueOrBlank": typeInfo{helpName: "DateTimeValue"},
	"IsoDateTimeOrUnknown": typeInfo{helpName: "IsoDateTimeValue"},
	"Ustr":                 typeInfo{helpName: "string"},
	"UstrMax30":            typeInfo{helpName: "string"},
}

func formatName(ty string) string {
	if probe := knownTypes[ty]; probe.formatter != "" {
		return probe.formatter
	}
	return "Format" + ty
}

func userFacingTypeName(ty string) string {
	if probe := knownTypes[ty]; probe.helpName != "" {
		return probe.helpName
	}
	return ty
}

func processBlock(block *parser.TableBlock) {
	fmt.Fprintf(output, "// DO NOT EDIT.  Generated from %s by generate-table\n\n", inputName)
	for _, p := range block.Prefix {
		fmt.Fprintf(output, "%s\n", p.Text)
	}
	fieldList := fieldSection(block.TableName, block.Fields)
	if block.Generate != nil {
		generateSection(*block.Generate, fieldList)
	}
	if block.Help != nil {
		helpSection(block.TableName, block.Help)
	}
	if len(block.Aliases) != 0 {
		aliasesSection(block.TableName, block.Aliases)
	}
	if block.Defaults != nil {
		defaultsSection(block.TableName, *block.Defaults)
	}
}

// Arguable whether we should be checking for valid attribute names here or in the parser, I'm
// thinking it's better to do it here.

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
			fmt.Fprintf(output, "\t\tHelp: \"(%s) %s\",\n", userFacingTypeName(displayType), d)
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

func defaultsSection(tableName, defaultName string) {
	fmt.Fprintf(output, "const %sDefaultFields = \"%s\"\n\n", tableName, defaultName)
}

var validAttr = map[string]bool{
	"desc":     true,
	"alias":    true,
	"field":    true,
	"indirect": true,
	"config":   true,
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
