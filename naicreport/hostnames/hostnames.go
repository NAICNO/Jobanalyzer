// Hostnames: Compute a set of host names from a set of log file names.
//
// Enumerate the files in the log directory.  For each that has the form <something>-<word>.json
// where <word> is from a known set, remember <something>.  Sort the list lexicographically, remove
// duplicates, and print the list as a JSON array of strings on stdout.
//
// Obviously this could be a shell script but it's better integrated into naicreport.
//
// Obviously this could also be implemented by scanning the logs.

package hostnames

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
)

var glob *regexp.Regexp = regexp.MustCompile("^(.*)-(?:minutely|daily|weekly|monthly|quarterly).json$")

func Hostnames(progname string, args []string) (err error) {
	if len(args) != 1 {
		err = errors.New("Requires one argument: the log directory")
		return
	}
	if args[0] == "-h" {
		fmt.Fprintf(
			os.Stderr,
			`Usage of %s hostnames:
    %s hostnames <dirname>
      <dirname> names the directory containing the load reports
`,
			progname,
			progname,
		)
		return
	}
	dirname := args[0]
	err = hostnames(dirname)
	return
}

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
