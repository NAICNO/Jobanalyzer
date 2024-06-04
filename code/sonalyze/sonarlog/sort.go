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
	return t[i].S.Timestamp < t[j].S.Timestamp
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
	if ss[i].S.Host != ss[j].S.Host {
		return ss[i].S.Host.String() < ss[j].S.Host.String()
	}
	return ss[i].S.Timestamp < ss[j].S.Timestamp
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
	if (*sss[i])[0].S.Host == (*sss[j])[0].S.Host {
		if (*sss[i])[0].S.Timestamp == (*sss[j])[0].S.Timestamp {
			if (*sss[i])[0].S.Job == (*sss[j])[0].S.Job {
				return (*sss[i])[0].S.Cmd.String() < (*sss[j])[0].S.Cmd.String()
			}
			return (*sss[i])[0].S.Job < (*sss[j])[0].S.Job
		} else {
			return (*sss[i])[0].S.Timestamp < (*sss[j])[0].S.Timestamp
		}
	} else {
		return (*sss[i])[0].S.Host.String() < (*sss[j])[0].S.Host.String()
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
	return (*sss[i])[0].S.Host.String() < (*sss[j])[0].S.Host.String()
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
	return (*sss[i])[0].S.Timestamp < (*sss[j])[0].S.Timestamp
}
