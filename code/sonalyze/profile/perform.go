package profile

import (
	"cmp"
	"errors"
	"io"
	"math"
	"slices"
	"sort"

	"go-utils/config"
	"go-utils/hostglob"
	"go-utils/maps"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

func (pc *ProfileCommand) NeedsBounds() bool {
	return false
}

func (pc *ProfileCommand) Perform(
	out io.Writer,
	_ *config.ClusterConfig,
	_ db.SampleCluster,
	streams sonarlog.InputStreamSet,
	_ sonarlog.Timebounds,
	_ *hostglob.HostGlobber,
	_ *db.SampleFilter,
) error {
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
	// feel to the list of processes for each timestamp.  Sorting ascending by first timestamp, then
	// by command name and finally by PID will accomplish that as well as it is possible.  (There
	// are still going to be cases where two runs might print different data: see processId().)
	processes := maps.Values(streams)
	slices.SortStableFunc(processes, func(a, b *sonarlog.SampleStream) int {
		c := cmp.Compare((*a)[0].Timestamp, (*b)[0].Timestamp)
		if c == 0 {
			c = cmp.Compare((*a)[0].Cmd.String(), (*b)[0].Cmd.String())
			if c == 0 {
				c = cmp.Compare((*a)[0].Pid, (*b)[0].Pid)
			}
		}
		return c
	})

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
				currentTime = min(currentTime, (*p)[indices[i]].Timestamp)
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
					m.set(currentTime, processId(r), newProfDatum(r, pc.Max))
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
			myrowNames := rowNames[r:min(r+b, len(rowNames))]
			newTime := myrowNames[len(myrowNames)/2]
			for _, cn := range colNames {
				var count int
				var cpuUtilPct, gpuPct float32
				var cpuKib, gpuKib, rssAnonKib uint64
				var base *db.Sample
				for _, rn := range myrowNames {
					if probe := m.get(rn, cn); probe != nil {
						count++
						cpuUtilPct += probe.cpuUtilPct
						gpuPct += probe.gpuPct
						cpuKib += probe.cpuKib
						gpuKib += probe.gpuKib
						rssAnonKib += probe.rssAnonKib
						base = probe.s
					}
				}
				if count > 0 {
					avg := &profDatum{
						cpuUtilPct: cpuUtilPct / float32(count),
						gpuPct:     gpuPct / float32(count),
						cpuKib:     cpuKib / uint64(count),
						gpuKib:     gpuKib / uint64(count),
						rssAnonKib: rssAnonKib / uint64(count),
						s:          base,
					}
					m2.set(newTime, cn, avg)
				}
			}
		}
		m = m2
	}

	if pc.Verbose {
		Log.Infof("Number of processes: %d", initialNonempty)
		Log.Infof("Any rolled-up processes: %v", hasRolledup)
		Log.Infof("Number of time steps: %d", timesteps)
	}

	return pc.printProfile(out, uint32(jobId), hostName, userName, hasRolledup, m, processes)
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

func processId(s sonarlog.Sample) uint32 {
	// Rolled-up processes have pid=0
	if s.Pid != 0 {
		return s.Pid
	}
	// But in that case the Ustr value of the command should be unique enough
	return uint32(s.Cmd)
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
	cpuUtilPct float32
	gpuPct     float32
	cpuKib     uint64
	gpuKib     uint64
	rssAnonKib uint64
	s          *db.Sample
}

func newProfDatum(r sonarlog.Sample, max float64) *profDatum {
	var v profDatum
	v.cpuUtilPct = r.CpuUtilPct
	v.gpuPct = r.GpuPct
	v.cpuKib = r.CpuKib
	v.gpuKib = r.GpuKib
	v.rssAnonKib = r.RssAnonKib
	v.s = r.Sample

	if max != 0 {
		// Clamping is a hack but it works.
		// We print memory in GiB so -max should be expressed in GiB, but we use KiB internally.  Scale here.
		v.cpuUtilPct = clampMaxF32(v.cpuUtilPct, float32(max))
		v.cpuKib = clampMaxU64(v.cpuKib, uint64(max*1024*1024))
		v.rssAnonKib = clampMaxU64(v.rssAnonKib, uint64(max*1024*1024))
		v.gpuPct = clampMaxF32(v.gpuPct, float32(max))
		v.gpuKib = clampMaxU64(v.gpuKib, uint64(max*1024*1024))
	}

	return &v
}

type profData struct {
	// It's possible to do better than this in various ways but this is simple.
	hasRowIndex map[int64]bool
	rowNames    []int64
	rowDirty    bool
	hasColIndex map[uint32]bool
	colNames    []uint32
	colDirty    bool
	entries     map[profIndex]*profDatum
}

func newProfData() *profData {
	return &profData{
		hasRowIndex: make(map[int64]bool),
		rowNames:    make([]int64, 0),
		rowDirty:    false,
		hasColIndex: make(map[uint32]bool),
		colNames:    make([]uint32, 0),
		colDirty:    false,
		entries:     make(map[profIndex]*profDatum),
	}
}

func (pd *profData) get(y int64, x uint32) *profDatum {
	return pd.entries[profIndex{y, x}]
}

func (pd *profData) set(y int64, x uint32, v *profDatum) {
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
