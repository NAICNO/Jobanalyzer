// Options parser utilities for naicreport.
//
// TODO: allow -f and -t as abbreviations for --from and --to since sonalyze allows this.  How?  The
// syntax may still not be quite compatible, sonalyze allows eg -f1d which would not work here.

package util

import (
	"errors"
	"flag"
	"time"

	"go-utils/options"
	gut "go-utils/time"
)

type DataFilesOptions struct {
	Path  string
	Files []string // For -- filename ...
	Name  string   // Option name
}

func AddDataFilesOptions(opts *flag.FlagSet, canonicalName, explanation string) *DataFilesOptions {
	logOpts := DataFilesOptions{
		Path:  "",
		Files: nil,
		Name:  canonicalName,
	}
	opts.StringVar(&logOpts.Path, canonicalName, "", explanation)
	return &logOpts
}

func RectifyDataFilesOptions(s *DataFilesOptions, opts *flag.FlagSet) error {
	// Figure out files
	var err error
	if s.Path == "" {
		files := []string{}
		for _, f := range opts.Args() {
			fn, err := options.RequireCleanPath(f, "--")
			if err != nil {
				return err
			}
			files = append(files, fn)
		}
		if len(files) == 0 {
			return errors.New("No file arguments provided")
		}
		s.Files = files
	} else {
		// Clean the Path and make it absolute.
		s.Path, err = options.RequireCleanPath(s.Path, s.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func ForwardDataFilesOptions(arguments []string, optName string, opts *DataFilesOptions) []string {
	if opts.Files != nil {
		arguments = append(arguments, "--")
		arguments = append(arguments, opts.Files...)
	} else {
		arguments = append(arguments, optName, opts.Path)
	}
	return arguments
}

type DateFilterOptions struct {
	HaveFrom bool
	From     time.Time
	FromStr  string
	HaveTo   bool
	To       time.Time
	ToStr    string
}

func AddDateFilterOptions(opts *flag.FlagSet) *DateFilterOptions {
	logOpts := DateFilterOptions{
		HaveFrom: false,
		From:     time.Now(),
		FromStr:  "",
		HaveTo:   false,
		To:       time.Now(),
		ToStr:    "",
	}
	opts.StringVar(&logOpts.FromStr, "from", "1d",
		"Start `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	opts.StringVar(&logOpts.ToStr, "to", "",
		"End `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	return &logOpts
}

func RectifyDateFilterOptions(s *DateFilterOptions, opts *flag.FlagSet) error {
	var err error

	// Figure out the date range.  From has a sane default so always parse; To has no default so
	// grab current day if nothing is specified.

	s.HaveFrom = true
	s.From, err = gut.ParseRelativeDate(s.FromStr)
	if err != nil {
		return err
	}
	// Strip h/m/s
	s.From = gut.ThisDay(s.From)

	if s.ToStr == "" {
		s.To = time.Now().UTC()
	} else {
		s.HaveTo = true
		s.To, err = gut.ParseRelativeDate(s.ToStr)
		if err != nil {
			return err
		}
	}

	// For To, we want tomorrow's date because the date range is not inclusive on the right.  Then
	// strip h/m/s.
	s.To = gut.NextDay(s.To)

	return nil
}

func ForwardDateFilterOptions(arguments []string, progOpts *DateFilterOptions) []string {
	if progOpts.HaveFrom {
		arguments = append(arguments, "--from", progOpts.FromStr)
	}
	if progOpts.HaveTo {
		arguments = append(arguments, "--to", progOpts.ToStr)
	}
	return arguments
}
