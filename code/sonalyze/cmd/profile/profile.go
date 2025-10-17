package profile

import (
	"errors"

	. "sonalyze/cmd"
	. "sonalyze/table"
)

type ProfileCommand struct /* implements SampleAnalysisCommand */ {
	SampleAnalysisArgs
	FormatArgs

	// Filtering and aggregation
	Max    float64
	Bucket uint

	// Synthesized and other
	htmlOutput   bool
	testNoMemory bool
}

var _ = SampleAnalysisCommand((*ProfileCommand)(nil))

func (pc *ProfileCommand) Add(fs *CLI) {
	pc.SampleAnalysisArgs.Add(fs)
	pc.FormatArgs.Add(fs)

	fs.Group("printing")
	fs.Float64Var(&pc.Max, "max", 0, "Clamp values to this (helps deal with noise) (memory in GiB)")

	fs.Group("aggregation")
	fs.UintVar(&pc.Bucket, "bucket", 0, "Bucket these many consecutive elements (helps reduce noise)")
}

func (pc *ProfileCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := errors.Join(
		pc.SampleAnalysisArgs.ReifyForRemote(x),
		pc.FormatArgs.ReifyForRemote(x),
	)
	x.Float64("max", pc.Max)
	x.Uint("bucket", pc.Bucket)
	return e1
}

func (pc *ProfileCommand) Validate() error {
	// FormatArgs are handled specially below
	e1 := pc.SampleAnalysisArgs.Validate()

	var e2 error
	if pc.Max < 0 {
		e2 = errors.New("Invalid -max, must be nonnegative")
	}

	var e3 error
	if len(pc.Job) != 1 || len(pc.ExcludeJob) != 0 {
		e3 = errors.New("Exactly one specific job number is required by `profile`")
	}

	var others map[string]bool
	var e4 error
	pc.PrintFields, others, e4 = ParseFormatSpec(
		profileDefaultFields,
		pc.Fmt,
		profileFormatters,
		profileAliases,
	)
	if e4 == nil && len(pc.PrintFields) == 0 && !others["json"] {
		e4 = errors.New("No valid output fields were selected in format string")
	}

	// Options for profile are restrictive, and a little wonky because html is handled on the side,
	// but mostly we don't error out for nonsensical settings, we just override or ignore them.
	pc.htmlOutput = others["html"]
	pc.testNoMemory = others["nomemory"]

	def := DefaultFixed
	if pc.htmlOutput {
		def = DefaultNone
	}
	pc.PrintOpts = StandardFormatOptions(others, def)

	var e5 error
	if pc.htmlOutput && !pc.PrintOpts.IsDefaultFormat() {
		e5 = errors.New("Multiple output formats requested")
	}

	// The printing code uses custom logic for everything but Fixed layout, and the custom logic
	// does not support named fields or nodefaults.  Indeed the "profile" is always a fixed matrix
	// of data, so nodefaults is disabled even for Fixed.
	pc.PrintOpts.NoDefaults = false

	// The Header setting is grandfathered from the Rust code, but it makes more sense than the
	// opposite.  The main reason to not perpetuate this hack is that it is different from all the
	// other commands.
	if pc.PrintOpts.Csv && !others["noheader"] {
		pc.PrintOpts.Header = true
	}

	return errors.Join(e1, e2, e3, e4, e5)
}

func (pc *ProfileCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	allUsers, skipSystemUsers, determined := pc.RecordFilterArgs.DefaultUserFilters()
	if !determined {
		allUsers, skipSystemUsers = false, false
		if pc.QueryStmt != "" {
			allUsers = true
		}
	}
	excludeSystemCommands = false
	excludeHeartbeat = true
	return
}
