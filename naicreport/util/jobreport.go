// Shared functionality for a multi-line report that is to be sorted by job key.

package util

import (
	"sort"
)

type JobReport struct {
	Id uint32
	Host string
	Report string
}

type byJobKey []*JobReport

func (a byJobKey) Len() int {
	return len(a)
}

func (a byJobKey) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byJobKey) Less(i, j int) bool {
	if a[i].Host != a[j].Host {
		return a[i].Host < a[j].Host
	}
	return a[i].Id < a[j].Id
}

// Sort reports by ascending host name first and job ID second (there could be other criteria).

func SortReports(reports []*JobReport) {
	sort.Sort(byJobKey(reports))
}
