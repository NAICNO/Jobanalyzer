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
	nodeSamples []*repr.NodeSample,
	diskSamples []*repr.DiskSample,
	loadData []*repr.CpuSamples,
	gpuData []*repr.GpuSamples,
	softErrors int,
	err error,
) {
	samples = make([]*repr.Sample, 0)
	nodeSamples = make([]*repr.NodeSample, 0)
	diskSamples = make([]*repr.DiskSample, 0)
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
		bo, err := time.Parse(time.RFC3339, string(data.System.Boot))
		var boot int64
		if err != nil {
			boot = bo.Unix()
		}
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
		disks := data.System.Disks
		if len(disks) > 0 {
			for _, disk := range disks {
				stats := disk.Stats
				diskSamples = append(diskSamples, &repr.DiskSample{
					Timestamp:         timestamp,
					Hostname:          node,
					Name:              StringToUstr(disk.Name),
					Major:             disk.Major,
					Minor:             disk.Minor,
					ReadsCompleted:    maybe(stats, 0),
					ReadsMerged:       maybe(stats, 1),
					SectorsRead:       maybe(stats, 2),
					MsReading:         maybe(stats, 3),
					WritesCompleted:   maybe(stats, 4),
					WritesMerged:      maybe(stats, 5),
					SectorsWritten:    maybe(stats, 6),
					MsWriting:         maybe(stats, 7),
					IOsInProgress:     maybe(stats, 8),
					MsDoingIO:         maybe(stats, 9),
					WeightedMsDoingIO: maybe(stats, 10),
					DiscardsCompleted: maybe(stats, 11),
					DiscardsMerged:    maybe(stats, 12),
					SectorsDiscarded:  maybe(stats, 13),
					MsDiscarding:      maybe(stats, 14),
					FlushesCompleted:  maybe(stats, 15),
					MsFlushing:        maybe(stats, 16),
				})
			}
		}
		nodeSamples = append(nodeSamples, &repr.NodeSample{
			Timestamp:        timestamp,
			Boot:             boot,
			Hostname:         node,
			UsedMemory:       data.System.UsedMemory,
			Load1:            data.System.Load1,
			Load5:            data.System.Load5,
			Load15:           data.System.Load15,
			RunnableEntities: data.System.RunnableEntities,
			ExistingEntities: data.System.ExistingEntities,
		})
		if len(data.Jobs) == 0 {
			samples = append(samples, &repr.Sample{
				Timestamp: timestamp,
				Version:   version,
				Cluster:   cluster,
				Hostname:  node,
				Flags:     repr.FlagHeartbeat,
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
					Timestamp:         timestamp,
					MemtotalKB:        0,
					CpuKB:             process.VirtualMemory,
					RssAnonKB:         process.ResidentMemory,
					GpuKB:             gpuKib,
					CpuTimeSec:        process.CpuTime,
					Epoch:             job.Epoch,
					Version:           version,
					Cluster:           cluster,
					Hostname:          node,
					NumCores:          uint32(len(cpus)),
					NumThreads:        uint32(process.NumThreads) + 1,
					User:              user,
					Job:               uint32(job.Job),
					Pid:               process.Pid,
					Ppid:              uint32(process.ParentPid),
					Cmd:               ustrs.Alloc(process.Cmd),
					CpuPct:            float32(process.CpuAvg),
					Gpus:              pgpus,
					GpuPct:            float32(gpuPct),
					GpuMemPct:         float32(gpuMemPct),
					Rolledup:          uint32(process.Rolledup),
					GpuFail:           gpuFail,
					Flags:             0,
					InContainer:       process.InContainer,
					CpuSampledUtilPct: float32(process.CpuUtil),
					DataReadKB:        process.Read,
					DataWrittenKB:     process.Written,
					DataCancelledKB:   process.Cancelled,
				})
			}
		}
	})
	return
}

func maybe(xs []uint64, x int) uint64 {
	if x >= len(xs) {
		return 0
	}
	return xs[x]
}
