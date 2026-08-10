package common

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"slices"
	"strconv"
	"strings"
)

// A HostnameSet is never empty, and it is immutable.  It represent a set of host names a . b . ...
// where the first element can contain a node range (hence "set").  The node range is either a node
// set represented explicitly as [ a-b, x, y, ... ] or an implied range of size 1, the last digit
// string in a.
//
// The set can be constructed by parsing a string or by unioning existing sets.
//
// Sets can be formatted as strings.  Those strings are canonical: For a given set of hosts,
// regardless of how the set was constructed, the printed representation is the same.
//
// The "nodes" can be empty (for host names that contain no implied-ranges); then the suffix is also
// empty and the name is <prefix "." tail[0] "."  tail[1] ...>.
//
// TODO: There is a grammar for this, which needs to be included here.
type HostnameSet struct {
	prefix string
	suffix string
	tail   []string
	nodes  nodeset
}

// String returns the canonical string representation for the set.  If self represents a single host
// (no node set, or the node set has one element) then this must return a string without range
// syntax.
func (self *HostnameSet) String() string {
	var tail string
	if self.tail != nil {
		tail = "." + strings.Join(self.tail, ".")
	}
	switch (self.nodes.size()) {
	case 0:
		return self.prefix + tail
	case 1:
		return self.prefix + self.nodes.singletonName() + self.suffix + tail
	default:
		return self.prefix + self.nodes.compressedName() + self.suffix + tail
	}
}

// Expand returns an iterator that will yield the individual host names in the set, without any
// range syntax.  If there's a set with more than one element, they are generated in numerically
// ascending order.
func (self *HostnameSet) Expand() iter.Seq[string] {
	var tail string
	if self.tail != nil {
		tail = "." + strings.Join(self.tail, ".")
	}
	switch (self.nodes.size()) {
	case 0:
		return func(yield func(string) bool) {
			yield(self.prefix + tail)
		}
	case 1:
		return func(yield func(string) bool) {
			yield(self.prefix + self.nodes.singletonName() + self.suffix + tail)
		}
	default:
		return func(yield func(string) bool) {
			for n := range self.nodes.iter() {
				if !yield(self.prefix + strconv.Itoa(n) + self.suffix + tail) {
					return
				}
			}
		}
	}
}

// IsSingle returns true if the set contains zero or one host names.  Zero is unusual, it cannot
// result from parsing a string, mostly it means somebody has used a zero HostnameSet value, usually
// in an error return.
func (self *HostnameSet) IsSingle() bool {
	return self.nodes.size() <= 1
}

// Match matches the other set against self, which means that every host in the other set must be in
// the self set, ie, it is a subset query.
func (self *HostnameSet) Match(other HostnameSet) bool {
	if self.prefix != other.prefix || self.suffix != other.suffix || !slices.Equal(self.tail, other.tail) {
		return false
	}
	if !self.nodes.empty() && other.nodes.empty() {
		return false
	}
	for o := range other.nodes.iter() {
		if !self.nodes.member(o) {
			return false
		}
	}
	return true
}

// UnionHostnameSets unions the unionable sets in the input list, returning a list of unions.  Two
// sets are unionable if they have the same prefix, suffix, and tail.
func UnionHostnameSets(hss []HostnameSet) []HostnameSet {
	type hnsInfo struct {
		base   *HostnameSet
		nodes *nodeset
	}
	fixed := make(map[string]*hnsInfo)
	for _, x := range hss {
		key := x.prefix + "|" + x.suffix + "|" + strings.Join(x.tail, "|")
		if probe := fixed[key]; probe != nil {
			probe.nodes.insertAll(&x.nodes)
		} else {
			fixed[key] = &hnsInfo{
				base:   &x,
				nodes:  x.nodes.clone(),
			}
		}
	}
	result := make([]HostnameSet, 0, len(fixed))
	for _, v := range fixed {
		result = append(result, HostnameSet{
			prefix: v.base.prefix,
			suffix: v.base.suffix,
			tail:   v.base.tail,
			nodes:  *v.nodes,
		})
	}
	return result
}

