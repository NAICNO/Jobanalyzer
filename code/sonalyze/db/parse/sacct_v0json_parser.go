package parse

import (
	"io"
	"strings"
	"time"

	"github.com/NordicHPC/sonar/util/formats/newfmt"

	. "sonalyze/common"
	"sonalyze/db/repr"
)

func ParseSlurmV0JSON(
	input io.Reader,
	ustrs UstrAllocator,
	verbose bool,
) (
	records []*repr.SacctInfo,
	softErrors int,
	err error,
) {
	// Let's not go via the translation layer here, as the JSON data are much closer to what we
	// want.
	records = make([]*repr.SacctInfo, 0)
	var fakeSacct newfmt.SacctData
	var gputmp = make([]byte, 100)
	err = newfmt.ConsumeJSONJobs(input, false, func(r *newfmt.JobsEnvelope) {
		if r.Errors != nil {
			softErrors++
			return
		}
		t, _ := time.Parse(time.RFC3339, string(r.Data.Attributes.Time))
		timestamp := t.Unix()
		for i := range r.Data.Attributes.SlurmJobs {
			job := &r.Data.Attributes.SlurmJobs[i]
			sacct := job.Sacct
			if sacct == nil {
				sacct = &fakeSacct
			}
			var t time.Time
			t, _ = time.Parse(time.RFC3339, string(job.Start))
			startTime := t.Unix()
			t, _ = time.Parse(time.RFC3339, string(job.End))
			endTime := t.Unix()
			t, _ = time.Parse(time.RFC3339, string(job.SubmitTime))
			submitTime := t.Unix()
			step := StringToUstr(job.JobStep)
			arrayStep := UstrEmpty
			if job.ArrayJobID != 0 {
				arrayStep = step
			}
			hetStep := UstrEmpty
			if job.HetJobID != 0 {
				hetStep = step
			}
			timelimit := uint64(0)
			if job.Timelimit >= newfmt.ExtendedUintBase {
				timelimit, _ = job.Timelimit.ToUint()
			}
			priority := uint64(0)
			if job.Priority >= newfmt.ExtendedUintBase {
				priority, _ = job.Priority.ToUint()
			}
			// This is probably not quite the logic we want yet: it conflates requested and
			// allocated, and we probably want to avoid that.
			var allocGPUS Ustr
			allocGPUS, gputmp = ParseSlurmGPUResources([]byte(job.AllocTRES), ustrs, gputmp)
			if allocGPUS == UstrEmpty {
				allocGPUS, gputmp = ParseSlurmGPUResources([]byte(job.ReqTRES), ustrs, gputmp)
			}
			// In older data, go to the sacct auxiliary data.
			if allocGPUS == UstrEmpty {
				allocGPUS, gputmp = ParseSlurmGPUResources([]byte(sacct.AllocTRES), ustrs, gputmp)
			}
			var allocRes Ustr
			if job.AllocTRES != "" {
				allocRes = ustrs.Alloc(job.AllocTRES)
			} else if sacct.AllocTRES != "" {
				allocRes = ustrs.Alloc(sacct.AllocTRES)
			}
			var reqRes Ustr
			if job.ReqTRES != "" {
				reqRes = ustrs.Alloc(job.ReqTRES)
			}
			var nodes strings.Builder
			for _, v := range job.NodeList {
				if nodes.Len() > 0 {
					nodes.WriteRune(',')
				}
				nodes.WriteString(string(v))
			}
			records = append(records, &repr.SacctInfo{
				Time:         timestamp,
				Start:        startTime,
				End:          endTime,
				Submit:       submitTime,
				SystemCPU:    sacct.SystemCPU,
				UserCPU:      sacct.UserCPU,
				AveCPU:       sacct.AveCPU,
				MinCPU:       sacct.MinCPU,
				Version:      ustrs.Alloc(string(r.Meta.Version)),
				User:         ustrs.Alloc(job.UserName),
				JobName:      ustrs.Alloc(job.JobName),
				State:        ustrs.Alloc(string(job.JobState)),
				Account:      ustrs.Alloc(job.Account),
				Layout:       ustrs.Alloc(job.Layout),
				Reservation:  ustrs.Alloc(job.Reservation),
				JobStep:      step,
				ArrayStep:    arrayStep,
				HetStep:      hetStep,
				NodeList:     ustrs.Alloc(nodes.String()),
				Partition:    ustrs.Alloc(job.Partition),
				ReqGPUS:      allocGPUS, // "ReqGPUS" is misnamed
				AllocRes:     allocRes,
				ReqRes:       reqRes,
				JobID:        uint32(job.JobID),
				ArrayJobID:   uint32(job.ArrayJobID),
				ArrayTaskID:  uint32(job.ArrayTaskID),
				HetJobID:     uint32(job.HetJobID),
				HetJobOffset: uint32(job.HetJobOffset),
				AveDiskRead:  sacct.AveDiskRead,
				AveDiskWrite: sacct.AveDiskWrite,
				AveRSS:       sacct.AveRSS,
				AveVMSize:    sacct.AveVMSize,
				ElapsedRaw:   uint32(sacct.ElapsedRaw),
				MaxRSS:       sacct.MaxRSS,
				MaxVMSize:    sacct.MaxVMSize,
				ReqCPUS:      uint32(job.ReqCPUS),
				ReqMem:       job.ReqMemoryPerNode,
				ReqNodes:     uint32(job.ReqNodes),
				Suspended:    uint32(job.Suspended),
				TimelimitRaw: uint32(timelimit),
				ExitCode:     uint8(job.ExitCode),
				Priority:     priority,
			})
		}
	})
	return
}
