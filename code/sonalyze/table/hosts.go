// `Hostnames` is a specialized set of host names: if a.b.c is entered into the set then a and a.b
// will also be found to be in the set, but only one element is printed (and if a or a.b is later
// added to the set this is effectively a no-op).  Add order is invariant: if a.b is added to the
// set and then a.b.c then the set will still have a, a.b, and a.b.c.
//
// This serves the set operations for the query engine, eg, Hosts >= /gpu-[1-3]/ will find jobs that
// use at least gpu-1.fox, gpu-2.fox and gpu-3.fox; Hosts <= /gpu-[1-3]/ will find jobs that use no
// other nodes than those three, and may use just one of them.
//
// When printing to the "brief" form, only "a" is printed even if a.b.c (and thus also a.b) was in
// the set.
//
// When printing to the "full" form, only "a.b.c" is printed, the shorter variants are ignored.
//
// Thus for every name in the set there is a "brief" form (first element) and a "full" form (all
// elements).
//
// Set equality: A = B if every member of B is a (possibly improper) prefix of some member of A and
// there are no members of A that are not matched in this way.  Note A = B does not imply B = A.
//
// Subset: B < A if every member of B is a (possibly improper) prefix of some member of A.
//
// It's easiest to consider the sets as N-way trees.  If A = {a.b.c.d,a.b.c.e} then we see it as the
// tree A={a => {b => {c => {d => {}, e => {}}}}}.  If B = {a} then we have B={a => {}}.
//
// Intuitively, A = B if B covers A in such a way that there are no parts of B outside of A and all
// nodes of A outside of B are reached from nodes that are fully overlapped by some node of B, while
// for B < A the latter condition is removed:
//
// - To determine if B < A, walk A and B in parallel.  At every level we descend to, every member of
//   B also has to be in A.  We don't descend into levels for empty sets.
//
// - To determine if A = B, then in addition at every level we descend to have to have equal keys.
//
// Consider A={a.x, a.c, b.c} and B={a, b}, which is to say
//
//    A = {a => {x => {}, c => {}}, b => {c => {}}}
//    B = {a => {}, b => {}}
//
// Here the first level of A and the first level of B are the same {a,b} so we continue.  None
// of the values in B have nonempty sets so we're done.
//
// More complicated, consider B = {a, b.c}:
//
//    A = {a => {x => {}, c => {}}, b => {c => {}}}
//    B = {a => {}, b => {c => {}}}
//
// The first levels are the same, we descend to the second level for "b", the sets are the same {c}
// and the values are empty and we're done: they're equal, as desired.
//
//
// Under typical circumstances (queries):
//
//  - the lhs set will be created once for each row in the table and then used for a single
//    comparison before being discarded
//  - the rhs set will be created once for a ton of rows and then used for many comparisons
//  - the lhs set will typically be very small, often only one full name
//  - the rhs set can be of more variable size, but will often have only brief names
//  - the lhs set will be printed at most once
//  - the rhs set will not be printed at all
//
// Consider the query `Hostnames <= /gpu-[1-100]/` (to filter jobs that ran only on GPU nodes) as
// somewhat typical though often the rhs will not be that big.

package table

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"go-utils/hostglob"
)

// The data structure implements the tree literally, with backlinks so that it's possible to walk
// the tree up and generate full names.  Probably there are more efficient ways of doing this.

type node struct {
	me   string           // My element name
	back *node            // My parent, first nodes point back to head node, head has nil
	next map[string]*node // My children
}

func makeNode(me string, back *node) *node {
	return &node{
		me:   me,
		back: back,
		next: make(map[string]*node),
	}
}

func (n *node) String() string {
	s := ""
	for k, v := range n.next {
		if s != "" {
			s += ", "
		}
		s += fmt.Sprintf("%s => %s", k, v.String())
	}
	return "{" + s + "}"
}

type set struct {
	all       []*node // all nodes in the set except head node
	sources   *node   // head node with me == "" and not in all
	lazySinks []*node // inserts will clear this, reconstructed as needed
}

func makeSet() *set {
	return &set{
		all:       make([]*node, 0),  // head node not in the list
		sources:   makeNode("", nil), // head node
		lazySinks: nil,
	}
}

func (s *set) String() string {
	return s.sources.String()
}

func (s *set) isEmpty() bool {
	return len(s.all) == 0
}

// Descend the tree and add the node if not already there.  Dirty the set and return true if new
// information was added.
func (s *set) add(elements []string) bool {
	n := s.sources
	added := false
	for _, e := range elements {
		if probe := n.next[e]; probe != nil {
			n = probe
		} else {
			added = true
			x := makeNode(e, n)
			s.all = append(s.all, x)
			n.next[e] = x
			n = x
		}
	}
	if added {
		s.lazySinks = nil
	}
	return added
}

