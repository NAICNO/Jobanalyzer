// There are four operations on host name patterns and sets of host names.
//
// - We can *match* a pattern or multi-pattern against a set of concrete host names, yielding a
//   selection of those host names
// - We can *expand* a pattern or multi-pattern into a set of concrete host names
// - We can *compress* a set of concrete host names into a pattern or multi-pattern
// - We can *split* a multi-pattern into a set of patterns
//
// The following grammar pertains to all of these:
//
//   multi-pattern   ::= pattern ("," pattern)*
//   pattern         ::= pattern-element ("." pattern-element)*
//   pattern-element ::= fragment+
//   fragment        ::= literal | range | wildcard
//   literal         ::= <longest nonempty string of characters not containing "[" or "," or "*" or ".">
//   range           ::= "[" range-elt ("," range-elt)* "]"
//   range-elt       ::= number | number "-" number
//   number          ::= <nonempty string of 0..9, to be interpreted as decimal>
//   wildcard        ::= "*"
//
//   hostname        ::= host-element ("." host-element)*
//   host-element    ::= literal
//
// The following restrictions apply:
//
// - In a range A-B, A must be no greater than B or the pattern is invalid
// - It is not possible to expand a pattern or multi-pattern that contains a wildcard
// - The expansion of the result of compression of a set of hostnames H must yield exactly
//   the set H
// - Compression does not have a unique result and is not required to be optimal
// - However, compressing the list [y,x] and the list [x,y] must yield the same result (modulo the
//   ordering of the names in the result set), this is important for some consumers.
// - The semantics of matching currently follow the semantics of the regular expression expansion of
//   the pattern-element.  Numbers a, b, ..., z in a range mean /(:?a|b|...|z)/.  A wildcard means
//   /[^.]*/.  The expansion of a pattern-element always starts with /^/ and ends with /$/.  Hence
//   "a[1-3]*b" becomes /^a(?:1|2|3)[^.]*b$/.  This can be confusing: "a[1-3]*b" will actually match
//   the host-element "a14b" because the "1" is matched by the disjunction and the "4" is matched by
//   the wildcard.
//
// Note: There are implementations of the algorithms here both in the Rust code (sonarlog) and in
// the JS code (dashboard/hostglob.js).
//
// Note: matching can be exact (the number of pattern-elements must equal the number of
// host-elements in the hostname) or by prefix (there are fewer pattern-elements than
// host-elements).  In both cases the pattern-elements must match the corresponding host-elements,
// from left towards the right.  In particular, if the pattern is "a*" and the host name is "a.b",
// these must not match if the matching is exact.
//
// Note: an argument can be made that glob semantics are better, so that "a[1-3]*b" would not match
// "a14b": The 14 would be considered "a number" and read as such and matched against {1,2,3}, and
// that would fail.

package hostglob

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go-utils/slices"
)

// This takes a <multi-pattern> according to the grammar above and returns a list of individual
// <pattern>s in that list.  It requires a bit of logic because each pattern may contain a fragment
// that contains a comma.

func SplitMultiPattern(s string) ([]string, error) {
	strings := make([]string, 0)
	if s == "" {
		return strings, nil
	}
	insideBrackets := false
	start := -1
	for ix, c := range s {
		if c == '[' {
			if insideBrackets {
				return nil, fmt.Errorf("Illegal pattern: nested brackets")
			}
			insideBrackets = true
		} else if c == ']' {
			if !insideBrackets {
				return nil, fmt.Errorf("Illegal pattern: unmatched end bracket")
			}
			insideBrackets = false
		} else if c == ',' && !insideBrackets {
			if start == -1 {
				return nil, fmt.Errorf("Illegal pattern: Empty host name")
			}
			strings = append(strings, s[start:ix])
			start = -1
		} else if start == -1 {
			start = ix
		}
	}
	if insideBrackets {
		return nil, fmt.Errorf("Illegal pattern: Missing end bracket")
	}
	if start == len(s) || start == -1 {
		return nil, fmt.Errorf("Illegal pattern: Empty host name")
	}
	strings = append(strings, s[start:])
	return strings, nil
}

// This takes a single <pattern> from the grammar above and expands it.  Restriction: The pattern
// must contain no "*" wildcards.

func ExpandPattern(s string) ([]string, error) {
	before, after, has_tail := strings.Cut(s, ".")
	head_expansions, err := expandPatternElement(before)
	if err != nil {
		return nil, err
	}
	if !has_tail {
		return head_expansions, nil
	}

	tail_expansions, err := ExpandPattern(after)
	if err != nil {
		return nil, err
	}
	expansions := []string{}
	for _, h := range head_expansions {
		for _, t := range tail_expansions {
			expansions = append(expansions, h+"."+t)
		}
	}
	return expansions, nil
}

// MT: Constant after initialization; immutable
var noMoreFragments = errors.New("No more fragments")

type wildcard struct{}

