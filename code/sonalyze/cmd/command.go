package cmd

import (
	"io"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
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
}

type RemotableCommand interface {
	Command

	// Reify all arguments including shared arguments for remote execution, with checking
	ReifyForRemote(x *ArgReifier) error

	RemotingFlags() *RemotingArgsNoCluster
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a generic analysis command that can be run remotely, independently of the data that
// are manipulated.

type AnalysisCommand interface {
	SetRestArgumentsAPI
	FormatHelpAPI
	RemotableCommand
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a simple command that handles its own logic completely

type SimpleCommand interface {
	Command

	Perform(in io.Reader, stdout, stderr io.Writer) error
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a sonalyze "sonar sample" analysis command: jobs, load, parse, etc

type SampleAnalysisCommand interface {
	AnalysisCommand

	// Retrieve shared arguments
	SampleAnalysisFlags() *SampleAnalysisArgs

	// Provide appropriate default settings for these flags
	DefaultRecordFilters() (allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool)

	// Return true if the data ingestion step also should compute time bounds for all the hosts
	// (this is somewhat costly and often not necessary)
	NeedsBounds() bool

	// Perform the operation.  The streams and bounds are as read from the DB using the globber and
	// the filter, but the globber and filter are passed since the client may want to do something
	// more / something different.
	Perform(
		out io.Writer,
		cfg *config.ClusterConfig,
		cluster db.SampleDataProvider,
		streams sonarlog.InputStreamSet,
		bounds sonarlog.Timebounds,
		hostGlobber *Hosts,
		recordFilter *db.SampleFilter,
	) error

	// Retrieve configfile for those commands that allow it, otherwise "", or "" for absent
	ConfigFile() string
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// This is a container for behavior.  It's probably important that it has no mutable state.  There
// could be several.
//
// CommandLineHandler is a hack that's necessary to deal with Go's prohibition against circular
// package dependencies.

type CommandLineHandler interface {
	// Translate `maybeVerb` into a Command and return a normalized verb.  If the translation failed
	// then `cmd` will be nil and `verb` will be "".  The `cmdName` is the name of the program
	// (argv[0]).
	ParseVerb(cmdName, maybeVerb string) (cmd Command, verb string)

	// Given a verb and command returned from ParseVerb, and a list of arguments and an empty but
	// otherwise initialized flag set, set up argument parsing, perform it, and validate the result.
	ParseArgs(verb string, args []string, cmd Command, fs *CLI) error

	// The `profileFile` should be the cpu profile file name in the DevArgs structure.  If not
	// empty, this will start the profiler and return a stop function to be deferred until the end
	// of the program.
	StartCPUProfile(profileFile string) (func(), error)

	// Given a command initialized with parsed commands, and i/o streams, run the command.
	HandleCommand(anyCmd Command, stdin io.Reader, stdout, stderr io.Writer) error
}
