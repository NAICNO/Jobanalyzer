package profile

import (
	"errors"
	"flag"

	. "sonalyze/command"
)

type ProfileCommand struct /* implements Command */ {
	SharedArgs

	// Filtering and aggregation
	Max    float64
	Bucket uint

	// Printing
	Fmt string

	// Synthesized and other
	printFields  []string
	printOpts    *FormatOptions
	htmlOutput   bool
	testNoMemory bool
}

func (pc *ProfileCommand) Add(fs *flag.FlagSet) {
	pc.SharedArgs.Add(fs)

	fs.Float64Var(&pc.Max, "max", 0, "Clamp values to this (helps deal with noise) (memory in GiB)")
	fs.UintVar(&pc.Bucket, "bucket", 0, "Bucket these many consecutive elements (helps reduce noise)")
	fs.StringVar(&pc.Fmt, "fmt", "",
		"Select `field,...` and format for the output [default: try -fmt=help]")
}

func (pc *ProfileCommand) ReifyForRemote(x *Reifier) error {
	e1 := pc.SharedArgs.ReifyForRemote(x)
	x.Float64("max", pc.Max)
	x.Uint("bucket", pc.Bucket)
	x.String("fmt", pc.Fmt)
	return e1
}

func (pc *ProfileCommand) Validate() error {
	e1 := pc.SharedArgs.Validate()

	var e2 error
	if pc.Max < 0 {
		e2 = errors.New("Invalid -max, must be nonnegative")
	}

	var e3 error
	spec := profileDefaultFields
	if pc.Fmt != "" {
		spec = pc.Fmt
	}
	var others map[string]bool
	pc.printFields, others, e3 = ParseFormatSpec(spec, profileFormatters, profileAliases)
	if e3 == nil && len(pc.printFields) == 0 && !others["json"] {
		e3 = errors.New("No output fields were selected in format string")
	}
	pc.printOpts = StandardFormatOptions(others)
	pc.htmlOutput = others["html"]
	pc.testNoMemory = others["nomemory"]

	// Ad-hoc defaulting
	if pc.printOpts.Csv && !others["noheader"] {
		pc.printOpts.Header = true
	}
	if pc.htmlOutput && !others["header"] {
		pc.printOpts.Header = false
	}

	var e4 error
	if len(pc.Job) != 1 {
		e4 = errors.New("Exactly one specific job number is required by `profile`")
	}

	// html is not part of the usual set of formats (yet)
	var e5 error
	if pc.htmlOutput {
		if pc.printOpts.Fixed || pc.printOpts.Csv || pc.printOpts.Json && pc.printOpts.Awk {
			e5 = errors.New("Multiple printing options specified")
		} else if pc.printOpts.Header {
			e5 = errors.New("Can't use `header` and `html` together")
		}
	} else {
		// TODO: ABSTRACTME: This is the same defaulting logic as in jobs and parse except it does
		// not toggle header.
		if !pc.printOpts.Fixed && !pc.printOpts.Csv && !pc.printOpts.Json && !pc.printOpts.Awk {
			pc.printOpts.Fixed = true
		}
	}

	var e6 error
	if pc.printOpts.Named {
		e6 = errors.New("Named fields are not supported for `profile`")
	}
	pc.printOpts.NoDefaults = false

	return errors.Join(e1, e2, e3, e4, e5, e6)
}

func (pc *ProfileCommand) ConfigFile() string {
	// No config file for profile
	return ""
}
