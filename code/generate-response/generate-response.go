package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"go-utils/table/parser"
)

var (
	outName = flag.String("o", "respond.go", "Output `filename`")
)

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Exactly one input file required")
		os.Exit(2)
	}
	inName := flag.Args()[0]
	in, err := os.Open(inName)
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(*outName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(out, "// Generated from %s by generate-response.  DO NOT EDIT.\n\n", inName)
	scanner := bufio.NewScanner(in)
	var inside bool
	for scanner.Scan() {
		l := scanner.Text()
		if inside {
			if strings.HasPrefix(l, "ESNOPSER*/") {
				inside = false
				end()
			} else {
				process(l)
			}
		} else {
			if strings.HasPrefix(l, "/*RESPONSE") {
				inside = true
				start(out)
			}
		}
	}
}

type fld struct {
	name string
	ty   string
}

var (
	out         *os.File
	prefix      bool
	hasDefaults bool
	fields      []parser.Field
	tyname      string
	basety      string
)

func start(out_ *os.File) {
	out = out_
	prefix = true
	hasDefaults = false
	fields = nil
	tyname = ""
	basety = ""
}

func end() {
	if tyname == "" || fields == nil || basety == "" {
		log.Fatal("Incomplete information")
	}
	fmt.Fprintf(out, "type %s struct {\n", tyname)
	for _, f := range fields {
		fmt.Fprintf(out, "\t%s %s `json:\"%s,omitempty\"`\n", f.Name, f.Type, f.Name)
	}
	fmt.Fprintf(out, "}\n\n")
	fmt.Fprintf(out, "func respond(flds *apiutil.FieldMap, r %s) %s {\n", basety, tyname)
	fmt.Fprintf(out, "\tvar x %s\n", tyname)
	for _, f := range fields {
		fmt.Fprintf(out, "\tif flds.Has(\"%s\") {\n", f.Name)
		actualFieldName := f.Name
		if fattr, found := attr(f.Attrs, "field"); found {
			actualFieldName = fattr.Value
		}
		indir, isIndirect := attr(f.Attrs, "indirect")
		if isIndirect {
			ptrName := indir.Value
			fmt.Fprintf(out, "\t\tif (r.%s) != nil {\n", ptrName)
			fmt.Fprintf(
				out, "\t\t\tx.%s = r.%s.%s\n", f.Name, ptrName, actualFieldName)
			fmt.Fprintf(out, "\t\t}\n")
		} else {
			fmt.Fprintf(out, "\t\tx.%s = r.%s\n", f.Name, actualFieldName)
		}
		fmt.Fprintf(out, "\t}\n")
	}
	fmt.Fprintf(out, "\treturn x\n")
	fmt.Fprintf(out, "}\n\n")
}

func attr(attrs []parser.NV, name string) (parser.NV, bool) {
	for _, a := range attrs {
		if a.Name == name {
			return a, true
		}
	}
	return parser.NV{}, false
}

func process(l string) {
	if prefix {
		if strings.HasPrefix(l, "%%") {
			prefix = false
			return
		}
		fmt.Fprintln(out, l)
		return
	}
	if strings.HasPrefix(l, "TYPE ") {
		if tyname != "" {
			log.Fatal("Duplicate TYPE")
		}
		tyname = strings.TrimSpace(l[5:])
		return
	}
	if strings.HasPrefix(l, "TABLE ") {
		if fields != nil {
			log.Fatal("Duplicate TABLE")
		}
		fn := strings.TrimSpace(l[6:])
		getFields(fn)
		return
	}
	if strings.HasPrefix(l, "DEFAULTS ") {
		if hasDefaults {
			log.Fatal("Duplicate DEFAULTS")
		}
		hasDefaults = true
		fmt.Fprintf(out, "const responseDefaults = \"%s\"\n", strings.TrimSpace(l[9:]))
		return
	}
	if strings.TrimSpace(l) != "" {
		log.Print("WARNING: Junk: " + l)
	}
}

func getFields(fn string) {
	in, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
	p := parser.NewParser(fn, in)
	block, err := p.Parse()
	if err != nil {
		log.Fatal(err)
	}
	basety = block.Fields.Type
	fields = block.Fields.Fields
}