func parseConcretePatterns(patterns []string) ([]HostnameSet, error) {
	result := make([]HostnameSet, 0, len(patterns))
	for _, p := range patterns {
		parsed, err := parseConcretePattern(p)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

func parseConcretePattern(pattern string) (HostnameSet, error) {
	return parsePattern(pattern, true)
}

func parseHostname(s string) (HostnameSet, error) {
	return parsePattern(s, false)
}

func parsePattern(s string, allowRange bool) (HostnameSet, error) {
	elements := strings.Split(s, ".")

	// Parse head element
	head, err := parsePatternElement(elements[0], allowRange, /*allowImpliedRange=*/ true)
	if err != nil {
		return HostnameSet{}, err
	}

	// Convert the last implied range to a node set, the rest to string
	convert := false
	for j := len(head)-1 ; j >= 0 ; j-- {
		if r, ok := head[j].(impliedRange); ok {
			if convert {
				head[j] = strconv.Itoa(int(r))
			} else {
				head[j] = nodeset{[]nrange{nrange{int(r), int(r)}}}
				convert = true
			}
		}
	}

	// Merge adjacent strings and check that there's at most one range
	head = mergeAdjacentStrings(head)
	var prefix, suffix string
	var nodes nodeset
	i := 0
	if s, ok := head[i].(string); ok {
		prefix = s
		i++
	}
	if i < len(head) {
		if ns, ok := head[i].(nodeset); ok {
			nodes = ns
			i++
		}
	}
	if i < len(head) {
		if s, ok := head[i].(string); ok {
			suffix = s
			i++
		}
	}
	if i != len(head) {
		return HostnameSet{}, errors.New("Malformed host name head")
	}

	// Parse / syntax check the tail elements
	tail := make([]string, len(elements)-1)
	for i, e := range elements[1:] {
		var err error
		xs, err := parsePatternElement(e, false, /*allowImpliedRange=*/ false)
		if err != nil {
			return HostnameSet{}, err
		}
		if len(xs) != 1 {
			return HostnameSet{}, errors.New("Malformed host name tail element")
		}
		if s, ok := xs[0].(string); ok {
			tail[i] = s
		} else {
			return HostnameSet{}, errors.New("Malformed host name tail element")
		}
	}

	return HostnameSet{
		prefix: prefix,
		suffix: suffix,
		nodes: nodes,
		tail: tail,
	}, nil
}

type impliedRange int

var noMoreFragments = errors.New("No more fragments")

// The members of the []any result can be implied-range, nodeset, or string, with the following
// restrictions.
//
// If allowRange is false then a range will trigger an error and nodeset will not be returned.
//
// If allowImpliedRange is false then impliedRange will be converted to string here and merged into
// adjacent strings, and impliedRange will not be returned.

func parsePatternElement(s string, allowRange, allowImpliedRange bool) ([]any, error) {
	r := strings.NewReader(s)
	fragments := make([]any, 0)
	for {
		fragment, err := tokenizeFragment(r)
		if err != nil {
			if err == noMoreFragments {
				break
			}
			return nil, err
		}
		if !allowRange {
			if _, ok := fragment.(nodeset); ok {
				return nil, errors.New("Ranges not allowed")
			}
		}
		if !allowImpliedRange {
			if i, ok := fragment.(impliedRange); ok {
				fragment = strconv.Itoa(int(i))
			}
		}
		fragments = append(fragments, fragment)
		if len(fragments) > 100 {
			return nil, errors.New("Unlikely hostname pattern")
		}
	}
	if len(fragments) == 0 {
		return nil, errors.New("Empty element")
	}
	if !allowImpliedRange {
		fragments = mergeAdjacentStrings(fragments)
	}
	return fragments, nil
}

func mergeAdjacentStrings(xs []any) []any {
	var result []any
	curr := ""
	for _, v := range xs {
		if s, ok := v.(string); ok {
			curr += s
		} else {
			if curr != "" {
				result = append(result, curr)
				curr = ""
			}
			result = append(result, v)
		}
	}
	if curr != "" {
		result = append(result, curr)
	}
	return result
}

// Note that an implied-range or a value in a range cannot start with zero, this is an interesting
// wrinkle.  It's most interesting for implied-range: the 0 becomes part of the preceding literal.
// Surely something will break somewhere because of this, but I think it's inevitable.

func tokenizeFragment(r *strings.Reader) (any, error) {
	switch c := getc(r); c {
	case 0:
		return nil, noMoreFragments
	case ',':
		return nil, errors.New("Unexpected ','")
	case '.':
		return nil, errors.New("Unexpected '.'")
	case '*':
		return nil, errors.New("Unexpected '*'")
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		ungetc(r, c)
		n, _ := readNumber(r)
		return impliedRange(n), nil
	case '[':
		needOne := true
		var nodes nodeset
		var count int
		for {
			if count > 50000 {
				return nil, errors.New("Too many elements")
			}
			if eatc(r, ']') {
				if needOne {
					return nil, errors.New("Expected number")
				}
				break
			}
			needOne = false
			n, err := readNumber(r)
			if err != nil {
				return nil, err
			}
			if eatc(r, '-') {
				m, err := readNumber(r)
				if err != nil {
					return nil, err
				}
				if n > m {
					return nil, errors.New("Bad range")
				}
				nodes.insertRange(nrange{n, m})
			} else {
				nodes.insertRange(nrange{n, n})
			}
			count++
			if eatc(r, ',') {
				needOne = true
			} else if eatc(r, ']') {
				ungetc(r, ']')
			} else {
				return nil, errors.New("Unexpected character")
			}
		}
		return nodes, nil
	default:
		literal := string(c)
		for {
			c := getc(r)
			if c == 0 || c == '[' || c == ',' || c == '.' || c == '*' || c >= '1' && c <= '9' {
				ungetc(r, c)
				break
			}
			// TODO: restrictions on spaces?
			literal = literal + string(c)
		}
		return literal, nil
	}
}

func readNumber(r io.RuneScanner) (int, error) {
	cs := ""
	for {
		c := getc(r)
		if c < '0' || c > '9' {
			ungetc(r, c)
			break
		}
		cs = cs + string(c)
	}
	if cs == "" {
		return 0, errors.New("Expected number")
	}
	n, err := strconv.ParseInt(cs, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func eatc(r io.RuneScanner, x rune) bool {
	c := getc(r)
	if c == x {
		return true
	}
	ungetc(r, c)
	return false
}

func getc(r io.RuneScanner) rune {
	c, _, err := r.ReadRune()
	if err == io.EOF {
		return 0
	}
	return c
}

func ungetc(r io.RuneScanner, c rune) {
	if c != 0 {
		r.UnreadRune()
	}
}

// Numeric sets of nodes.
//
// Various ways to represent this, but normally there's a small set of ranges and individual nodes,
// so just use a slice of ranges.  Ranges are from..to inclusive, no ranges can be empty.
//
// Invariants:
// - ranges are sorted in ascending 'from' order
// - all 'from' values are distinct
// - no two ranges are adjacent or overlapping: if b directly follows a then b.from > a.to + 1

type nrange struct {
	from, to int
}

type nodeset struct {
	// Sorted ascending, and no ranges overlap or abut the next.
	ranges []nrange
}

func (n *nodeset) clone() *nodeset {
	return &nodeset{ ranges: slices.Clone(n.ranges) }
}

func (n *nodeset) empty() bool {
	return n.size() == 0
}

func (n *nodeset) size() int {
	sz := 0
	for _, r := range n.ranges {
		sz += r.to - r.from + 1
	}
	return sz
}

func (n *nodeset) insertRange(r nrange) {
	// note that a new range can cover many existing ranges which may need to be
	// absorbed into it.
	i := 0
	for i < len(n.ranges) && n.ranges[i].to < r.from-1 {
		i++
	}
	j := i
	for j < len(n.ranges) && r.to >= n.ranges[j].from-1 {
		j++
	}
	// Affected ranges are from i..j-1 inclusive.  These can be removed and their bounds folded into
	// the bounds of r.  Note there may be none of them.
	//
	// TODO: Suboptimal for sure, because commonly r has no overlap and follows all existing ranges.
	result := make([]nrange, i+len(n.ranges)-j+1)
	copy(result[:i], n.ranges[:i])
	if j < len(n.ranges) {
		copy(result[i+1:], n.ranges[j:])
	}
	if i < len(n.ranges) {
		r.from = min(r.from, n.ranges[i].from)
	}
	if j > 0 {
		r.to = max(r.to, n.ranges[j-1].to)
	}
	result[i] = r
	//fmt.Println(i, j, len(n.ranges), r, result)
	n.ranges = result
}

func (n *nodeset) insertAll(other *nodeset) {
	// TODO: suboptimal for sure, since they are sorted.
	for _, o := range other.ranges {
		n.insertRange(o)
	}
}

func (n *nodeset) member(i int) bool {
	for _, r := range n.ranges {
		if i >= r.from && i <= r.to {
			return true
		}
	}
	return false
}

func (n *nodeset) singletonName() string {
	if n.size() != 1 {
		panic("Size must be 1")
	}
	return strconv.Itoa(n.ranges[0].from)
}

func (n *nodeset) compressedName() string {
	if n.empty() {
		panic("Must not be empty")
	}
	var s string
	for _, r := range n.ranges {
		if s != "" {
			s += ","
		}
		if r.from == r.to {
			s += strconv.Itoa(r.from)
		} else {
			s += fmt.Sprintf("%d-%d", r.from, r.to)
		}
	}
	return "[" + s + "]"
}

func (n *nodeset) iter() iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, r := range n.ranges {
			for v := r.from ; v <= r.to ; v++ {
				if !yield(v) {
					return
				}
			}
		}
	}
}