func expandPatternElement(s string) ([]string, error) {
	r := strings.NewReader(s)
	fragments := make([]any, 0)
	for {
		fragment, err := parseFragment(r)
		if err != nil {
			if err == noMoreFragments {
				break
			}
			return nil, err
		}
		if _, ok := fragment.(*wildcard); ok {
			return nil, errors.New("Wildcard not allowed in expansions")
		}
		fragments = append(fragments, fragment)
		if len(fragments) > 100 {
			return nil, errors.New("Unlikely hostname pattern")
		}
	}
	if len(fragments) == 0 {
		return nil, errors.New("Empty element")
	}
	tails := []string{""}
	for i := len(fragments) - 1; i >= 0; i-- {
		switch f := fragments[i].(type) {
		case string:
			xs := make([]string, 0, len(tails))
			for _, t := range tails {
				xs = append(xs, f+t)
			}
			tails = xs
		case []uint64:
			xs := make([]string, 0, len(tails)*len(f))
			for _, t := range tails {
				for _, n := range f {
					xs = append(xs, fmt.Sprintf("%d%s", n, t))
				}
			}
			tails = xs
		default:
			panic("???")
		}
	}
	return tails, nil
}

func parseFragment(r *strings.Reader) (any, error) {
	switch c := getc(r); c {
	case 0:
		return nil, noMoreFragments
	case '*':
		return &wildcard{}, nil
	case '[':
		needOne := true
		nodes := []uint64{}
		for {
			if len(nodes) > 50000 {
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
				for n <= m {
					if len(nodes) > 50000 {
						return nil, errors.New("Too many elements")
					}
					nodes = append(nodes, n)
					n++
				}
			} else {
				nodes = append(nodes, n)
			}
			if eatc(r, ',') {
				needOne = true
			} else if eatc(r, ']') {
				ungetc(r, ']')
			} else {
				return nil, errors.New("Unexpected character")
			}
		}
		return nodes, nil
	case ',':
		return nil, errors.New("Unexpected ','")
	case '.':
		return nil, errors.New("Unexpected '.'")
	default:
		literal := string(c)
		for {
			c := getc(r)
			if c == 0 || c == '[' || c == ',' || c == '.' || c == '*' {
				ungetc(r, c)
				break
			}
			literal = literal + string(c)
		}
		return literal, nil
	}
}

// Given a list of valid <hostname>s by the grammar above, return an abbreviated list that uses
// <pattern> syntax where possible.  The patterns will contain no "*" wildcards.
//
// In general, if there are several compressible ranges within the host names we must pick one.  For
// example, for the set {a1.b1, a2.b2} we naively have two ranges if we consider the host-elements
// separately, but a[1,2].b[1,2] is not a valid compression as it names too many hosts.  While there
// are sets of host names that would allow multiple ranges to be compressed (for example, the set
// resulting from the expansion of that incorrect pattern), this is not a special case worth looking
// for.
//
// For simplicity, for host names of the form `a.b.c...` we will not try to compress anything in the
// `b.c...` portion, and within the `a` portions we will try to compress only the rightmost digit
// strings.  This will yield good results in general.

// MT: Constant after initialization; immutable (except for configuration methods).
var withDigitsRe = regexp.MustCompile(`^(.*?)(\d+)(\D*)$`)

func CompressHostnames(hosts []string) []string {
	// Suffixes is a map from `b.c...` portion to `a` portion of name.
	suffixes := make(map[string][]string)
	for _, h := range hosts {
		before, after, _ := strings.Cut(h, ".")
		if probe, ok := suffixes[after]; ok {
			suffixes[after] = append(probe, before)
		} else {
			suffixes[after] = []string{before}
		}
	}

	// Complete host names
	result := make([]string, 0)

	// Compress the first host-elements, catenate with suffixes
	for suffix, firstelts := range suffixes {
		same := make(map[string][]int)
		for _, elt := range firstelts {
			ms := withDigitsRe.FindStringSubmatch(elt)
			if ms == nil {
				result = pushHostName(result, elt, suffix)
				continue
			}
			n, err := strconv.ParseInt(ms[2], 10, 64)
			if err != nil {
				result = pushHostName(result, elt, suffix)
				continue
			}
			name := ms[1] + "," + ms[3]
			if bag, found := same[name]; found {
				same[name] = append(bag, int(n))
			} else {
				same[name] = []int{int(n)}
			}
		}
		for k, v := range same {
			a, b, _ := strings.Cut(k, ",")
			result = pushHostName(result, a+compressRange(v)+b, suffix)
		}
	}

	return result
}

func pushHostName(result []string, elt, suffix string) []string {
	if suffix != "" {
		return append(result, elt+"."+suffix)
	}
	return append(result, elt)
}

