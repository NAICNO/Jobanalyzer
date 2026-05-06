package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	fields      []fld
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
		fmt.Fprintf(out, "\t%s %s `json:\"%s,omitempty\"`\n", f.name, f.ty, f.name)
	}
	fmt.Fprintf(out, "}\n\n")
	fmt.Fprintf(out, "func respond(flds *apiutil.FieldMap, r %s) %s {\n", basety, tyname)
	fmt.Fprintf(out, "\tvar x %s\n", tyname)
	for _, f := range fields {
		fmt.Fprintf(out, "\tif flds.Has(\"%s\") {\n", f.name)
		fmt.Fprintf(out, "\t\tx.%s = r.%s\n", f.name, f.name)
		fmt.Fprintf(out, "\t}\n")
	}
	fmt.Fprintf(out, "\treturn x\n")
	fmt.Fprintf(out, "}\n\n")
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
	scanner := bufio.NewScanner(in)
	var found bool
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "FIELDS ") {
			found = true
			basety = strings.TrimSpace(l[7:])
			break
		}
	}
	if !found {
		log.Fatal("No FIELDS line in table " + fn)
	}
	for scanner.Scan() {
		t := scanner.Text()
		if t == "" {
			continue
		}
		for t[len(t)-1] == '\\' {
			if !scanner.Scan() {
				log.Fatal("Missing continuation")
			}
			t = t + scanner.Text()
		}
		x := strings.TrimSpace(t)
		if x == "" {
			continue
		}
		if x[0] == '#' {
			continue
		}
		if t[0] != ' ' {
			break
		}
		fs := strings.Fields(t)
		fields = append(fields, fld{fs[0], fs[1]})
	}
}
