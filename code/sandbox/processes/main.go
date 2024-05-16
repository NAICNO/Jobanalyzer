// computeDeduction() will process the samples with valid `cputime` fields (representing
// self+children time) and will update the `deduction` field in each sample.  The true cputime for a
// sample is the difference cputime-deduction.  Deductions will be zero in all samples in all
// processes in all leaf jobs; in parent jobs, the deduction in a sample will represent the sum of
// the final cputime from subjobs underneath the process of that sample at the times those subjobs
// terminated.

type sample struct {
	// these are given
	time int64
	cputime uint64

	// this will be computed
	deduction uint64
}

type sampleStream struct {
	start, end int64
	pid, ppid, pgid int
	samples []sample
}

func main() {
	// All of these streams come from the same host.  There is one stream per process, and each
	// stream (process) belongs to a job.  Processes and jobs are (normally) nested hierarchically.
	var streams []*sampleStream = getInputStreams()
	computeDeductions(streams)
}

// computeDeductions()
//
// Sample streams are sorted in ascending time order and are local to a single host.  All
// observations made during the same Sonar run have the exact same timestamp in the stream.  Under
// normal circumstances a parent process will be present for every child process, and if we can't
// find the parent it's OK to assume the child has become detached somehow and can be considered its
// own root.  Ditto, jobs are normally nested hierarchically and no attempts are made to deal with
// other cases other than not to crash.

type proc struct {
	pid, pgid int
	start, end int64
	label uint64
	parent *proc
	children map[int]*proc
	samples *sampleStream
}

// The event queue is all about the `end` timestamp of a process, so a process A is *less than*
// another B for the purposes of an event stream if A ends at an earlier time than B or otherwise if
// A has a smaller label than B.  The label is used to fix parent-child relationships at records
// with the same timestamp and is applied when new processes are incorporated into the process tree
// and event queue.

type evQueue []*proc

func (eq evQueue) Len() int {
	return len(eq)
}

func (eq evQueue) Swap(i, j) {
	eq[i], eq[j] = eq[j], eq[i]
}

func (eq evQueue) Less(i, j int) bool {
	if eq[i].end == eq[j].end {
		return eq[i].label < eq[j].label
	}
	return eq[i].end < eq[j].end
}

func (eq *evQueue) Push(heap.Interface, x any) {
	*eq = append(*eq, x.(*proc))
}

func (eq *evQueue) Pop(heap.Interface) any {
	old := *eq
	l := len(old)
	item := old[l-1]
	old[l-1] = nil
	*eq = old[:l-1]
	return item
}

var tree = make(map[int]*proc)

var initproc = &proc{
	pid: 1,
	pgid: 1,
	start: 0,
	end: math.MaxInt64,
	label: 0
	parent: nil,
	children: make(map[int]*proc),
	input: nil,
}

func init() {
	tree[1] = initproc
}

var label uint64
var events = make(evQueue, 0)

func computeDeduction(streams []*sampleStream) {
	ix := 0
	lim := len(input)
	for ix < lim {
		start := input[ix].start

		for {
			if ev := pq.getFirstIfLess(start) {
				break
			}
			// TODO: Add to all future records in the event stream of the parent
			//
			// The finalCputime and finalDeductible must come from the last record of the event stream
			// of the process
			if ev.isRoot() {
				ev.parent.deductible += ev.finalCputime()
			} else {
				ev.parent.deductible += ev.finalDeductible()
			}
			delete(tree, ev.pid)
			// TODO: How to propagate this to the input??  AT THIS POINT IN TIME, and forward, all
			// data for ev.parent must be adjusted by its deductible, but this must be done just
			// once for that deductible, and it must be retained somehow.  The easy answer is that
			// there is a deductible field (also) per event record and that at the subjob's
			// deductible is accrued into those fields of the parent going into the future.  (Also
			// for subprocesses.)
		}

		// Collect the processes from the first time step of the input and place them in a
		// hierarchy, and record the roots.
		jx := ix
		procs := make(map[int]*proc, 0)
		for jx < lim && input[jx].start == input[ix].start {
			x := input[jx]
			procs[x.pid] = &proc{
				pid: x.pid,
				pgid: x.pgid,
				start: x.start,
				end: x.end,
				children: make(map[int]*proc),
				cputime: x.cputime,
				input: x,
			}
			jx++
		}

		unrooted := make([]*proc, 0)
		for _, p := range input[ix:jx] {
			probe := procs[p.input.ppid]
			if probe != nil {
				p.parent = probe
				probe.children[p.pid] = p
			} else {
				unrooted = append(unrooted, p)
			}
		}

		// Working postorder among the new processes, add the "end" events to the event queue in
		// sorted order: things that die sooner will end up earlier in the queue, but if a child and
		// parent die at the same time they child will sort before the parent.  It's possible the
		// heap code does not do this.  In that case, we want some disambiguating value.  We can
		// have a running counter and just working postorder we can label the children in this
		// arbitrary order, and the primary insertion key will be time and the secondary will be
		// this label.
		// The label could start at zero every time I guess
		for _, r := range unrooted {
			label = labelProcs(r, label)
			events = insertProc(r, events)
		}

		// Now the ones that were unrooted may actually be rooted in the existing process tree, so
		// we need to link them in.  If we can't find a parent, root them to the init process.
		//
		// FOO, this breaks the labeling: the parents will have lower labels than the children!
		// That's possibly OK, the labels are only used to disambiguate within the same time step.
		// But we must be careful.
		//
		// Also, after this, tree[p.input.ppid] will not always yield the same as p.parent, may wish
		// to fix that?
		for _, r := range unrooted {
			probe := tree[r.input.ppid]
			if probe == nil {
				probe = tree.init
			}
			r.parent = probe
			probe.children[r.pid] = r
		}

		ix = jx
	}
}

// Postorder labeling of the nodes
func labelProcs(p *proc, label uint64) uint64 {
	for _, c := range p.children {
		label = labelProcs(c, label)
	}
	label++
	p.label = label
	return label
}

func insertProc(p *proc, evs []*proc) *proc {
	for _, c := range p.children {
		evs = insertProc(c, evs)
	}
	evs = INSERT(p, evs)
	return evs
}

// Propagate a deduction to the parent process when the process dies.  The deduction is to account
// for the accrued (by Linux) cpu time coming from children in other jobs that have terminated;
// being from other jobs, we don't want to see it here.
//
// When a job root process dies its entire self+children time is added to the deduction in the
// parent process.  When *that* process dies, its deduction is added to its parent, too, and so it
// continues.  This is the vital difference: *within* a job, the deductions are summed and
// propagated, while *across* jobs, the deduction value propagated is the self+child time, as it is
// more accurate.
func (p *proc) propagateOnDeath() {
	if p.parent != nil {
		if p.isRoot() {
			p.parent.deduction += p.cputime
		} else {
			p.parent.deduction += p.deduction
		}
	}
}



func (p *proc) isRoot() bool {
	return p.parent == nil || p.parent.pgid != p.pgid
}

