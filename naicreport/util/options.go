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
	ut "go-utils/time"
)

type SonarLogOptions struct {
	DataPath  string
	DataFiles []string // For -- filename ...
	HaveFrom  bool
	From      time.Time
	FromStr   string
	HaveTo    bool
	To        time.Time
	ToStr     string
}

func AddSonarLogOptions(opts *flag.FlagSet) *SonarLogOptions {
	logOpts := SonarLogOptions{
		DataPath:  "",
		DataFiles: nil,
		HaveFrom:  false,
		From:      time.Now(),
		FromStr:   "",
		HaveTo:    false,
		To:        time.Now(),
		ToStr:     "",
	}
	opts.StringVar(&logOpts.DataPath, "data-path", "", "Root `directory` of data store (required)")
	opts.StringVar(&logOpts.FromStr, "from", "1d",
		"Start `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	opts.StringVar(&logOpts.ToStr, "to", "",
		"End `date` of log window, yyyy-mm-dd or Nd (days ago) or Nw (weeks ago)")
	return &logOpts
}

func RectifySonarLogOptions(s *SonarLogOptions, opts *flag.FlagSet) error {
	// Figure out files
	var err error
	if s.DataPath == "" {
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
		s.DataFiles = files
	} else {
		// Clean the DataPath and make it absolute.
		s.DataPath, err = options.RequireCleanPath(s.DataPath, "-data-path")
		if err != nil {
			return err
		}
	}

	// Figure out the date range.  From has a sane default so always parse; To has no default so
	// grab current day if nothing is specified.

	s.HaveFrom = true
	s.From, err = ut.ParseRelativeDate(s.FromStr)
	if err != nil {
		return err
	}

	if s.ToStr == "" {
		s.To = time.Now().UTC()
	} else {
		s.HaveTo = true
		s.To, err = ut.ParseRelativeDate(s.ToStr)
		if err != nil {
			return err
		}
	}

	// For To, we really want tomorrow's date because the date range is not inclusive on the right.

	s.To = s.To.AddDate(0, 0, 1)
	s.To = time.Date(s.To.Year(), s.To.Month(), s.To.Day(), 0, 0, 0, 0, time.UTC)

	return nil
}

// Simple utility: Forward standard from/to and data file arguments to an argument list

func ForwardSonarLogOptions(arguments []string, progOpts *SonarLogOptions) []string {
	if progOpts.HaveFrom {
		arguments = append(arguments, "--from", progOpts.FromStr)
	}
	if progOpts.HaveTo {
		arguments = append(arguments, "--to", progOpts.ToStr)
	}
	if progOpts.DataFiles != nil {
		arguments = append(arguments, "--")
		arguments = append(arguments, progOpts.DataFiles...)
	} else {
		arguments = append(arguments, "--data-path", progOpts.DataPath)
	}
	return arguments
}
