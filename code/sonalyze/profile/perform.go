package profile

import (
	"errors"
	"io"
	"log"
	"math"
	"sort"

	"go-utils/config"
	"go-utils/hostglob"
	"go-utils/maps"
	. "go-utils/minmax"
	"sonalyze/sonarlog"
)

func (pc *ProfileCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	_ *sonarlog.LogDir,
	samples sonarlog.SampleStream,
	_ *hostglob.HostGlobber,
	recordFilter func(*sonarlog.Sample) bool,
) error {
	streams := sonarlog.PostprocessLog(samples, recordFilter, nil)
	jobId := pc.Job[0]

	if len(streams) == 0 {
		return errors.New("No processes")
	}

	// Simplify: Assert that the input has only a single host.
	// Precompute: check whether we need to print the `nproc` field.
	// TODO: IMPLEMENTME: Remove single-host restriction, it is too limiting.

	var hasRolledup bool
	var hostName string
	for k, vs := range streams {
		for _, v := range *vs {
			hasRolledup = hasRolledup || v.Rolledup > 0
		}
		if hostName != "" && k.Host.String() != hostName {
			return errors.New("`profile` only implemented for single-host jobs")
		}
		hostName = k.Host.String()
	}
	if hostName == "" {
		hostName = "unknown"
	}

	// The input is a matrix of per-process-per-point-in-time data, with time running down the
	// column, process index running across the row, and where each datum can have one or more
	// measurements of interest for that process at that time (cpu, mem, gpu, gpumem, nproc).  THE
	// MATRIX IS SPARSE, as processes only have data at points in time when they are running.
	//
	// We apply (optional) clamping to all the pertinent fields during the matrix conversion step to
	// make values sane, and build an explicit sparse matrix.

	// `processes` has the event streams for the processes (or group of rolled-up processes).
	//
	// We want these sorted in the order in which they start being shown, so that there is a natural
	// feel to the list of processes for each timestamp.  Sorting ascending by first timestamp will
	// accomplish that.
	processes := maps.Values(streams)
	sort.Stable(sonarlog.TimeSortableSampleStreams(processes))

	userName := (*processes[0])[0].User.String()

	// Number of nonempty streams remaining, this is the termination condition.
	nonempty := 0
	for _, p := range processes {
		if len(*p) > 0 {
			nonempty++
		}
	}

	// Indices into those streams of the next record we want.
	indices := make([]int, len(processes))

	initialNonempty := nonempty
	m := newProfData()
	timesteps := 0
	prevTime := int64(0)

	// Generate the initial matrix.
	//
	// This loop is quadratic-ish but `processes` will tend (modulo non-rolled-up MPI jobs, TBD) to
	// be very short and it's not clear what's to be gained yet by doing something more complicated
	// here like a priority queue, say.

	for nonempty > 0 {
		// The current time is the minimum time across the lists that are not exhausted.
		currentTime := int64(math.MaxInt64)
		for i, p := range processes {
			if indices[i] < len(*p) {
				currentTime = MinInt64(currentTime, (*p)[indices[i]].Timestamp)
			}
		}
		if currentTime == math.MaxInt64 {
			panic("currentTime")
		}
		if currentTime != prevTime {
			timesteps++
			prevTime = currentTime
		}
		for i, p := range processes {
			if indices[i] < len(*p) {
				r := (*p)[indices[i]]
				if r.Timestamp == currentTime {
					newr := pc.clampFields(r)
					m.set(currentTime, processId(newr), profDatum{currentTime, newr})
					indices[i]++
					if indices[i] == len(*p) {
						nonempty--
					}
				}
			}
		}
	}

	// Bucketing will average consecutive records in the clamped record stream running down a column
	// (within the same process).  We count only present entries in the divisor for the average.
	// The time value will be the midpoint in the chunk.

	if pc.Bucket > 1 {
		b := int(pc.Bucket)
		m2 := newProfData()
		// row names are timestamps
		rowNames := m.rows()
		colNames := m.cols()
		for r := 0; r < len(rowNames); r += b {
			myrowNames := rowNames[r:MinInt(r+b, len(rowNames))]
			newTime := myrowNames[len(myrowNames)/2]
			for _, cn := range colNames {
				var count int
				var cpuUtilPct, gpuPct float32
				var cpuKib, gpuKib, rssAnonKib uint64
				var avg *sonarlog.Sample
				for _, rn := range myrowNames {
					if probe := m.get(rn, cn); !probe.isEmpty() {
						proc := probe.s
						if avg == nil {
							var s sonarlog.Sample = *proc
							avg = &s
						}
						count++
						cpuUtilPct += proc.CpuUtilPct
						gpuPct += proc.GpuPct
						cpuKib += proc.CpuKib
						gpuKib += proc.GpuKib
						rssAnonKib += proc.RssAnonKib
					}
				}
				if avg != nil {
					avg.CpuUtilPct = cpuUtilPct / float32(count)
					avg.GpuPct = gpuPct / float32(count)
					avg.CpuKib = cpuKib / uint64(count)
					avg.GpuKib = gpuKib / uint64(count)
					avg.RssAnonKib = rssAnonKib / uint64(count)
					avg.Timestamp = newTime
					m2.set(newTime, cn, profDatum{newTime, avg})
				}
			}
		}
		m = m2
	}

	if pc.Verbose {
		log.Printf("Number of processes: %d", initialNonempty)
		log.Printf("Any rolled-up processes: %v", hasRolledup)
		log.Printf("Number of time steps: %d", timesteps)
	}

	pc.printProfile(out, uint32(jobId), hostName, userName, hasRolledup, m, processes)
	return nil
}

