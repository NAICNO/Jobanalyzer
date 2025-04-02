package tree

import (
	"errors"

	. "sonalyze/cmd"
	. "sonalyze/table"
)

type TreeCommand struct /* implements SampleAnalysisCommand */ {
	SampleAnalysisArgs
	FormatArgs
}

var _ SampleAnalysisCommand = (*TreeCommand)(nil)

func (pc *TreeCommand) Add(fs *CLI) {
	pc.SampleAnalysisArgs.Add(fs)
	pc.FormatArgs.Add(fs)
}

func (pc *TreeCommand) ReifyForRemote(x *ArgReifier) error {
	return errors.Join(
		pc.SampleAnalysisArgs.ReifyForRemote(x),
		pc.FormatArgs.ReifyForRemote(x),
	)
}

func (pc *TreeCommand) Validate() error {
	var e1, e2 error
	e1 = errors.Join(
		pc.SampleAnalysisArgs.Validate(),
		pc.FormatArgs.Validate(),
	)
	if len(pc.Job) != 1 {
		e2 = errors.New("Exactly one specific job number is required by `tree`")
	}
	return errors.Join(e1, e2)
}

func (pc *TreeCommand) DefaultRecordFilters() (
	allUsers, skipSystemUsers, excludeSystemCommands, excludeHeartbeat bool,
) {
	// Same as for `profile`
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

func (jc *TreeCommand) NeedsBounds() bool {
	return false
}

func (pc *TreeCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	_ db.SampleCluster,
	streams sonarlog.InputStreamSet,
	_ sonarlog.Timebounds,
	_ *hostglob.HostGlobber,
	_ *db.SampleFilter,
) error {
	jobId := pc.Job[0]

	if len(streams) == 0 {
		return fmt.Errorf("No processes matching job ID(s): %v", pc.Job)
	}

	parents := make(map[uint32]uint32)
	for _, s := range streams {
		if _, found := parents[s[0].Pid]; !found {
			parents[s[0].Pid] = s[0].Ppid
		}
	}

	// Note, there's a risk that there's a forest here
}
