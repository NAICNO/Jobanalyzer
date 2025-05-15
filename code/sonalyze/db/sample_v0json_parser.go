// Parser for v0 "new format" JSON files holding Sonar `sample` data.

package db

import (
	"io"
	"slices"
	"time"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"go-utils/gpuset"

	. "sonalyze/common"
)

// If an error is encountered we will return the records successfully parsed before the error along
// with an error, but there is no ability to skip erroneous records and continue going after an
// error has been encountered.

func ParseSamplesV0JSON(
	input io.Reader,
	ustrs UstrAllocator,
	verbose bool,
) (
	samples []*Sample,
	loadData []*LoadDatum,
	gpuData []*GpuDatum,
	softErrors int,
	err error,
) {
	samples = make([]*Sample, 0)
	loadData = make([]*LoadDatum, 0)
	gpuData = make([]*GpuDatum, 0)
	err = newfmt.ConsumeJSONSamples(input, false, func(r *newfmt.SampleEnvelope) {
		data, errdata := newfmt.NewSampleToOld(r)
		if errdata != nil {
			softErrors++
			return
		}

		ti, err := time.Parse(time.RFC3339, data.Timestamp)
		if err != nil {
			// Can't recover from a bad timestamp
			return
		}
		t := ti.Unix()
		h := ustrs.Alloc(data.Hostname)
		if data.CpuLoad != nil {
			encodedLoadData := slices.Clone(data.CpuLoad)
			loadData = append(loadData, &LoadDatum{
				Timestamp: t,
				Hostname: h,
				Encoded: EncodedLoadDataFromValues(encodedLoadData),
			})
		}

		if data.GpuSamples != nil {
			encodedGpuData := make([]PerGpuSample, len(data.GpuSamples))
			for i := range data.GpuSamples {
				s := &encodedGpuData[i]
				o := &data.GpuSamples[i]
				s.FanPct = int(o.FanPct)
				s.PerfMode = 0	// Nobody cares, right now
				s.MemUsedKB = int64(o.MemUse)
				s.TempC = int(o.Temp)
				s.PowerDrawW = int(o.Power)
				s.PowerLimitW = int(o.PowerLimit)
				s.CeClockMHz = int(o.CEClock)
				s.MemClockMHz = int(o.MemClock)
			}
			gpuData = append(gpuData, &GpuDatum{
				Timestamp: t,
				Hostname: h,
				Encoded: EncodedGpuDataFromValues(encodedGpuData),
			})
		}

		for _, sample := range data.Samples {
			gpus, _ := gpuset.NewGpuSet(sample.Gpus)
			samples = append(samples, &Sample{
				Timestamp: t,
				MemtotalKB: data.MemtotalKib,
				CpuKB: sample.CpuKib,
				RssAnonKB: sample.RssAnonKib,
				GpuKB: sample.GpuKib,
				CpuTimeSec: sample.CpuTimeSec,
				Version:    ustrs.Alloc(data.Version),
				Cluster:    ustrs.Alloc(string(r.Data.Attributes.Cluster)),
				Hostname:   ustrs.Alloc(data.Hostname),
				Cores:      uint32(data.Cores),
				User:       ustrs.Alloc(sample.User),
				Job:        uint32(sample.JobId),
				Pid:        uint32(sample.Pid),
				Ppid:       uint32(sample.ParentPid),
				Cmd:        ustrs.Alloc(sample.Cmd),
				CpuPct:     float32(sample.CpuPct),
				Gpus:       gpus,
				GpuPct:     float32(sample.GpuPct),
				GpuMemPct:  float32(sample.GpuMemPct),
				Rolledup:   uint32(sample.Rolledup),
				Flags:      uint8(sample.GpuFail),
			})
		}
	})
	return
}
