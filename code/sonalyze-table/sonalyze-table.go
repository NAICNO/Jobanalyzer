// Usage: sonalyze-table -o output-name input-name
//
// This will scan input-name for one or more table definition blocks.  Optionally there is a package
// prefix, which must come before any blocks.

// The idea here is that the directive //TABLE <name> <record-type> introduces the <name> of the
// table for row type <record-type> and everything in that sequence of comments is either a blank
// line or defines a field in the printable representation.  The <name> should be the name of the
// package as well.
//
// Each line starts with the name and type of the data field being exposed followed by attributes.
// The type is the *formatting type*.  The data field may have an underlying type that is different
// but which must be convertible to the formatting type with a cast.
//
// Attributes can be:
//   desc  - description for -fmt help
//   alias - comma-separated aliases
//   attr  - formatting attributes of field, if necessary
//
// For this to work, you must first build *and install* the sonalyze-table tool, something
// we could do with `make setup` or something.  Anyway, it's just a detail - clients need not
// do it, nor the build bot.
//
// The description will (re)generate the file node-table.go in the current directory, which will
// contain definitions of formatters and query predictates.  THESE CAN EVENTUALLY BE STRONGLY TYPED,
// no more of the `any`.
//
// Formatters will be called <name>Formatters.  Query operators tbd.
//
// Eventually I see more being included in this comment, eg the summary help, defaults, and various
// aliases.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type packageBlock struct {
	packageName string
	imports     []string
}

type tableBlock struct {
	baseName   string
	recordType string
	lines      []blockLine
}

type blockLine struct {
	fieldName, displayType string
	attrs                  map[string]string
}

var (
	packageLine      = regexp.MustCompile(`^//PACKAGE\s+(\S+)`)
	importLine       = regexp.MustCompile(`^//IMPORT\s+(\S+)`)
	tableStartLine   = regexp.MustCompile(`^//TABLE\s+(\S+)\s+(\S+)`)
	tableContentLine = regexp.MustCompile(`^//\s+(\S+)\s+(\S+)(?:\s+(.*))?`)
	tableFieldAttr   = regexp.MustCompile(`^\s*([a-z]+):"([^"]*)"`)
)

var (
	outputName = flag.String("o", "table-generated.go", "Name of outputFile")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] inputFile\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	inputName := flag.Args()[0]
	inputFile, err := os.Open(inputName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(inputFile)
	out, err := os.Create(*outputName)
	if err != nil {
		log.Fatal(err)
	}
	havePackage := false
	haveTable := false
	for x := scanBlock(scanner); x != nil; x = scanBlock(scanner) {
		switch b := x.(type) {
		case *packageBlock:
			if havePackage {
				log.Fatal("Already seen a package prefix")
			}
			if haveTable {
				log.Fatal("Too late for a package prefix: already seen a table")
			}
			havePackage = true
			fmt.Fprintf(out, "// DO NOT EDIT.  Generated from %s by sonalyze-table\n\n", inputName)
			fmt.Fprintf(out, "package %s\n", b.packageName)
			fmt.Fprintf(out, "import (\n")
			for _, i := range b.imports {
				fmt.Fprintf(out, "    %s\n", i)
			}
			fmt.Fprintf(out, "    . \"sonalyze/table\"\n")
			fmt.Fprintf(out, ")\n\n")

		case *tableBlock:
			haveTable = true
			if !havePackage {
				fmt.Fprintf(out, "// DO NOT EDIT.  Generated from %s by sonalyze-table\n\n", inputName)
			}
			fmt.Fprintf(out, "// MT: Constant after initialization; immutable\n")
			fmt.Fprintf(out, "var %sFormatters = map[string]Formatter{\n", b.baseName)
			type aliasDef struct {
				field, alias string
			}
			aliasDefs := make([]aliasDef, 0)
			for _, l := range b.lines {
				var formatFunc string
				switch l.displayType {
				case "string":
					formatFunc = "FormatString"
				case "int":
					formatFunc = "FormatInt"
				case "bool":
					formatFunc = "FormatBool"
				default:
					panic("NYI")
				}
				fmt.Fprintf(out, "    \"%s\": {\n", l.fieldName)
				fmt.Fprintf(out, "        Fmt: func(d any, ctx PrintMods) string {\n")
				fmt.Fprintf(out, "            return %s(d.(*%s).%s, ctx)\n",
					formatFunc, b.recordType, l.fieldName)
				fmt.Fprintf(out, "        },\n")
				if d := l.attrs["desc"]; d != "" {
					fmt.Fprintf(out, "        Help: \"%s\",\n", d)
				}
				fmt.Fprintf(out, "    },\n")
				if aliases := l.attrs["alias"]; aliases != "" {
					names := strings.Split(aliases, ",")
					for _, n := range names {
						aliasDefs = append(aliasDefs, aliasDef{l.fieldName, n})
					}
				}
			}
			fmt.Fprintf(out, "}\n\n")
			if len(aliasDefs) > 0 {
				fmt.Fprintf(out, "func init() {\n")
				for _, d := range aliasDefs {
					fmt.Fprintf(out, "    DefAlias(%sFormatters, \"%s\", \"%s\")\n",
						b.baseName, d.field, d.alias)
				}
				fmt.Fprintf(out, "}\n")
			}

		default:
			panic("NYI")
		}
	}
}

func scanBlock(scanner *bufio.Scanner) any {
	for scanner.Scan() {
		l := scanner.Text()
		if m := packageLine.FindStringSubmatch(l); m != nil {
			return scanPackage(scanner, m)
		}
		if m := tableStartLine.FindStringSubmatch(l); m != nil {
			return scanTable(scanner, m)
		}
	}
	return nil
}

// TODO: It's a bug here that we can't run the prefix and the table defn together in one comment block,
// for that we'll need to ability to unget a line

func scanPackage(scanner *bufio.Scanner, m []string) *packageBlock {
	packageName := m[1]
	imports := make([]string, 0)
	for scanner.Scan() {
		l := scanner.Text()
		if !strings.HasPrefix(l, "//") {
			break
		}
		if m := importLine.FindStringSubmatch(l); m != nil {
			imports = append(imports, m[1])
		}
	}
	return &packageBlock{packageName, imports}
}

func scanTable(scanner *bufio.Scanner, m []string) *tableBlock {
	baseName := m[1]
	recordType := m[2]
	lines := make([]blockLine, 0)

	for scanner.Scan() {
		l := scanner.Text()
		if !strings.HasPrefix(l, "//") {
			break
		}
		if m := tableContentLine.FindStringSubmatch(l); m != nil {
			lines = append(lines, blockLine{m[1], m[2], parseAttrs(m[3])})
		}
		// Otherwise blank which is OK or broken which is not
	}
	return &tableBlock{baseName, recordType, lines}
}

func parseAttrs(s string) map[string]string {
	attrs := make(map[string]string)
	for s != "" {
		m := tableFieldAttr.FindStringSubmatch(s)
		if m == nil {
			break
		}
		attrs[m[1]] = m[2]
		s = s[len(m[0]):]
	}
	// if s not "" then maybe not ok
	return attrs
}
