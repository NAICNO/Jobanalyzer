// Parser for v0 "new format" JSON files holding Sonar `sample` data.

package parse

import (
	"io"
	"slices"
	"time"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"go-utils/gpuset"

	. "sonalyze/common"
	"sonalyze/db/repr"
)

// If an error is encountered we will return the records successfully parsed before the error along
// with an error, but there is no ability to skip erroneous records and continue going after an
// error has been encountered.

func ParseSamplesV0JSON(
	input io.Reader,
	ustrs UstrAllocator,
	verbose bool,
) (
	samples []*repr.Sample,
	loadData []*repr.CpuSamples,
	gpuData []*repr.GpuSamples,
	softErrors int,
	err error,
) {
	samples = make([]*repr.Sample, 0)
	loadData = make([]*repr.CpuSamples, 0)
	gpuData = make([]*repr.GpuSamples, 0)
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
			loadData = append(loadData, &repr.CpuSamples{
				Timestamp: t,
				Hostname:  h,
				Encoded:   repr.EncodedCpuSamplesFromValues(encodedLoadData),
			})
		}

		// Ignore the translated GPU data, use the original
		if r.Data.Attributes.System.Gpus != nil {
			gpus := r.Data.Attributes.System.Gpus
			encodedGpuData := make([]repr.PerGpuSample, len(data.GpuSamples))
			for i := range gpus {
				encodedGpuData[i].Attr =
					repr.GpuHasUuid | repr.GpuHasComputeMode | repr.GpuHasUtil | repr.GpuHasFailing
				encodedGpuData[i].SampleGpu = &gpus[i]
			}
			gpuData = append(gpuData, &repr.GpuSamples{
				Timestamp: t,
				Hostname:  h,
				Encoded:   repr.EncodedGpuSamplesFromValues(encodedGpuData),
			})
		}

		for _, sample := range data.Samples {
			gpus, _ := gpuset.NewGpuSet(sample.Gpus)
			samples = append(samples, &repr.Sample{
				Timestamp:  t,
				MemtotalKB: data.MemtotalKib,
				CpuKB:      sample.CpuKib,
				RssAnonKB:  sample.RssAnonKib,
				GpuKB:      sample.GpuKib,
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
