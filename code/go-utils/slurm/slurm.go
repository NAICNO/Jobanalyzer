package slurm

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-utils/filesys"
)

// Get the SLURM job ID from the PID.  Returns -1 if we could not get the information.

func SlurmJobIdFromPid(pid uint) int {
	// We want \1 of the *first* line that matches "/job_(.*?)/", as unsigned.
	//
	// The reason is that there are several lines in that file that look roughly like this,
	// with different contents (except for the job info) but with the pattern the same:
	//
	//    10:devices:/slurm/uid_2101171/job_280678/step_interactive/task_0

	lines, err := filesys.FileLines(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err == nil {
		for _, l := range lines {
			ix := strings.Index(l, "/job_")
			if ix != -1 {
				l = l[ix+5:]
				iy := strings.Index(l, "/")
				if iy != -1 {
					n, err := strconv.ParseUint(strings.TrimSpace(l[:iy]), 10, 32)
					if err == nil && n <= math.MaxInt {
						return int(n)
					}
				}
			}
		}
	}
	return -1
}
