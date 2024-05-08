package command

import (
	"flag"
	"io"

	"go-utils/config"
	"go-utils/hostglob"
	"sonalyze/sonarlog"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Represents a sonalyze command: jobs, load, parse, etc

type Command interface {
	// Retrieve shared arguments
	Args() *SharedArgs

	// Add all arguments including shared arguments
	Add(fs *flag.FlagSet)

	// Validate all arguments including shared arguments
	Validate() error

	// Reify all arguments including shared arguments for remote execution, with checking
	ReifyForRemote(x *Reifier) error

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

	// If the command accepts a -fmt argument and the value of that argument is "help", return a
	// non-nil object here with formatter help.
	MaybeFormatHelp() *FormatHelp

	// Retrieve configfile for those commands that allow it, otherwise "", or "" for absent
	ConfigFile() string
}
