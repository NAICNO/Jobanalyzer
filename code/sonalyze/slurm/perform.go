package slurm

import (
	"io"
)

func (sc *SlurmCommand) Slurm(_ io.Reader, out, _ io.Writer) error {
	// For "modules", we want to look at data for the given date range and for each job, extract the
	// modules loaded by the slurm script.  We want to build a table of these with the number of
	// uses (probably).  Sort this descending by uses and print it.

	// For "projects", process each slurm script to get the project number and maybe look at each
	// job to get duration and then maybe sum number of jobs and duration to present with each
	// project, maybe sorted descending by duration.
	panic("NYI")
}
