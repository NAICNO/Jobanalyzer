// Various sorting types for use with the old-timey sort.Sort function.

package sonarlog

// Sort SampleStream by time

type TimeSortableSampleStream SampleStream

func (t TimeSortableSampleStream) Len() int { return len(t) }

func (t TimeSortableSampleStream) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TimeSortableSampleStream) Less(i, j int) bool {
	return t[i].Timestamp < t[j].Timestamp
}

// Sort SampleStream by host (primary) and time (secondary)

type HostTimeSortableSampleStream SampleStream

func (ss HostTimeSortableSampleStream) Len() int {
	return len(ss)
}

func (ss HostTimeSortableSampleStream) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

func (ss HostTimeSortableSampleStream) Less(i, j int) bool {
	if ss[i].Host != ss[j].Host {
		return ss[i].Host.String() < ss[j].Host.String()
	}
	return ss[i].Timestamp < ss[j].Timestamp
}


