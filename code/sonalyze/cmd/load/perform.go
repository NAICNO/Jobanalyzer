package load

import (
	"fmt"
	"io"
	"math"
	"slices"
	"time"

	"go-utils/config"
	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/sonarlog"
	. "sonalyze/table"
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
	hostGlobber *Hosts,
	_ *db.SampleFilter,
) error {
	fromIncl, toIncl := lc.InterpretFromToWithBounds(bounds)

	if NeedsConfig(loadFormatters, lc.PrintFields) {
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
				probe := cfg.LookupHost((*stream)[0].Hostname.String())
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

	var queryNeg func(*ReportRecord) bool
	if lc.ParsedQuery != nil {
		var err error
		queryNeg, err = CompileQueryNeg(loadFormatters, loadPredicates, lc.ParsedQuery)
		if err != nil {
			return fmt.Errorf("Could not compile query: %v", err)
		}
	}

	// Generate data to be printed
	reports := make([]LoadReport, 0)
	for _, stream := range mergedStreams {
		hostname := (*stream)[0].Hostname.String()
		conf := mergedConf
		if conf == nil && cfg != nil {
			conf = cfg.LookupHost(hostname)
		}
		rs := generateReport(*stream, time.Now().Unix(), conf)
		if queryNeg != nil {
			rs = slices.DeleteFunc(rs, queryNeg)
		}
		reports = append(reports, LoadReport{
			hostname: hostname,
			records:  rs,
			conf:     conf,
		})
	}

	// And print it
	lc.printStreams(out, reports)

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
	host := (*ss)[0].Hostname
	t := trunc(fromIncl)
	result := make(sonarlog.SampleStream, 0)

	for _, s := range *ss {
		for t < s.Timestamp {
			newS := sonarlog.Sample{Sample: &repr.Sample{Timestamp: t, Hostname: host}}
			result = append(result, newS)
			t = step(t)
		}
		result = append(result, s)
		t = step(t)
	}
	ending := trunc(toIncl)
	for t <= ending {
		newS := sonarlog.Sample{Sample: &repr.Sample{Timestamp: t, Hostname: host}}
		result = append(result, newS)
		t = step(t)
	}
	*ss = result
}

// `sys` may be nil if none of the requested fields use its data, so we must guard against that.
func generateReport(
	input []sonarlog.Sample,
	now int64,
	sys *config.NodeConfigRecord,
) (result []*ReportRecord) {
	result = make([]*ReportRecord, 0, len(input))
	for _, d := range input {
		var relativeCpu, relativeVirtualMem, relativeResidentMem, relativeGpu, relativeGpuMem int
		if sys != nil {
			if sys.CpuCores > 0 {
				relativeCpu = int(math.Round(float64(d.CpuUtilPct) / float64(sys.CpuCores)))
			}
			if sys.MemGB > 0 {
				relativeVirtualMem =
					int(math.Round(float64(d.CpuKB) / (1024 * 1024) / float64(sys.MemGB) * 100.0))
				relativeResidentMem =
					int(math.Round(float64(d.RssAnonKB) / (1024 * 1024) / float64(sys.MemGB) * 100.0))
			}
			if sys.GpuCards > 0 {
				// GpuPct is already scaled by 100 so don't do it again
				relativeGpu = int(math.Round(float64(d.GpuPct) / float64(sys.GpuCards)))
			}
			if sys.GpuMemGB > 0 {
				relativeGpuMem =
					int(math.Round(float64(d.GpuKB) / (1024 * 1024) / float64(sys.GpuMemGB) * 100))
			}
		}
		result = append(result, &ReportRecord{
			Now:                 DateTimeValue(now),
			DateTime:            DateTimeValue(d.Timestamp),
			Date:                DateValue(d.Timestamp),
			Time:                TimeValue(d.Timestamp),
			Cpu:                 int(d.CpuUtilPct),
			RelativeCpu:         relativeCpu,
			VirtualGB:           int(d.CpuKB / (1024 * 1024)),
			RelativeVirtualMem:  relativeVirtualMem,
			ResidentGB:          int(d.RssAnonKB / (1024 * 1024)),
			RelativeResidentMem: relativeResidentMem,
			Gpu:                 int(d.GpuPct),
			RelativeGpu:         relativeGpu,
			GpuGB:               int(d.GpuKB / (1024 * 1024)),
			RelativeGpuMem:      relativeGpuMem,
			Gpus:                d.Gpus,
			Hostname:            d.Hostname,
		})
	}
	return
}
