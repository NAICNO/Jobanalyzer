package gpusample

import (
	"cmp"
	"maps"
	"slices"
	"time"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
)

// Ditto for GPU data

type GpuSamplesByHost struct {
	Hostname Ustr
	Data     []GpuSamples // one per timestamp
}

type PerGpuSample = repr.PerGpuSample

type GpuSamples struct {
	Time    int64
	Decoded []PerGpuSample // one per gpu at the given time
}

type GpuSamplesByHostSet map[Ustr]*GpuSamplesByHost

func ReadGpuSamplesByHost(
	c db.GpuSampleDataProvider,
	fromDate, toDate time.Time,
	hostGlobber *Hosts,
	verbose bool,
) (
	streams GpuSamplesByHostSet,
	bounds Timebounds,
	read, dropped int,
	err error,
) {
	// Read and establish invariants

	dataBlobs, dropped, err := c.ReadGpuSamples(fromDate, toDate, hostGlobber, verbose)
	if err != nil {
		return
	}
	for _, data := range dataBlobs {
		read += len(data)
	}
	streams, bounds, errors := rectifyGpuSamplesByHost(dataBlobs)
	dropped += errors
	return
}

func rectifyGpuSamplesByHost(dataBlobs [][]*repr.GpuSamples) (streams GpuSamplesByHostSet, bounds Timebounds, errors int) {
	// Divide data among the hosts and decode each array
	streams = make(GpuSamplesByHostSet)
	for _, data := range dataBlobs {
		for _, d := range data {
			decoded, err := repr.DecodeEncodedGpuSamples(d.Encoded)
			if err != nil {
				errors++
				continue
			}
			datum := GpuSamples{
				Time:    d.Timestamp,
				Decoded: decoded,
			}
			if stream, found := streams[d.Hostname]; found {
				stream.Data = append(stream.Data, datum)
			} else {
				stream := GpuSamplesByHost{
					Hostname: d.Hostname,
					Data:     []GpuSamples{datum},
				}
				streams[d.Hostname] = &stream
			}
		}
	}

	// Sort each per-host list by ascending time
	for _, v := range streams {
		slices.SortFunc(v.Data, func(a, b GpuSamples) int {
			return cmp.Compare(a.Time, b.Time)
		})
	}

	// Compute bounds
	bounds = make(Timebounds)
	for k, v := range streams {
		// By construction, v.Data is never empty here
		bounds[k] = Timebound{Earliest: v.Data[0].Time, Latest: v.Data[len(v.Data)-1].Time}
	}

	// Remove duplicates in each per-host list
	for _, v := range streams {
		v.Data = slices.CompactFunc(v.Data, func(a, b GpuSamples) bool {
			return a.Time == b.Time
		})
	}

	// Remove completely empty streams
	maps.DeleteFunc(streams, func(k Ustr, v *GpuSamplesByHost) bool {
		return len(v.Data) == 0
	})

	return
}
