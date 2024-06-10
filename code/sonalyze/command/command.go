package command

import (
	"flag"
	"io"

	"go-utils/config"
	"go-utils/hostglob"
	"sonalyze/sonarlog"
)

type FormatHelpAPI interface {
	// If the command accepts a -fmt argument and the value of that argument is "help", return a
	// non-nil object here with formatter help.
	MaybeFormatHelp() *FormatHelp
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

	// Documentation, one line per string
	Summary() []string

	// Add all arguments including shared arguments
	Add(fs *flag.FlagSet)

	// Validate all arguments including shared arguments
	Validate() error

	// The -v flag
	VerboseFlag() bool
}

type RemotableCommand interface {
	Command

	// Reify all arguments including shared arguments for remote execution, with checking
	ReifyForRemote(x *Reifier) error

	RemotingFlags() *RemotingArgs
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a sonalyze data analysis command: jobs, load, parse, etc

type AnalysisCommand interface {
	SetRestArgumentsAPI
	FormatHelpAPI
	RemotableCommand

	// Retrieve shared arguments
	SharedFlags() *SharedArgs

	// Provide appropriate default settings for these flags
	DefaultRecordFilters() (allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool)

	// Perform the operation, using the filters to select records if appropriate
	Perform(
		out io.Writer,
		cfg *config.ClusterConfig,
		logDir sonarlog.Cluster,
		samples sonarlog.SampleStream,
		hostGlobber *hostglob.HostGlobber,
		recordFilter func(*sonarlog.Sample) bool,
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
	ParseArgs(verb string, args []string, cmd Command, fs *flag.FlagSet) error

	// The `profileFile` should be the cpu profile file name in the DevArgs structure.  If not
	// empty, this will start the profiler and return a stop function to be deferred until the end
	// of the program.
	StartCPUProfile(profileFile string) (func(), error)

	// Given a command initialized with parsed commands, and i/o streams, run the command.
	HandleCommand(anyCmd Command, stdin io.Reader, stdout, stderr io.Writer) error
}
