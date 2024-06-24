package slurm

import (
	"errors"
	"flag"

	// . "sonalyze/common"
	. "sonalyze/command"
)

// This will be very similar in structure to `sonalyze top` because it does not do samples.  So
// perhaps there are commonalities to exploit?  It's like AnalysisCommand is SampleAnalysisCommand
// and top and slurm are SystemDataAnalysisCommand?
//
// The rest arguments... probably these could be used to source the data, useful for testing?

type SlurmCommand struct /* implements AnalysisCommand */ {
	// Almost SharedArgs, but HostArgs instead of RecordFilterArgs
	DevArgs
	SourceArgs
	HostArgs
	VerboseArgs
	ConfigFileArgs

	Modules bool
	Projects bool
	Failed bool					// a filtering option - only failed jobs (exit != 0)
	Array bool
	Stepped bool
	Het bool
}

var _ = AnalysisCommand((*SlurmCommand)(nil))

func (_ *SlurmCommand) Summary() []string {
	return []string{
		"Extract information from slurm data independent of sample data",
	}
}

func (sc *SlurmCommand) Add(fs *flag.FlagSet) {
	sc.DevArgs.Add(fs)
	sc.SourceArgs.Add(fs)
	sc.HostArgs.Add(fs)
	sc.VerboseArgs.Add(fs)
	sc.ConfigFileArgs.Add(fs)
	fs.BoolVar(&sc.Modules, "modules", false, "Display information about modules loaded")
	fs.BoolVar(&sc.Modules, "projects", false, "Display information about active projects")
	// TODO: More
}

func (sc *SlurmCommand) Validate() error {
	var e1, e2, e3, e4, e5 error
	e1 = sc.DevArgs.Validate()
	e2 = sc.SourceArgs.Validate()
	e3 = sc.HostArgs.Validate()
	e4 = sc.VerboseArgs.Validate()
	e5 = sc.ConfigFileArgs.Validate()
	// TODO: We need exactly one of -modules, -projects
	// TODO: More
	return errors.Join(e1, e2, e3, e4, e5)
}

func (sc *SlurmCommand) ReifyForRemote(x *Reifier) error {
	x.Bool("modules", sc.Modules)
	x.Bool("projects", sc.Projects)
	// TODO: More
	// tc.Verbose is not reified, as for SharedArgs.
	return errors.Join(
		sc.DevArgs.ReifyForRemote(x),
		sc.SourceArgs.ReifyForRemote(x),
		sc.HostArgs.ReifyForRemote(x),
		sc.ConfigFileArgs.ReifyForRemote(x),
	)
}

func (sc *SlurmCommand) MaybeFormatHelp() *FormatHelp {
	// FIXME, but currently no format options at all
	return nil
}
