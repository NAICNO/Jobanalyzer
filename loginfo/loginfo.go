// Simple utility that is used to produce various information about the generated logs, in a format
// usable to various consumers.
//
// Usage:
//
// loginfo hostnames <directory>
//   Enumerate the files in the directory.  For each name that has the form <something>-<word>.json
//   where <word> is from a known set, remember <something>.  Finally, sort the list
//   lexicographically, remove duplicates, and print the list as a JSON array of strings on
//   stdout.

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
)

func main() {
	if len(os.Args) < 2 {
		toplevelUsage(1)
	}
	var err error
	switch os.Args[1] {
	case "help":
		toplevelUsage(0)

	case "hostnames":
		if len(os.Args) < 3 {
			toplevelUsage(1)
		}
		err = hostnames(os.Args[2])

	default:
		toplevelUsage(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n\n", err)
		toplevelUsage(1)
	}
}

func toplevelUsage(code int) {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s <verb> <option> ...\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "where <verb> is one of\n\n")
	fmt.Fprintf(os.Stderr, "  help\n")
	fmt.Fprintf(os.Stderr, "    Print help\n\n")
	fmt.Fprintf(os.Stderr, "  hostnames <dir>\n")
	fmt.Fprintf(os.Stderr, "    Analyze the names of log files in <dir> to generate a list of host names\n\n")
	os.Exit(code)
}

var glob *regexp.Regexp = regexp.MustCompile("^(.*)-(?:minutely|daily|weekly|monthly|quarterly).json$")

func hostnames(dirname string) (err error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return
	}
	files := make(map[string]bool)
	for _, entry := range entries {
		if (entry.Type() & fs.ModeType) != 0 {
			continue
		}
		// File, we hope
		if matches := glob.FindStringSubmatch(entry.Name()); matches != nil {
			hostname := matches[1]
			files[hostname] = true
		}
	}
	hosts := make(sort.StringSlice, 0, len(files))
	for k, _ := range files {
		hosts = append(hosts, k)
	}
	hosts.Sort()
	encoding, err := json.Marshal(hosts)
	if err == nil {
		fmt.Println(string(encoding))
	}
	return
}
