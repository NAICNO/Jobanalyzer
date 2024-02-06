// Expand, split, and otherwise manipulate host name sets.
//
// For a specification of this, see the block comment starting sonarlog/src/hosts.rs and the code
// for expand_patterns() in that file, and the code for expand_element() in sonarlog/src/pattern.rs.
//
// This is a very limited version, for now.  We expand single trailing ranges of node indices in
// each element of the host name, according to this grammar:
//
//  hostname ::= element ("." element)*
//  element ::= literal range?
//  literal ::= <nonempty string of characters not containing '[' or ',' or '*'>
//  range ::= '[' range-elt ("," range-elt)* ']'
//  range-elt ::= number | number "-" number
//  number ::= <nonempty string of 0..9, to be interpreted as decimal>
//
// If the element is syntactically invalid, the unexpanded value is returned.  In a range A-B, A
// must be no greater than B.

package hostglob

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// This takes as its input a string representing a comma-separated list of hostnames (according to
// the grammar above) and returns a list of individual hostnames in that list.  It requires a bit of
// logic because each hostname may contain a pattern that contains a comma.

func SplitHostnames(s string) ([]string, error) {
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

func ExpandPatterns(s string) []string {
	before, after, has_tail := strings.Cut(s, ".")
	head_expansions := ExpandPattern(before)
	if !has_tail {
		return head_expansions
	}

	tail_expansions := ExpandPatterns(after)
	expansions := []string{}
	for _, h := range head_expansions {
		for _, t := range tail_expansions {
			expansions = append(expansions, h+"."+t)
		}
	}
	return expansions
}

func ExpandPattern(s string) []string {
	prefix, nodenums := parseElement(s)
	if len(nodenums) == 0 {
		return []string{s}
	}

	expansions := []string{}
	for _, nn := range nodenums {
		expansions = append(expansions, fmt.Sprintf("%s%d", prefix, nn))
	}
	return expansions
}

func parseElement(element string) (string, []int) {
	r := strings.NewReader(element)
	literal := ""
	for {
		c := getc(r)
		if c == 0 || c == '[' || c == '*' || c == ',' {
			ungetc(r, c)
			break
		}
		literal = literal + string(c)
	}
	nodes := []int{}
	needOne := false
	switch getc(r) {
	case 0:
		// Nothing
	case '[':
		for {
			if eatc(r, ']') {
				if needOne {
					goto fail
				}
				break
			}
			needOne = false
			n, err := readNumber(r)
			if err != nil {
				goto fail
			}
			if eatc(r, '-') {
				m, err := readNumber(r)
				if err != nil || n > m {
					goto fail
				}
				for n <= m {
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
				goto fail
			}
		}
		if getc(r) != 0 {
			goto fail
		}
	default:
		goto fail
	}
	return literal, nodes

fail:
	return element, []int{}
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
	return strconv.Atoi(cs)
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

// Given a list of host names, return an abbreviated list that uses host number sets where possible.
//
// Good enough: Host names match /^(.*)-(\d+)$/ and we can compress the values of \2 for equal \1.
//
// Better: there can be several runs of digits that could be compressed individually, but let's not
// worry about that yet.

func CompressHostnames(hosts []string) ([]string, error) {
	same := make(map[string][]int)
	for _, h := range hosts {
		k := strings.LastIndexAny(h, "-")
		if !(k > 0 && k < len(h)-1) {
			return nil, fmt.Errorf("Bad host name %s", h)
		}
		n, err := strconv.ParseInt(h[k+1:], 10, 64)
		if err != nil {
			return nil, err
		}
		name := h[:k+1]
		if bag, found := same[name]; found {
			same[name] = append(bag, int(n))
		} else {
			same[name] = []int{int(n)}
		}
	}
	result := make([]string, 0)
	for k, v := range same {
		result = append(result, k+compressRange(v))
	}
	return result, nil
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
