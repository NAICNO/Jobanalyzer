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
		logStore *sonarlog.LogStore,
		samples sonarlog.SampleStream,
		hostGlobber *hostglob.HostGlobber,
		recordFilter func(*sonarlog.Sample) bool,
	) error

	// Retrieve configfile for those commands that allow it, otherwise "", or "" for absent
	ConfigFile() string
}
