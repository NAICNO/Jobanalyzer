// Various sorting types for use with the old-timey sort.Sort function.  Once we move to Go 1.21 we
// can get rid of all of this.

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

// Sort SampleStreams by host, time, job, and command (in that order)

type HostTimeJobCmdSortableSampleStreams SampleStreams

func (sss HostTimeJobCmdSortableSampleStreams) Len() int {
	return len(sss)
}

func (sss HostTimeJobCmdSortableSampleStreams) Swap(i, j int) {
	sss[i], sss[j] = sss[j], sss[i]
}

func (sss HostTimeJobCmdSortableSampleStreams) Less(i, j int) bool {
	if (*sss[i])[0].Host == (*sss[j])[0].Host {
		if (*sss[i])[0].Timestamp == (*sss[j])[0].Timestamp {
			if (*sss[i])[0].Job == (*sss[j])[0].Job {
				return (*sss[i])[0].Cmd.String() < (*sss[j])[0].Cmd.String()
			}
			return (*sss[i])[0].Job < (*sss[j])[0].Job
		} else {
			return (*sss[i])[0].Timestamp < (*sss[j])[0].Timestamp
		}
	} else {
		return (*sss[i])[0].Host.String() < (*sss[j])[0].Host.String()
	}
}

// Sort SampleStreams by host

type HostSortableSampleStreams SampleStreams

func (sss HostSortableSampleStreams) Len() int {
	return len(sss)
}

func (sss HostSortableSampleStreams) Swap(i, j int) {
	sss[i], sss[j] = sss[j], sss[i]
}

func (sss HostSortableSampleStreams) Less(i, j int) bool {
	return (*sss[i])[0].Host.String() < (*sss[j])[0].Host.String()
}

// Sort SampleStreams by time

type TimeSortableSampleStreams SampleStreams

func (sss TimeSortableSampleStreams) Len() int {
	return len(sss)
}

func (sss TimeSortableSampleStreams) Swap(i, j int) {
	sss[i], sss[j] = sss[j], sss[i]
}

func (sss TimeSortableSampleStreams) Less(i, j int) bool {
	return (*sss[i])[0].Timestamp < (*sss[j])[0].Timestamp
}