// TODO: IMPROVEME: Pids are not unique b/c rolled-up and merged pids are zero and there may be
// several of these.  The following is a hack to work around that so that we can use our sparse
// matrix abstraction, indexed by pid.  But a better solution would be for processes with pid=0 to
// have a synthesized unique pid, probably.  For a first cut, the merging algorithm could pick
// something from a context object (basically random but with the guarantee of uniqueness), but even
// better would be something stable.  The solution used here, using the Ustr for the command name,
// is not the best thing, because there can be multiple processes with pid=0 and the same command
// name, even within the same job - two distinct rolled-up subtrees of processes each with the same
// command name would be enough.

func processId(s *sonarlog.Sample) uint32 {
	// Rolled-up processes have pid=0
	if s.Pid != 0 {
		return s.Pid
	}
	// But in that case the Ustr value of the command should be unique enough
	return uint32(s.Cmd)
}

// Clamping is a hack but it works.
func (pc *ProfileCommand) clampFields(r *sonarlog.Sample) *sonarlog.Sample {
	// Always return a copy of the sample, I guess?
	var newr sonarlog.Sample = *r
	if pc.Max != 0 {
		// We print memory in GiB so -max should be expressed in GiB, but we use KiB internally.  Scale here.
		newr.CpuUtilPct = clampMaxF32(newr.CpuUtilPct, float32(pc.Max))
		newr.CpuKib = clampMaxU64(newr.CpuKib, uint64(pc.Max*1024*1024))
		newr.RssAnonKib = clampMaxU64(newr.RssAnonKib, uint64(pc.Max*1024*1024))
		newr.GpuPct = clampMaxF32(newr.GpuPct, float32(pc.Max))
		newr.GpuKib = clampMaxU64(newr.GpuKib, uint64(pc.Max*1024*1024))
	}
	return &newr
}

// Max clamping: If the value x is greater than the clamp then return the clamp c, except if x is
// more than 2c, in which case return 0 - the assumption is that it's a wild outlier / noise.

func clampMaxF32(x, c float32) float32 {
	if x > c {
		if x > 2*c {
			return 0
		}
		return c
	}
	return x
}

func clampMaxU64(x, c uint64) uint64 {
	if x > c {
		if x > 2*c {
			return 0
		}
		return c
	}
	return x
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Sparse matrix from [timestamp, process-id] to a datum for the cell

type profIndex struct {
	row int64
	col uint32
}

type profDatum struct {
	t int64 // timestamp - this is redundant (it's the row name) but useful?
	s *sonarlog.Sample
}

func (pd profDatum) isEmpty() bool {
	return pd.s == nil
}

type profData struct {
	// It's possible to do better than this in various ways but this is simple.
	hasRowIndex map[int64]bool
	rowNames    []int64
	rowDirty    bool
	hasColIndex map[uint32]bool
	colNames    []uint32
	colDirty    bool
	entries     map[profIndex]profDatum
}

func newProfData() *profData {
	return &profData{
		hasRowIndex: make(map[int64]bool),
		rowNames:    make([]int64, 0),
		rowDirty:    false,
		hasColIndex: make(map[uint32]bool),
		colNames:    make([]uint32, 0),
		colDirty:    false,
		entries:     make(map[profIndex]profDatum),
	}
}

func (pd *profData) get(y int64, x uint32) profDatum {
	return pd.entries[profIndex{y, x}]
}

func (pd *profData) set(y int64, x uint32, v profDatum) {
	if !pd.hasRowIndex[y] {
		pd.rowDirty = true
		pd.rowNames = append(pd.rowNames, y)
		pd.hasRowIndex[y] = true
	}
	if !pd.hasColIndex[x] {
		pd.colDirty = true
		pd.colNames = append(pd.colNames, x)
		pd.hasColIndex[x] = true
	}
	pd.entries[profIndex{y, x}] = v
}

func (pd *profData) rows() []int64 {
	if pd.rowDirty {
		sort.Sort(Int64Slice(pd.rowNames))
		pd.rowDirty = false
	}
	return pd.rowNames
}

func (pd *profData) cols() []uint32 {
	if pd.colDirty {
		sort.Sort(Uint32Slice(pd.colNames))
		pd.colDirty = false
	}
	return pd.colNames
}

type Int64Slice []int64

func (s Int64Slice) Len() int           { return len(s) }
func (s Int64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Int64Slice) Less(i, j int) bool { return s[i] < s[j] }

type Uint32Slice []uint32

func (s Uint32Slice) Len() int           { return len(s) }
func (s Uint32Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Uint32Slice) Less(i, j int) bool { return s[i] < s[j] }
