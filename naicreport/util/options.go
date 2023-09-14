// TODO: allow -f and -t as abbreviations for --from and --to since sonalyze allows this.  How?  The
// syntax may still not be quite compatible, sonalyze allows eg -f1d which would not work here.

package util

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

// A container for some common options and a FlagSet that can be extended with more options.  For
// --from and --to there's both the computed from/to time and the input strings (after vetting).

type StandardOptions struct {
	Container *flag.FlagSet
	DataPath string				// After parsing: absolute, Cleaned path
	HaveFrom bool				// After parsing: true if there's a --from argument
	From time.Time				// After parsing: UTC timestamp
	FromStr string				// After parsing: the --from input string, if present
	HaveTo bool					// After parsing: true if there's a --to argument
	To time.Time				// After parsing: UTC timestamp
	ToStr string				// After parsing: the --to input string, if present
	Verbose bool				// After parsing: flag

	verbosePtr *bool
	dataPathPtr *string
	fromPtr *string				// Unparsed 'from' value
	toPtr *string				// Unparsed 'to' value
}

// The idea is that the program calls NewStandardOptions to get a structure with standard options
// added to the FlagSet, and with some helpers to parse the arguments.  The program can add more
// flags to opts.container before calling the parser (saving the the flag pointers elsewhere) so
// that the parsing of everything is properly integrated.

func NewStandardOptions(progname string) *StandardOptions {
	container := flag.NewFlagSet(progname+" ml-cpuhog", flag.ExitOnError)
	dataPathPtr := container.String("data-path", "", "Root directory of data store (required)")
	fromPtr := container.String("from", "1d", "Start of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	toPtr := container.String("to", "", "End of log window, ditto")
	verbosePtr := container.Bool("v", false, "Verbose (debugging) output")

	return &StandardOptions {
		Container: container,
		DataPath: "",
		HaveFrom: false,
		From: time.Now(),
		FromStr: "",
		HaveTo: false,
		To: time.Now(),
		ToStr: "",
		Verbose: false,

		dataPathPtr: dataPathPtr,
		fromPtr: fromPtr,
		toPtr: toPtr,
		verbosePtr: verbosePtr,
	}
}

func (s *StandardOptions) Parse(args []string) error {
	err := s.Container.Parse(args)
	if err != nil {
		return err
	}

	// Copy anything copyable

	s.Verbose = *s.verbosePtr

	// Clean the DataPath and make it absolute.

	s.DataPath, err = CleanPath(*s.dataPathPtr, "-data-path")
	if err != nil {
		return err
	}

	// Figure out the date range.  From has a sane default so always parse; To has no default so
	// grab current day if nothing is specified.

	s.HaveFrom = true
	s.FromStr = *s.fromPtr
	s.From, err = matchWhen(*s.fromPtr)
	if err != nil {
		return err
	}

	if *s.toPtr == "" {
		s.To = time.Now().UTC()
	} else {
		s.HaveTo = true
		s.ToStr = *s.toPtr
		s.To, err = matchWhen(*s.toPtr)
		if err != nil {
			return err
		}
	}

	// For To, we really want tomorrow's date because the date range is not inclusive on the right.

	s.To = s.To.AddDate(0, 0, 1)
	s.To = time.Date(s.To.Year(), s.To.Month(), s.To.Day(), 0, 0, 0, 0, time.UTC)

	return nil
}

func CleanPath(p, optionName string) (newp string, e error) {
	if p == "" {
		e = errors.New(fmt.Sprintf("%s requires a value", optionName))
	} else if path.IsAbs(p) {
		newp = path.Clean(p)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			e = err
		} else {
			newp = path.Join(wd, p)
		}
	}
	return
}

// The format of `from` and `to` is one of:
//  YYYY-MM-DD
//  Nd (days ago)
//  Nw (weeks ago)

var dateRe = regexp.MustCompile(`^(\d\d\d\d)-(\d\d)-(\d\d)$`)
var daysRe = regexp.MustCompile(`^(\d+)d$`)
var weeksRe = regexp.MustCompile(`^(\d+)w$`)

func matchWhen(s string) (time.Time, error) {
	probe := dateRe.FindSubmatch([]byte(s))
	if probe != nil {
		yyyy, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		mm, _ := strconv.ParseUint(string(probe[2]), 10, 32)
		dd, _ := strconv.ParseUint(string(probe[3]), 10, 32)
		return time.Date(int(yyyy), time.Month(mm), int(dd), 0, 0, 0, 0, time.UTC), nil
	}
	probe = daysRe.FindSubmatch([]byte(s))
	if probe != nil {
		days, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		t := time.Now().UTC().AddDate(0, 0, -int(days))
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	probe = weeksRe.FindSubmatch([]byte(s))
	if probe != nil {
		weeks, _ := strconv.ParseUint(string(probe[1]), 10, 32)
		t := time.Now().UTC().AddDate(0, 0, -int(weeks)*7)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	return time.Now(), errors.New("Bad time specification")
}

