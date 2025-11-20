package cmd

import (
	"io"

	. "sonalyze/common"
	"sonalyze/data/sample"
	"sonalyze/db/types"
	"sonalyze/table"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Interfaces that the various commands can implement to respond to various situations.

type FormatHelpAPI interface {
	// If the command accepts a -fmt argument and the value of that argument is "help", return a
	// non-nil object here with formatter help.
	MaybeFormatHelp() *table.FormatHelp
}

type SetRestArgumentsAPI interface {
	// Install any left-over arguments into the arguments object
	SetRestArguments(args []string)
}

var _ = SetRestArgumentsAPI((*DatabaseArgs)(nil))

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Any command of any type must be able to define and validate command line args, and handle some
// developer arguments.

type Command interface {
	// Return the name of the cpu profile file, if requested
	CpuProfileFile() string

	// Documentation, with formatting and line breaks
	Summary(out io.Writer)

	// Add all arguments including shared arguments
	Add(fs *CLI)

	// Validate all arguments including shared arguments
	Validate() error

	// The -v flag
	VerboseFlag() bool

	// Reify all arguments including shared arguments for remote execution, with checking
	ReifyForRemote(x *ArgReifier) error

	// DatabaseArgs API (evolving)
	Dataless() bool
	JobanalyzerDir() string
	DatabaseURI() string
	DataDir() string
	ReportDir() string
	LogFiles() []string
	ConfigFile() string
	CacheSize() int64
	ClusterName() string
	RemoteHost() string
	Remoting() bool
	AuthFile() string
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a generic analysis command that can be run remotely, independently of the data that
// are manipulated.

type AnalysisCommand interface {
	Command
	SetRestArgumentsAPI
	FormatHelpAPI
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a simple command that handles its own logic completely, on no cluster

type PrimitiveCommand interface {
	Command

	Perform(in io.Reader, stdout, stderr io.Writer) error
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a simple command that handles its own logic completely, but on one cluster

type SimpleCommand interface {
	Command

	Perform(meta types.Context, in io.Reader, stdout, stderr io.Writer) error
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a sonalyze "sonar sample" analysis command: jobs, load, parse, etc

type SampleAnalysisParameters interface {
	// Retrieve shared arguments
	SampleAnalysisFlags() *SampleAnalysisArgs

	// Provide appropriate default settings for these flags
	DefaultRecordFilters() (allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool)
}

type SampleAnalysisCommand interface {
	AnalysisCommand
	SampleAnalysisParameters

	// Perform the operation.  The recordFilter has been compiled from the filter.
	Perform(
		out io.Writer,
		meta types.Context,
		filter sample.QueryFilter,
		hosts *Hosts,
		recordFilter *sample.SampleFilter,
	) error

	// Retrieve configfile for those commands that allow it, otherwise "", or "" for absent
	ConfigFile() string
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// This is a container for behavior.  There are two of these: one for the one-shot behavior and one
// for the daemon behavior.  CommandLineHandler is a hack that's really only necessary to deal with
// Go's prohibition against circular package dependencies: the daemon code calls indirect back up
// to the application level, which can then call down to the engine again.

type CommandLineHandler struct {
	// Translate `maybeVerb` into a Command and return a normalized verb.  If the translation failed
	// then `cmd` will be nil and `verb` will be "".  The `cmdName` is the name of the program
	// (argv[0]).
	ParseVerb func(cmdName, maybeVerb string) (cmd Command, verb string)

	// Given a verb and command returned from ParseVerb, and a list of arguments and an empty but
	// otherwise initialized flag set, set up argument parsing, perform it, and validate the result.
	ParseArgs func(verb string, args []string, cmd Command, fs *CLI) error

	// The `profileFile` should be the cpu profile file name in the DevArgs structure.  If not
	// empty, this will start the profiler and return a stop function to be deferred until the end
	// of the program.
	StartCPUProfile func(profileFile string) (func(), error)

	// Given a command initialized with parsed commands, and i/o streams, run the command.
	HandleCommand func(anyCmd Command, stdin io.Reader, stdout, stderr io.Writer) error
}
