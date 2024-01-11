// hostnames - compute a set of host names from a set of report file names.
//
// End-user options:
//
//  -report-dir <report-directory>
//  <report-directory> (obsolete form)
//     The name of the directory holding generated reports
//
// Description:
//
// Enumerate the files in the report directory.  For each file name that has the form
// <something>-<word>.json where <word> is from a known set, remember <something>.  Sort the list of
// <something>s lexicographically, remove duplicates, and print the resulting list as a JSON array
// of strings on stdout.
//
// The set of <word>s is: {`minutely`, `daily`, `weekly`, `monthly`, `quarterly`}.
//
// -------------------------------------------------------------------------------------------------
//
// Implementation notes.
//
// Obviously this could be a shell script but it's better integrated into naicreport.
//
// Obviously this could also be implemented by scanning the logs.

package hostnames

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
)

var glob *regexp.Regexp = regexp.MustCompile("^(.*)-(?:minutely|daily|weekly|monthly|quarterly).json$")

func Hostnames(progname string, args []string) (err error) {
	dirname, err := commandLine()
	if err != nil {
		return
	}
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

func commandLine() (dirname string, err error) {
	opts := flag.NewFlagSet(os.Args[0]+" hostnames", flag.ContinueOnError)
	opts.StringVar(&dirname, "report-dir", "", "Report `directory`")
	err = opts.Parse(os.Args[2:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		return
	}
	// Backwards compatible
	if dirname == "" {
		rest := opts.Args()
		if len(rest) == 0 {
			err = errors.New("No report directory specified")
			return
		}
		dirname = rest[0]
	}
	return
}
