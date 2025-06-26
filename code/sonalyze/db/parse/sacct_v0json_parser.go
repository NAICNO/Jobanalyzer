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
			var reqGPUS Ustr
			reqGPUS, gputmp = ParseAllocTRES([]byte(sacct.AllocTRES), ustrs, gputmp)
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
				NodeList:     ustrs.Alloc(strings.Join(job.NodeList, ",")),
				Partition:    ustrs.Alloc(job.Partition),
				ReqGPUS:      reqGPUS,
				JobID:        uint32(job.JobID),
				ArrayJobID:   uint32(job.ArrayJobID),
				ArrayIndex:   uint32(job.ArrayTaskID),
				HetJobID:     uint32(job.HetJobID),
				HetOffset:    uint32(job.HetJobOffset),
				AveDiskRead:  uint32(sacct.AveDiskRead),
				AveDiskWrite: uint32(sacct.AveDiskWrite),
				AveRSS:       uint32(sacct.AveRSS),
				AveVMSize:    uint32(sacct.AveVMSize),
				ElapsedRaw:   uint32(sacct.ElapsedRaw),
				MaxRSS:       uint32(sacct.MaxRSS),
				MaxVMSize:    uint32(sacct.MaxVMSize),
				ReqCPUS:      uint32(job.ReqCPUS),
				ReqMem:       uint32(job.ReqMemoryPerNode),
				ReqNodes:     uint32(job.ReqNodes),
				Suspended:    uint32(job.Suspended),
				TimelimitRaw: uint32(timelimit),
				ExitCode:     uint8(job.ExitCode),
				ExitSignal:   0, // not available
			})
		}
	})
	return
}
