// Parser for v0 "new format" JSON files holding Sonar `sample` data.

package parse

import (
	"io"
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
		if r.Errors != nil {
			softErrors++
			return
		}
		version := ustrs.Alloc(string(r.Meta.Version))
		data := r.Data.Attributes
		ti, err := time.Parse(time.RFC3339, string(data.Time))
		if err != nil {
			// Can't recover from a bad timestamp
			return
		}
		timestamp := ti.Unix()
		cluster := ustrs.Alloc(string(data.Cluster))
		node := ustrs.Alloc(string(data.Node))
		cpus := data.System.Cpus
		if len(cpus) > 0 {
			cpuLoad := make([]uint64, len(cpus))
			for i, n := range cpus {
				cpuLoad[i] = uint64(n)
			}
			loadData = append(loadData, &repr.CpuSamples{
				Timestamp: timestamp,
				Hostname:  node,
				Encoded:   repr.EncodedCpuSamplesFromValues(cpuLoad),
			})
		}
		// This is a map because we can't depend on `gpus` being populated in the same way as the
		// per-process gpu array.
		failing := make(map[uint64]bool)
		gpus := data.System.Gpus
		if len(gpus) > 0 {
			encodedGpuData := make([]repr.PerGpuSample, len(gpus))
			for i := range gpus {
				encodedGpuData[i].Attr =
					repr.GpuHasUuid | repr.GpuHasComputeMode | repr.GpuHasUtil | repr.GpuHasFailing
				encodedGpuData[i].SampleGpu = &gpus[i]
				failing[gpus[i].Index] = gpus[i].Failing != 0
			}
			gpuData = append(gpuData, &repr.GpuSamples{
				Timestamp: timestamp,
				Hostname:  node,
				Encoded:   repr.EncodedGpuSamplesFromValues(encodedGpuData),
			})
		}
		if len(data.Jobs) == 0 {
			samples = append(samples, &repr.Sample{
				Timestamp:  timestamp,
				Version:    version,
				Cluster:    cluster,
				Hostname:   node,
				Flags:      repr.FlagHeartbeat,
			})
		}
		for _, job := range data.Jobs {
			user := ustrs.Alloc(string(job.User))
			for _, process := range job.Processes {
				var pgpus gpuset.GpuSet
				var gpuPct float64
				var gpuMemPct float64
				var gpuKib uint64
				var gpuFail uint8
				for _, g := range process.Gpus {
					pgpus, _ = gpuset.Adjoin(pgpus, uint32(g.Index))
					gpuPct += g.GpuUtil
					gpuMemPct += g.GpuMemoryUtil
					gpuKib += g.GpuMemory
					if failing[g.Index] {
						gpuFail = 1
					}
				}
				samples = append(samples, &repr.Sample{
					Timestamp:  timestamp,
					MemtotalKB: 0,
					CpuKB:      process.VirtualMemory,
					RssAnonKB:  process.ResidentMemory,
					GpuKB:      gpuKib,
					CpuTimeSec: process.CpuTime,
					Epoch:      job.Epoch,
					Version:    version,
					Cluster:    cluster,
					Hostname:   node,
					Cores:      uint32(len(cpus)),
					User:       user,
					Job:        uint32(job.Job),
					Pid:        uint32(process.Pid),
					Ppid:       uint32(process.ParentPid),
					Cmd:        ustrs.Alloc(process.Cmd),
					CpuPct:     float32(process.CpuAvg),
					Gpus:       pgpus,
					GpuPct:     float32(gpuPct),
					GpuMemPct:  float32(gpuMemPct),
					Rolledup:   uint32(process.Rolledup),
					GpuFail:    gpuFail,
					Flags:      0,
				})
			}
		}
	})
	return
}
