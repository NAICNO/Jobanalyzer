package profile

import (
	"errors"
	"flag"

	. "sonalyze/command"
)

type ProfileCommand struct /* implements AnalysisCommand */ {
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

func (_ *ProfileCommand) Summary() []string {
	return []string{
		"Print profile information for one aspect of a particular job.",
	}
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
	if len(pc.Job) != 1 {
		e3 = errors.New("Exactly one specific job number is required by `profile`")
	}

	var others map[string]bool
	var e4 error
	pc.printFields, others, e4 = ParseFormatSpec(profileDefaultFields, pc.Fmt, profileFormatters, profileAliases)
	if e4 == nil && len(pc.printFields) == 0 && !others["json"] {
		e4 = errors.New("No output fields were selected in format string")
	}

	// Options for profile are restrictive, and a little wonky because html is handled on the side,
	// but mostly we don't error out for nonsensical settings, we just override or ignore them.
	pc.htmlOutput = others["html"]
	pc.testNoMemory = others["nomemory"]

	def := DefaultFixed
	if pc.htmlOutput {
		def = DefaultNone
	}
	pc.printOpts = StandardFormatOptions(others, def)

	var e5 error
	if pc.htmlOutput && !pc.printOpts.IsDefaultFormat() {
		e5 = errors.New("Multiple formats requested")
	}

	// The printing code uses custom logic for everything but Fixed layout, and the custom logic
	// does not support named fields or nodefaults.  Indeed the "profile" is always a fixed matrix
	// of data, so nodefaults is disabled even for Fixed.
	pc.printOpts.NoDefaults = false

	// The Header setting is grandfathered from the Rust code, but it makes more sense than the
	// opposite.  The main reason to not perpetuate this hack is that it is different from all the
	// other commands.
	if pc.printOpts.Csv && !others["noheader"] {
		pc.printOpts.Header = true
	}

	return errors.Join(e1, e2, e3, e4, e5)
}

func (pc *ProfileCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	allUsers, skipSystemUsers, determined := pc.RecordFilterArgs.DefaultUserFilters()
	if !determined {
		allUsers, skipSystemUsers = false, false
	}
	excludeSystemCommands = false
	excludeHeartbeat = true
	return
}
