// A concrete hostname is a string matching the <hostname> non-terminal of the grammar below.
// Hostnames can be merged into sets for fast matching and compact representation.  Those sets can
// be printed and parsed according to the <multi-pattern>, <pattern>, and <hostname> non-terminals.
//
//	multi-pattern              ::= pattern ("," pattern)*
//	pattern                    ::= initial-pattern-element ("." subsequent-pattern-element)*
//	initial-pattern-element    ::= (literal | multi)+
//  multi                      ::= range | wildcard
//  subsequent-pattern-element ::= literal
//	literal                    ::= <longest nonempty string of characters not containing "[" or "," or "*" or ".">
//	range                      ::= actual-range | implied-range
//  actual-range               ::= "[" range-elt ("," range-elt)* "]"
//	range-elt                  ::= number | number "-" number
//  implied-range              ::= number
//	number                     ::= <longest nonempty string of 0..9, to be interpreted as decimal>
//	wildcard                   ::= "*"
//	hostname                   ::= host-element ("." host-element)*
//	host-element               ::= literal
//
// Restrictions:
// - In a <range-elt> A-B, A must be no greater than B or the pattern is invalid.
// - No more than one wildcard or range may be used in the initial element (after range reduction).
// - Ranges and wildcards may not be used together.
//
// Note the grammar is ambiguous as it stands: the <number> "37" could be part of a <literal> or it
// could be an <implied-range>.  When parsing, preference is given to <implied-range>.
//
// After parsing, the last <implied-range> in the parse reduces to a <range>, while earlier
// <implied-range>s reduce to <literal>s.  (This is intended to match clusters that use the
// <prefix><island>-<node><suffix> node naming scheme: we capture the node number in a set but the
// island becomes part of the prefix.  It would be easy to generalize the parsing so that it doesn't
// have to be the last <implied-range> that is converted to a range, but (say) the first, should
// clusters use that host name scheme.)
//
// After range reduction, adjacent <literals>, should there be any, are merged.
//
// This leaves a prefix, at most a single range or wildcard, a suffix for the initial element and a
// tail sequence of strings for the subsequent elements.
//
// Note: by construction, no elements are empty.
//
// ---
//
// The grammar is general and the following additional restrictions apply to make it tractable to
// work with hostnames.
//
//   - In a range A-B, A must be no greater than B or the pattern is invalid
//   - It is not possible to expand a pattern or multi-pattern that contains a wildcard
//   - The expansion of the result of compression of a set of hostnames H must yield exactly
//     the set H
//   - Compression does not have a unique result and is not required to be optimal
//   - However, compressing the list [y,x] and the list [x,y] must yield the same result (modulo the
//     ordering of the names in the result set), this is important for some consumers.
//   - The semantics of matching currently follow the semantics of the regular expression expansion of
//     the pattern-element.  Numbers a, b, ..., z in a range mean /(:?a|b|...|z)/.  A wildcard means
//     /[^.]*/.  The expansion of a pattern-element always starts with /^/ and ends with /$/.  Hence
//     "a[1-3]*b" becomes /^a(?:1|2|3)[^.]*b$/.  This can be confusing: "a[1-3]*b" will actually match
//     the host-element "a14b" because the "1" is matched by the disjunction and the "4" is matched by
//     the wildcard.
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
//
// There are four operations on host name patterns and sets of host names.
//
//   - We can *match* a pattern or multi-pattern against a set of concrete host names, yielding a
//     selection of those host names
//   - We can *expand* a pattern or multi-pattern into a set of concrete host names
//   - We can *compress* a set of concrete host names into a pattern or multi-pattern
//   - We can *split* a multi-pattern into a set of patterns

package common

// An intset can be empty, be a combination of individual bits and ranges, or be "all" bits.  For
// anything but "all" bits we could use https://github.com/bits-and-blooms/bitset for simplicity.
// Then intset becomes union(false | bitset.Bitset) maybe.

type intSet struct {

}

// To match a host name hn against a HostnameSet x, first call ParseHostname on hn to get a
// HostnameSet h that represents hn (hn must have no * or range, at most implied-range).  Then the
// matching of h against x must test whether the suffix and prefix match and whether the integer (if
// any) in the nums of h is in the nums of x.  If so, then the tails must match: either exactly, or
// the tail of h matches a prefix of the tail of x.

// HostnameSets with no * can be intersected and unioned but that is logic that does not necessarily
// belong in this file.

// A list of HostnameSets with no * can be canonicalized by successively merging elements a and b
// with the same prefix, suffix, and tail, by unioning the nums of a and b.  As prefix, suffix, and
// tail are all string data, a matching key can be built as the string
// "prefix|suffix|tail[0]|tail[1]|...".  (This can be cached with the HostnameSet.)

// This is a simple set of names with the same prefix, wildcard, and suffix.
type HostnameSet struct {
	prefix string
	nums   intSet				// AnyIntSet for *, EmptyIntSet for nothing
	suffix string
	tail   []string
}

// This parses a multi-pattern and returns a list of HostnameSets, where each set is for a unique
// sequence of prefix-wildcard-suffix and unique tail.  The construction of the sets is such that
// they are *not* merged, but each pattern in the multi-pattern is represented, and they are in order:
//
//   a[1-5]b.foo,a6b.foo,a7.foo,a3b.foo.bar,b*.x
// =>
//   [{{a, [1-5], b}, [foo]},
//    {{a, [6], b}, [foo]},
//    {{a, [7], ""}, [foo]},
//    {{a, [3], b}, [foo, bar]},
//    {{b, *, ""}, [x]}]
//
// If any parse fails, the whole parse fails.
func ParseMultiPattern(mp string) ([]HostnameSet, error) {
	patterns, err := SplitMultiPattern(mp)
	if err != nil {
		return nil, err
	}
	var s HostnameSet
	for _, p := range patterns {
		elts := strings.Split(p, ".")
		pp, err := parseInitialElt(elts[0])
		if err != nil {
			return nil, err
		}
		// now convert all but the last implied-range to a literal, join any adjacent literals,
		// and convert the last implied-range to a range.
		for _, s := range elts[1:] {
			if err := syntaxCheckSubsequentElt(s); err != nil {
				return nil, err
			}
		}
	}
}

// string | wildcard | intset | number
type pelem any

// Syntax check and convert.  Each subslice represents a pattern-element.  Only the first subslice
// can have other than one element, and must have at least one literal.
func parsePattern(p string) ([][]pelem, error) {
}
