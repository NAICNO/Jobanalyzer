package load

import (
	"io"

	"go-utils/config"
	"go-utils/hostglob"
	. "sonalyze/command"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

func (lc *LoadCommand) NeedsBounds() bool {
	return true
}

func (lc *LoadCommand) Perform(
	out io.Writer,
	cfg *config.ClusterConfig,
	_ db.SampleCluster,
	streams sonarlog.InputStreamSet,
	bounds sonarlog.Timebounds,
	hostGlobber *hostglob.HostGlobber,
	_ *db.SampleFilter,
) error {
	fromIncl, toIncl := lc.InterpretFromToWithBounds(bounds)

	if lc.printRequiresConfig() {
		var err error
		streams, err = EnsureConfigForInputStreams(cfg, streams, "relative format arguments")
		if err != nil {
			return err
		}
	}

	// There one synthesized sample stream per host.  The samples will all have different
	// timestamps, and each stream will be sorted ascending by timestamp.
	mergedStreams := sonarlog.MergeByHost(streams)

	// Bucket the data, if applicable
	if lc.bucketing != bNone {
		newStreams := make(sonarlog.SampleStreams, 0)
		for _, s := range mergedStreams {
			var newS sonarlog.SampleStream
			switch lc.bucketing {
			case bHalfHourly:
				newS = sonarlog.FoldSamplesHalfHourly(*s)
			case bHourly:
				newS = sonarlog.FoldSamplesHourly(*s)
			case bHalfDaily:
				newS = sonarlog.FoldSamplesHalfDaily(*s)
			case bDaily:
				newS = sonarlog.FoldSamplesDaily(*s)
			case bWeekly:
				newS = sonarlog.FoldSamplesWeekly(*s)
			default:
				panic("Unexpected case")
			}
			newStreams = append(newStreams, &newS)
		}
		mergedStreams = newStreams
	}

	// If grouping, merge the streams across hosts and compute a system config that represents the
	// sum of the hosts in the group.
	var theConf config.NodeConfigRecord
	var mergedConf *config.NodeConfigRecord
	if lc.Group {
		if cfg != nil {
			for _, stream := range mergedStreams {
				// probe is non-nil by previous construction
				probe := cfg.LookupHost((*stream)[0].S.Host.String())
				if theConf.Description != "" {
					theConf.Description += "|||" // JSON-compatible separator
				}
				theConf.Description += probe.Description
				theConf.CpuCores += probe.CpuCores
				theConf.MemGB += probe.MemGB
				theConf.GpuCards += probe.GpuCards
				theConf.GpuMemGB += probe.GpuMemGB
			}
			mergedConf = &theConf
		}
		mergedStreams = sonarlog.MergeAcrossHostsByTime(mergedStreams)
		if len(mergedStreams) > 1 {
			panic("Too many results")
		}
	}

	// If not printing compactly then insert missing record in the streams
	if !lc.Compact && lc.All && lc.bucketing != bNone {
		for _, stream := range mergedStreams {
			// stream is a *SampleStream and is updated in-place
			lc.insertMissingRecords(stream, fromIncl, toIncl)
		}
	}

	lc.printStreams(out, cfg, mergedConf, mergedStreams)

	return nil
}

func (lc *LoadCommand) insertMissingRecords(ss *sonarlog.SampleStream, fromIncl, toIncl int64) {
	var trunc func(int64) int64
	var step func(int64) int64
	switch lc.bucketing {
	case bHalfHourly:
		trunc = TruncateToHalfHour
		step = AddHalfHour
	case bHourly:
		trunc = TruncateToHour
		step = AddHour
	case bHalfDaily, bNone:
		trunc = TruncateToHalfDay
		step = AddHalfDay
	case bDaily:
		trunc = TruncateToDay
		step = AddDay
	case bWeekly:
		trunc = TruncateToWeek
		step = AddWeek
	default:
		panic("Unexpected case")
	}
	host := (*ss)[0].S.Host
	t := trunc(fromIncl)
	result := make(sonarlog.SampleStream, 0)

	for _, s := range *ss {
		for t < s.S.Timestamp {
			newS := sonarlog.Sample{S: &db.Sample{Timestamp: t, Host: host}}
			result = append(result, newS)
			t = step(t)
		}
		result = append(result, s)
		t = step(t)
	}
	ending := trunc(toIncl)
	for t <= ending {
		newS := sonarlog.Sample{S: &db.Sample{Timestamp: t, Host: host}}
		result = append(result, newS)
		t = step(t)
	}
	*ss = result
}
