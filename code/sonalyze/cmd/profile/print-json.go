// We print an array of objects, each representing a "job" at a "time" (these are
// fields in each object).  In each object, there is a field "points" which is an array of data
// points.  Each data point has the value for all the data fields (regardless of what was requested)
// for that job at that time.

package profile

import (
	"encoding/json"
	"io"
	"math"
	"time"

	"sonalyze/data/sample"
)

func (pc *ProfileCommand) printJson(
	out io.Writer,
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
) error {
	formatJson(out, m, processes, pif, pc.testNoMemory)
	return nil
}

// TODO: Canonical names too?

func formatJson(
	out io.Writer,
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
	noMemory bool,
) {
	objects := ComputeJSONFromSamples(m, processes, pif, noMemory)
	e := json.NewEncoder(out)
	e.SetEscapeHTML(false)
	err := e.Encode(objects)
	if err != nil {
		panic("JSON encoding")
	}
}

type JsonPoint struct {
	Command    string `json:"command"`
	Host       string `json:"host,omitempty"`
	Pid        uint32 `json:"pid"`
	CpuUtilPct int    `json:"cpu"`
	CpuGB      uint64 `json:"mem"`
	RssAnonGB  uint64 `json:"res"`
	GpuPct     int    `json:"gpu"`
	GpuMemGB   uint64 `json:"gpumem"`
	Nproc      int    `json:"nproc"`
}

type JsonTimestep struct {
	Timestamp time.Time   `json:"-"`
	Time      string      `json:"time"` // TODO: Is this right?
	Job       uint32      `json:"job"`
	Points    []JsonPoint `json:"points"`
}

func ComputeJSONFromSamples(
	m *profData,
	processes []sample.SampleStream,
	pif *processIndexFactory,
	noMemory bool,
) []JsonTimestep {
	objects := make([]JsonTimestep, 0)
	for _, rn := range m.rows() {
		points := make([]JsonPoint, 0)
		var e *profDatum
		for _, p := range processes {
			cn := pif.indexFor(p[0])
			entry := m.get(rn, cn)
			if entry == nil {
				continue
			}
			if e == nil {
				e = entry
			}
			var cpuGB, gpuGB uint64
			if !noMemory {
				cpuGB = entry.cpuKB / (1024 * 1024)
				gpuGB = entry.gpuKB / (1024 * 1024)
			}
			hn, pid := pif.hostAndPid(cn)
			var hostname string
			if pif.isMultiHost() {
				hostname = hn.String()
			}
			points = append(points, JsonPoint{
				Command:    entry.s.Cmd.String(),
				Host:       hostname,
				Pid:        pid,
				CpuUtilPct: int(math.Round(float64(entry.cpuUtilPct))),
				CpuGB:      cpuGB,
				RssAnonGB:  uint64(math.Round(float64(entry.rssAnonKB) / (1024 * 1024))),
				GpuPct:     int(math.Round(float64(entry.gpuPct))),
				GpuMemGB:   gpuGB,
				Nproc:      int(entry.s.Rolledup) + 1,
			})
		}
		objects = append(objects, JsonTimestep{
			Timestamp: time.Unix(rn, 0),
			Time:      formatTime(rn),
			Job:       e.s.Job,
			Points:    points,
		})
	}
	return objects
}
