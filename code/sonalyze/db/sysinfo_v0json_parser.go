// Parser for v0 "new format" JSON files holding Sonar `sysinfo` data.

package db

import (
	"io"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"sonalyze/db/repr"
)

// If an error is encountered we will return the records successfully parsed before the error along
// with an error, but there is no ability to skip erroneous records and continue going after an
// error has been encountered.
//
// NOTE: This is a really lossy translation and it needs to be replaced.  First, the old
// SysinfoEnvelope can't hold all the data that come from the new Sysinfo data.  Then, the
// NodeConfigRecord can hold even less.  What we need to do here is upgrade the pipeline in
// Jobanalyzer so that we no longer use NodeConfigRecord, ideally we move to the new SysinfoEnvelope
// but minimally to the old one.  Of course, then we also have to use the data.

func ParseSysinfoV0JSON(input io.Reader, verbose bool) (records []*repr.SysinfoData, err error) {
	records = make([]*repr.SysinfoData, 0)
	err = newfmt.ConsumeJSONSysinfo(input, false, func(r *newfmt.SysinfoEnvelope) {
		data, errdata := newfmt.NewSysinfoToOld(r)
		if errdata != nil {
			// Ignore error data
			return
		}

		// CrossNodeJobs is very complicated in reality and doesn't really make sense here: it
		// depends on whether the node in question was included in a slurm configuration on a
		// slurm cluster at some particular point in time.  As for GpuMemPCT, I'm not sure
		// what's happening with that, but that's mostly for older data predating the current
		// GPU API, so probably OK.
		records = append(records, &repr.SysinfoData{
			Timestamp:     data.Timestamp,
			Hostname:      data.Hostname,
			Description:   data.Description,
			CrossNodeJobs: false,
			CpuCores:      int(data.CpuCores),
			MemGB:         int(data.MemGB),
			GpuCards:      int(data.GpuCards),
			GpuMemGB:      int(data.GpuMemGB),
			GpuMemPct:     false,
		})
	})
	return
}