func (s *set) addNames(names ...string) bool {
	added := false
	for _, name := range names {
		x := s.add(splitName(name))
		added = added || x
	}
	return added
}

// Descend the tree and see if we get to a node, return true if so.
func (s *set) lookup(elements []string) bool {
	n := s.sources
	for _, e := range elements {
		probe := n.next[e]
		if probe == nil {
			return false
		}
		n = probe
	}
	return true
}

// Find all the sinks in the set and return them, caching the list for next time this might be
// needed.
func (s *set) sinks() []*node {
	if s.lazySinks == nil {
		sinks := make([]*node, 0)
		for _, n := range s.all {
			if len(n.next) == 0 {
				sinks = append(sinks, n)
			}
		}
		s.lazySinks = sinks
	}
	return s.lazySinks
}

// Returns 0 if a = b, -1 if b < a, 1 otherwise (in the sense of = and < defined earlier).
// This exits as soon as the "otherwise" case is encountered.  Thus
//
//   a = b    compare(a,b) == 0
//   a <= b   compare(b, a) <= 0
//   a < b    compare(b, a) < 0
//   a >= b   compare(a, b) <= 0
//   a > b    compare(a, b) < 0

func compare(a, b *node) int {
	var result int
	for bName, bNext := range b.next {
		aNext, found := a.next[bName]
		if !found {
			return 1
		}
		down := compare(aNext, bNext)
		if down == 1 {
			return 1
		}
		if down == -1 {
			result = -1
		}
	}
	if result == 0 {
		// everything in b was found in a
		//
		// the extra test on b.next is unpleasant and some kind of indication
		// that we're doing it wrong.  Maybe a precondition to compare() is that
		// b.next not is empty.  Not sure.
		if len(a.next) != len(b.next) && len(b.next) > 0 {
			// a has more than b
			result = -1
		}
	}
	return result
}

func splitName(hostname string) []string {
	elements := strings.Split(hostname, ".")
	if len(elements[0]) == 0 {
		return nil
	}
	elements = slices.DeleteFunc(elements, func(x string) bool {
		return x == ""
	})
	return elements
}

type Hostnames struct {
	s      *set
	serial uint // Used for testing, updated on modifications
}

func NewHostnames() *Hostnames {
	return &Hostnames{
		s: makeSet(),
	}
}

func (h *Hostnames) Add(hostname string) {
	if h.s.addNames(hostname) {
		h.serial++
	}
}

// The nodelist is a "multi-pattern" according to the grammar in ../../go-utils/hostglob.
func (h *Hostnames) AddCompressed(nodelist string) error {
	patterns, err := hostglob.SplitMultiPattern(nodelist)
	if err != nil {
		return err
	}
	for _, p := range patterns {
		names, err := hostglob.ExpandPattern(p)
		if err != nil {
			return err
		}
		for _, n := range names {
			h.Add(n)
		}
	}
	return nil
}

func (h *Hostnames) IsEmpty() bool {
	return h.s.isEmpty()
}

func (h *Hostnames) HasElement(hostname string) bool {
	return h.s.lookup(splitName(hostname))
}

func (this *Hostnames) Equal(that *Hostnames) bool {
	return compare(this.s.sources, that.s.sources) == 0
}

func (a *Hostnames) HasSubset(b *Hostnames, proper bool) bool {
	r := compare(a.s.sources, b.s.sources)
	if proper {
		return r < 0
	}
	return r <= 0
}

// Returns a string that is a comma-separated lists of the first elements of all the hosts in the
// set, in sorted order, without compression.  This is precisely the set of names in the map of
// the head node.

func (h *Hostnames) FormatBrief() string {
	xs := slices.Collect(maps.Keys(h.s.sources.next))
	slices.Sort(xs)
	return strings.Join(xs, ",")
}

// Returns a string that is a comma-separated lists of the hosts in the set, in sorted order,
// without compression.  For this, we need the sinks.  From each sink we walk the graph backward to
// the root, constructing full names as we go.

func (h *Hostnames) FormatFull() string {
	xs := slices.Collect(h.FullNames)
	slices.Sort(xs)
	return strings.Join(xs, ",")
}

func (h *Hostnames) FullNames(yield func(string) bool) {
	for _, n := range h.s.sinks() {
		x := ""
		for {
			x = n.me + x
			if n.back.me == "" {
				break
			}
			x = "." + x
			n = n.back
		}
		if !yield(x) {
			break
		}
	}
}