func compressRange(xs []int) string {
	if len(xs) == 1 {
		return fmt.Sprintf("%d", xs[0])
	}
	sort.Sort(sort.IntSlice(xs))
	s := ""
	for i := 0; i < len(xs); {
		first := xs[i]
		prev := first
		i++
		for i < len(xs) && xs[i] == prev+1 {
			prev = xs[i]
			i++
		}
		if s != "" {
			s += ","
		}
		if first != prev {
			s += fmt.Sprintf("%d-%d", first, prev)
		} else {
			s += fmt.Sprintf("%d", first)
		}
	}
	s = "[" + s + "]"
	return s
}

// A `HostGlobber` is a matcher of patterns against hostnames.
//
// The matcher holds a number of patterns, added with `insert`.  Each `pattern` is a <pattern> in
// the sense of the grammar referenced above.
//
// The `match_hostname` method attempts to match its argument against the patterns in the matcher,
// returning true if any of them match.
//
// The hostGlobber is immutable and thread-safe: the embedded regexes are defined to be thread-safe
// for concurrent use. cf https://pkg.go.dev/regexp#Regexp:
//
//  A Regexp is safe for concurrent use by multiple goroutines, except for configuration methods,
//  such as Regexp.Longest.
//
// However note that the RegExp machinery has some shared global state internally in a sync.Map, and
// so executing a regexp is not necessarily lock-free.  Hence the hostglobber is also not
// necessarily lock-free.  But I'm guessing that map is not heavily contended, TBD.  It could look
// like it is only used for fairly complex regexes; ours will tend to be simple (single host
// pattern that should require no backtracking).

type HostGlobber struct {
	isPrefixMatcher bool
	matchers        []*regexp.Regexp
}

// Create a new, empty filter.  The flag indicates whether the globbers in the filter could match
// only a prefix of the hostname elements or must match all of them.

func NewGlobber(isPrefixMatcher bool, patterns []string) (*HostGlobber, error) {
	matchers := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, _, err := compileGlobber(p, isPrefixMatcher)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, re)
	}
	return &HostGlobber{
		isPrefixMatcher: isPrefixMatcher,
		matchers:        matchers,
	}, nil
}

// Return true iff the filter has no patterns.

func (hg *HostGlobber) IsEmpty() bool {
	return len(hg.matchers) == 0
}

func (hg *HostGlobber) String() string {
	return fmt.Sprint(
		slices.Map(hg.matchers, func(re *regexp.Regexp) string {
			return re.String()
		}),
	)
}

// Match s against the patterns and return true iff it matches at least one pattern.

func (hg *HostGlobber) Match(host string) bool {
	for _, re := range hg.matchers {
		if re.Match([]byte(host)) {
			return true
		}
	}
	return false
}

func compileGlobber(p string, prefix bool) (*regexp.Regexp, string, error) {
	in := strings.NewReader(p)
	r := "^"
Parser:
	for {
		if len(r) > 50000 {
			return nil, "", errors.New("Expression too large, use more '*'")
		}
		switch c := getc(in); c {
		case 0:
			break Parser
		case '*':
			r += "[^.]*"
		case '[':
			ns := make([]string, 0)
			for {
				n0, err := readNumberLimited(in, 0xFFFFFFFF)
				if err != nil {
					return nil, "", err
				}
				if eatc(in, '-') {
					n1, err := readNumberLimited(in, 0xFFFFFFFF)
					if err != nil {
						return nil, "", err
					}
					if n0 > n1 {
						return nil, "", errors.New("Invalid range")
					}
					for n0 <= n1 {
						ns = append(ns, fmt.Sprint(n0))
						n0++
						if len(ns) > 10000 {
							return nil, "", errors.New("Range too large, use more '*'")
						}
					}
				} else {
					ns = append(ns, fmt.Sprint(n0))
				}
				if eatc(in, ']') {
					break
				}
				if !eatc(in, ',') {
					return nil, "", errors.New("Expected ','")
				}
			}
			r += "(?:"
			r += strings.Join(ns, "|")
			r += ")"
		case '.', '$', '^', '?', '\\', '(', ')', '{', '}', ']':
			// Mostly these special chars are not allowed in hostnames, but it doesn't hurt to
			// try to avoid disasters.
			r += "\\"
			r += string(c)
		default:
			r += string(c)
		}
	}
	if prefix {
		// Matching a prefix: we need to match entire host-elements, so following a prefix there
		// should either be EOS or `.` followed by whatever until EOS.
		r += "(?:\\..*)?$"
	} else {
		r += "$"
	}
	re, err := regexp.Compile(r)
	if err != nil {
		return nil, "", err
	}
	return re, r, nil
}

// Common operations on a RuneScanner

func readNumberLimited(r io.RuneScanner, max uint64) (uint64, error) {
	n, err := readNumber(r)
	if err != nil {
		return n, err
	}
	if n > max {
		return 0, errors.New("Number out of range")
	}
	return n, nil
}

func readNumber(r io.RuneScanner) (uint64, error) {
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
	return strconv.ParseUint(cs, 10, 64)
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
