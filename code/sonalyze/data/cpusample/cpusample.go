package cpusample

import (
	"maps"
	"slices"
	"time"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

// Per-cpu load data, expanded.

type CpuSamplesByHost struct {
	Hostname Ustr
	Data     []CpuSamples
}

type CpuSamples struct {
	Time    int64    // seconds since epoch UTC
	Decoded []uint64 // since-boot cpu time values for cores in core order
}

type CpuSampleDataProvider struct {
	theLog db.CpuSampleDataProvider
}

func OpenCpuSampleDataProvider(meta special.ClusterMeta) (*CpuSampleDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, special.CpuSampleData)
	if err != nil {
		return nil, err
	}
	return &CpuSampleDataProvider{theLog}, nil
}

func CpuSamplesLessByTime(a, b CpuSamples) int {
	if a.Time < b.Time {
		return -1
	}
	if a.Time > b.Time {
		return 1
	}
	return 0
}

// The table key is the same value as the value's Host member.
//
// After postprocessing, some invariants:
//
// - the CpuSamplesByHost.Data vectors are sorted ascending by time
// - no two adjacent timestamps are the same
//
// TODO: Not sure yet whether this really needs to be a map.

type CpuSampleSet map[Ustr]*CpuSamplesByHost

func (cdp *CpuSampleDataProvider) Query(
	fromDate, toDate time.Time,
	hostGlobber *Hosts,
	verbose bool,
) (
	streams CpuSampleSet,
	bounds Timebounds,
	read, dropped int,
	err error,
) {
	// Read and establish invariants

	dataBlobs, dropped, err := cdp.theLog.ReadCpuSamples(fromDate, toDate, hostGlobber, verbose)
	if err != nil {
		return
	}
	for _, data := range dataBlobs {
		read += len(data)
	}
	streams, bounds, errors := rectifyCpuSamples(dataBlobs)
	dropped += errors
	return
}

func rectifyCpuSamples(dataBlobs [][]*repr.CpuSamples) (streams CpuSampleSet, bounds Timebounds, errors int) {
	// Divide data among the hosts and decode each array
	streams = make(CpuSampleSet)
	for _, data := range dataBlobs {
		for _, d := range data {
			decoded, err := repr.DecodeEncodedCpuSamples(d.Encoded)
			if err != nil {
				errors++
				continue
			}
			datum := CpuSamples{
				Time:    d.Timestamp,
				Decoded: decoded,
			}
			if stream, found := streams[d.Hostname]; found {
				stream.Data = append(stream.Data, datum)
			} else {
				stream := CpuSamplesByHost{
					Hostname: d.Hostname,
					Data:     []CpuSamples{datum},
				}
				streams[d.Hostname] = &stream
			}
		}
	}

	// Sort each per-host list by ascending time
	for _, v := range streams {
		slices.SortFunc(v.Data, CpuSamplesLessByTime)
	}

	// Compute bounds
	bounds = make(Timebounds)
	for k, v := range streams {
		// By construction, v.Data is never empty here
		bounds[k] = Timebound{Earliest: v.Data[0].Time, Latest: v.Data[len(v.Data)-1].Time}
	}

	// Remove duplicates in each per-host list
	for _, v := range streams {
		v.Data = slices.CompactFunc(v.Data, func(a, b CpuSamples) bool {
			return a.Time == b.Time
		})
	}

	// Remove completely empty streams
	maps.DeleteFunc(streams, func(k Ustr, v *CpuSamplesByHost) bool {
		return len(v.Data) == 0
	})

	return
}
