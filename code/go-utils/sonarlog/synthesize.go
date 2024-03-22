package sonarlog

import (
	"math"
)

const (
	farFuture int64 = math.MaxInt64
)

// Merge streams that have the same host and job ID into synthesized data.
//
// Each output stream is sorted ascending by timestamp.  No two records have exactly the same time.
// All records within a stream have the same host, command, user, and job ID.
//
// The command name for synthesized data collects all the commands that went into the synthesized
// stream.

func MergeByHostAndJob(streams InputStreamSet) SampleStreams {
	type HostAndJob struct {
		host Ustr
		job  uint32
	}

	type CommandsAndSampleStreams struct {
		commands map[Ustr]bool
		streams  SampleStreams
	}

	// The value is a set of command names and a vector of the individual streams.
	collections := make(map[HostAndJob]*CommandsAndSampleStreams)

	// The value is a map (by host) of the individual streams with job ID zero, these can't be
	// merged and must just be passed on.
	zero := make(map[Ustr]*SampleStreams)

	// Partition into jobs with ID and jobs without, and group the former by host and job and
	// collect information about each bag.
	for key, stream := range streams {
		id := (*stream)[0].Job
		if id == 0 {
			if bag, found := zero[key.Host]; found {
				*bag = append(*bag, stream)
			} else {
				zero[key.Host] = &SampleStreams{stream}
			}
		} else {
			k := HostAndJob{key.Host, id}
			if box, found := collections[k]; found {
				box.commands[key.Cmd] = true
				box.streams = append(box.streams, stream)
			} else {
				collections[k] = &CommandsAndSampleStreams{
					commands: map[Ustr]bool{key.Cmd: true},
					streams:  SampleStreams{stream},
				}
			}
		}
	}

	merged := make(SampleStreams, 0)

	for key, cmdsAndStreams := range collections {
		if zeroes, found := zero[key.host]; found {
			delete(zero, key.host)
			merged = append(merged, *zeroes...)
		}
		commands := make([]Ustr, 0, len(cmdsAndStreams.commands))
		for c := range cmdsAndStreams.commands {
			commands = append(commands, c)
		}
		UstrSortAscending(commands)
        // Any user from any record is fine.  There should be an invariant that no stream is empty,
        // so this should always be safe.
		username := (*cmdsAndStreams.streams[0])[0].User
		merged = append(merged, mergeStreams(
			key.host,
			UstrJoin(commands, StringToUstr(",")),
			username,
			key.job,
			cmdsAndStreams.streams,
		))
	}

	return merged
}

// FIXME: Lots of comments
func mergeStreams(
	hostname Ustr,
	command Ustr,
	username Ustr,
	jobId uint32,
	streams SampleStreams,
) *SampleStream {
	// FIXME: Lots of comments

	records := make(SampleStream, 0)

	v000 := StringToUstr("0.0.0")
	indices := make([]int, len(streams))
	const StreamEnded = math.MaxInt
	selected := make([]*Sample, 0, len(streams))

	live := 0
	sentinelTime := farFuture
	for {
		for live < len(streams) && indices[live] == StreamEnded {
			live++
		}

		minTime := sentinelTime
		for i := live; i < len(streams); i++ {
			if indices[i] >= len(*streams[i]) {
				continue
			}
			if minTime > (*streams[i])[indices[i]].Timestamp {
				minTime = (*streams[i])[indices[i]].Timestamp
			}
		}

		if minTime == sentinelTime {
			break
		}

		limTime := minTime + 10
		nearPast := minTime - 30
		deepPast := minTime - 60

		for i := live; i < len(streams); i++ {
			s := *streams[i]
			ix := indices[i]
			lim := len(s)

			if ix < lim {
				if s[ix].Timestamp >= limTime {
					continue
				}

				if s[ix].Timestamp == minTime {
					selected = append(selected, s[ix])
					indices[i]++
					continue
				}

				if s[ix].Timestamp > minTime {
					selected = append(selected, s[ix])
					indices[i]++
					continue
				}

				if ix > 0 && s[ix-1].Timestamp >= nearPast {
					selected = append(selected, s[ix-1])
					continue
				}

				if ix > 0 && s[ix-1].Timestamp >= deepPast {
					selected = append(selected, s[ix-1])
					continue
				}

				continue
			} else if ix == StreamEnded {
				continue
			} else {
				if s[ix-1].Timestamp < deepPast {
					indices[i] = StreamEnded
					continue
				}

				if s[ix-1].Timestamp < minTime {
					selected = append(selected, s[ix-1])
					continue
				}

				continue
			}
		}

		records = append(records, sumRecords(
			v000,
			minTime,
			hostname,
			username,
			jobId,
			command,
			selected,
		))
		selected = selected[0:0]
	}

	return &records
}

func MergeGpuFail(a, b uint8) uint8 {
	if a > 0 || b > 0 {
		return 1
	}
	return 0
}

func sumRecords(
	version Ustr,
	timestamp int64,
	hostname Ustr,
	username Ustr,
	jobId uint32,
	command Ustr,
	selected []*Sample,
) *Sample {
	var cpuPct, gpuPct, gpuMemPct, cpuUtilPct float32
	var cpuKib, rssAnonKib, gpuKib, cpuTimeSec uint64
	var rolledup uint32
	var gpuFail uint8
	var gpus = EmptyGpuSet()
	for _, s := range selected {
		cpuPct += s.CpuPct
		gpuPct += s.GpuPct
		gpuMemPct += s.GpuMemPct
		cpuUtilPct += s.CpuUtilPct
		cpuKib += s.CpuKib
		rssAnonKib += s.RssAnonKib
		gpuKib += s.GpuKib
		cpuTimeSec += s.CpuTimeSec
		rolledup += s.Rolledup
		gpuFail = MergeGpuFail(gpuFail, s.GpuFail)
		gpus = UnionGpuSets(gpus, s.Gpus)
	}
	// The invariant is that rolledup is the number of *other* processes rolled up into this one.
	// So we add one for each in the list + the others rolled into each of those, and subtract one
	// at the end to maintain the invariant.
	rolledup -= uint32(len(selected) + 1)

	// Synthesize the record.
	return &Sample{
		Version:    version,
		Timestamp:  timestamp,
		Host:       hostname,
		User:       username,
		Job:        jobId,
		Cmd:        command,
		CpuPct:     cpuPct,
		CpuKib:     cpuKib,
		RssAnonKib: rssAnonKib,
		Gpus:       gpus,
		GpuPct:     gpuPct,
		GpuMemPct:  gpuMemPct,
		GpuKib:     gpuKib,
		GpuFail:    gpuFail,
		CpuTimeSec: cpuTimeSec,
		Rolledup:   rolledup,
		CpuUtilPct: cpuUtilPct,
	}
}
