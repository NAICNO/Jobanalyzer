// Usage: generate-table -o output-name input-name
//
// Output-name can be `-` for stdout but defaults to `table.out.go`; input-name can be `-` for
// stdin.
//
// See README.md for all documentation.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"generate-table/parser"
)

var (
	inputName  string
	output     io.Writer
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

	p := parser.NewParser(inputName, inputFile)
	for {
		block, err := p.Parse()
		if err != nil {
			log.Fatal(err)
		}
		if block == nil {
			break
		}
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
	helpName  string // default is the name as given
	formatter string // default is Format<Typename>
}

var knownTypes = map[string]typeInfo{
	"[]string": typeInfo{
		helpName:  "string list",
		formatter: "FormatStrings",
	},
	"F64Ceil": typeInfo{
		helpName: "int",
	},
	"U64Div1M": typeInfo{
		helpName: "int",
	},
	"IntOrEmpty": typeInfo{
		helpName: "int",
	},
	"DateTimeValueOrBlank": typeInfo{
		helpName: "DateTimeValue",
	},
	"IsoDateTimeOrUnknown": typeInfo{helpName: "IsoDateTimeValue"},
	"Ustr":                 typeInfo{helpName: "string"},
	"UstrMax30":            typeInfo{helpName: "string"},
	"gpuset.GpuSet": typeInfo{
		helpName:  "GpuSet",
		formatter: "FormatGpuSet",
	},
	"*Hostnames": typeInfo{
		helpName:  "Hostnames",
		formatter: "FormatHostnames",
	},
}

func formatName(ty string) string {
	if probe := knownTypes[ty]; probe.formatter != "" {
		return probe.formatter
	}
	return "Format" + capitalize(ty)
}

func userFacingTypeName(ty string) string {
	if probe := knownTypes[ty]; probe.helpName != "" {
		return probe.helpName
	}
	// TODO: Strip suffix size information
	return ty
}

// We know we're dealing with ASCII so this is good enough
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

type fieldSpec struct {
	name, ty string
}

func processBlock(block *parser.TableBlock) {
	fmt.Fprintf(output, "// DO NOT EDIT.  Generated from %s by generate-table\n\n", inputName)
	for _, p := range block.Prefix {
		fmt.Fprintf(output, "%s\n", p)
	}
	fmt.Fprintf(output, `
import (
	"cmp"
	"fmt"
	"io"
	"go-utils/gpuset"
	. "sonalyze/common"
	. "sonalyze/table"
)
var (
	_ = cmp.Compare(0,0)
	_ fmt.Formatter
	_ = io.SeekStart
	_ = UstrEmpty
	_ gpuset.GpuSet
)
`)
	fieldList := fieldSection(block.TableName, &block.Fields)
	if block.Generate != "" {
		generateSection(block.Generate, fieldList)
	}
	if block.Summary != nil {
		summarySection(block.TableName, block.Summary)
	}
	if block.Help != nil {
		helpSection(block.TableName, block.Help)
	}
	if len(block.Aliases) > 0 {
		aliasesSection(block.TableName, block.Aliases)
	}
	if len(block.Defaults) > 0 {
		defaultsSection(block.TableName, block.Defaults)
	}
}

// Arguable whether we should be checking for valid attribute names here or in the parser, I'm
// thinking it's better to do it here.

func fieldSection(tableName string, fields *parser.FieldSect) (fieldList []fieldSpec) {
	fieldList = fieldFormatters(tableName, fields)
	return
}

func fieldFormatters(tableName string, fields *parser.FieldSect) (fieldList []fieldSpec) {
	fieldList = make([]fieldSpec, 0)

	type aliasDef struct {
		field, alias string
	}
	aliasDefs := make([]aliasDef, 0)

	fmt.Fprintf(output, "// MT: Constant after initialization; immutable\n")
	fmt.Fprintf(output, "var %sFormatters = map[string]Formatter[%s]{\n", tableName, fields.Type)
	for _, field := range fields.Fields {
		attrs := make(map[string]string)
		for _, attr := range field.Attrs {
			if !validAttr[attr.Name] {
				log.Fatalf("Field %s: Invalid attribute name %s", field.Name, attr.Name)
			}
			attrs[attr.Name] = attr.Value
		}

		actualFieldName := field.Name
		if fn, found := attrs["field"]; found {
			actualFieldName = fn
		} else {
			fieldList = append(fieldList, fieldSpec{field.Name, field.Type})
		}

		var needsConfig bool
		if config, found := attrs["config"]; found {
			switch config {
			case "true":
				needsConfig = true
			case "false":
				// nothing
			default:
				log.Fatalf("Field %s: Bad attribute value for 'config'", field.Name)
			}
		} else if strings.Index(actualFieldName, "Relative") != -1 {
			needsConfig = true
		}

		fmt.Fprintf(output, "\t\"%s\": {\n", field.Name)
		fmt.Fprintf(output, "\t\tFmt: func(d %s, ctx PrintMods) string {\n", fields.Type)
		formatter := formatName(field.Type)
		if ptrName := attrs["indirect"]; ptrName != "" {
			fmt.Fprintf(output, "\t\t\tif (d.%s) != nil {\n", ptrName)
			fmt.Fprintf(
				output, "\t\t\t\treturn %s((d.%s.%s), ctx)\n", formatter, ptrName, actualFieldName)
			fmt.Fprintf(output, "\t\t\t}\n")
			fmt.Fprintf(output, "\t\t\treturn \"?\"\n")
		} else {
			fmt.Fprintf(output, "\t\t\treturn %s((d.%s), ctx)\n", formatter, actualFieldName)
		}
		fmt.Fprintf(output, "\t\t},\n")
		if d := attrs["desc"]; d != "" {
			fmt.Fprintf(output, "\t\tHelp: \"(%s) %s\",\n", userFacingTypeName(field.Type), d)
		}
		if needsConfig {
			fmt.Fprintf(output, "\t\tNeedsConfig: true,\n")
		}
		fmt.Fprintf(output, "\t},\n")
		if aliases := attrs["alias"]; aliases != "" {
			names := strings.Split(aliases, ",")
			for _, n := range names {
				aliasDefs = append(aliasDefs, aliasDef{field.Name, n})
			}
		}
	}
	fmt.Fprintf(output, "}\n\n")

	if len(aliasDefs) > 0 {
		fmt.Fprintf(output, "func init() {\n")
		for _, d := range aliasDefs {
			fmt.Fprintf(output, "\tDefAlias(%sFormatters, \"%s\", \"%s\")\n",
				tableName, d.field, d.alias)
		}
		fmt.Fprintf(output, "}\n\n")
	}
	return
}

var validAttr = map[string]bool{
	"desc":     true,
	"alias":    true,
	"field":    true,
	"indirect": true,
	"config":   true,
}

func generateSection(recordName string, fieldList []fieldSpec) {
	fmt.Fprintf(output, "type %s struct {\n", recordName)
	for _, fs := range fieldList {
		fmt.Fprintf(output, "\t%s %s\n", fs.name, fs.ty)
	}
	fmt.Fprintf(output, "}\n\n")
}

func helpSection(tableName string, help *parser.HelpSect) {
	fmt.Fprintf(output, "const %sHelp = `\n%s\n", tableName, tableName)
	for _, l := range stripBlanks(help.Text) {
		fmt.Fprintf(output, "%s\n", l)
	}
	fmt.Fprintf(output, "`\n\n")
	if help.Command != "" {
		fmt.Fprintf(output, "func (c *%s) MaybeFormatHelp() *FormatHelp {\n", help.Command)
		fmt.Fprintf(
			output,
			"\treturn StandardFormatHelp(c.Fmt, %sHelp, %sFormatters, %sAliases, %sDefaultFields)\n",
			tableName, tableName, tableName, tableName)
		fmt.Fprintf(output, "}\n\n")
	}
}

func summarySection(tableName string, summary *parser.HelpSect) {
	fmt.Fprintf(output, "func (c *%s) Summary(out io.Writer) {\n", summary.Command)
	fmt.Fprintf(output, "\tfmt.Fprint(out, `")
	for _, l := range stripBlanks(summary.Text) {
		fmt.Fprintf(output, "%s\n", l)
	}
	fmt.Fprintf(output, "`)\n")
	fmt.Fprintf(output, "}\n\n")
}

func stripBlanks(ls []string) []string {
	first := 0
	last := len(ls) - 1
	for first < len(ls) && isBlank(ls[first]) {
		first++
	}
	for last > first && isBlank(ls[last]) {
		last--
	}
	return ls[first : last+1]
}

func aliasesSection(tableName string, aliases []parser.Alias) {
	fmt.Fprintf(output, "// MT: Constant after initialization; immutable\n")
	fmt.Fprintf(output, "var %sAliases = map[string][]string{\n", tableName)
	for _, alias := range aliases {
		fmt.Fprintf(output, "\t\"%s\": []string{%s},\n",
			alias.Name, "\""+strings.Join(alias.Fields, "\",\"")+"\"")
	}
	fmt.Fprintf(output, "}\n\n")
}

func defaultsSection(tableName string, defaultNames []string) {
	fmt.Fprintf(
		output, "const %sDefaultFields = \"%s\"\n\n", tableName, strings.Join(defaultNames, ","))
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}
