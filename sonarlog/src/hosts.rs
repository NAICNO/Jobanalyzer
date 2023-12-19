/// Matcher and formatter for host names.
///
/// The matcher holds a number of patterns, added with `add_pattern`.  Each `pattern` is a vector of
/// element patterns.
///
/// During matching, a string is turned into an element vector (elements separated by `.`), and the
/// patterns are matched one-by-one against the element vector.  If a pattern is longer than the
/// elements vector then the match will fail.  Otherwise, if `exhaustive` is true and the pattern
/// differs in length from the element vector, then the match will fail.  Otherwise, the pattern
/// elements are applied elementwise to (a prefix of) the element vector, and the match succeeds if
/// all the element matches succeed.
///
/// Grammar for patterns:
///
///  pattern ::= element+
///  element ::= eltpat ("," eltpat)*
///  eltpat ::= primitive+ star?
///  primitive ::= literal | range
///  literal ::= <character not containing '[' or '*' or '.' or ','> +
///  range ::= '[' range-elt ("," range-elt)* ']'
///  range-elt ::= number | number "-" number
///  star ::= '*'
///
/// The meaning of a range is that it is expanded into the set of numbers it contains; by inserting
/// these numbers into the eltpat we get a number of new eltpats, which are subject to further
/// expansion.  This is guaranteed to terminate since the expansion cannot yield further expandable
/// primitives.
///
/// Thus after expansion a pattern is a number of literal strings optionally with a * at the end,
/// denoting an open suffix.
///
/// The formatter will reconstruct a pattern from a set of names (though it will not use wildcards).
use crate::pattern;

use anyhow::Result;

pub struct HostFilter {
    // The outer bool is `exhaustive`, the inner bool is `prefix`
    matchers: Vec<(bool, Vec<(bool, String)>)>,
}

impl HostFilter {
    /// Create a new, empty filter.

    pub fn new() -> HostFilter {
        HostFilter { matchers: vec![] }
    }

    /// Add a new pattern.

    pub fn add_pattern(&mut self, patterns: Vec<String>, exhaustive: bool) -> Result<()> {
        // Each element of `patterns` can be expanded into a set of strings and we basically need to
        // push a new pattern for each of these.  This is not the most efficient way to perform
        // matching but probably good enough, and in practice only the first element of `patterns`
        // will have multiple expansions.

        for patterns in expand_patterns(&patterns)? {
            self.matchers.push((exhaustive, patterns));
        }
        Ok(())
    }

    /// Convenience method: split the pattern string into element patterns and add a pattern with
    /// those element patterns.

    pub fn insert(&mut self, pattern: &str) -> Result<()> {
        self.add_pattern(
            pattern
                .split('.')
                .map(|x| x.to_string())
                .collect::<Vec<String>>(),
            false,
        )
    }

    /// Return true iff the filter has no patterns.

    pub fn is_empty(&self) -> bool {
        self.matchers.len() == 0
    }

    /// Match s against the patterns and return true iff it matches at least one pattern.

    pub fn contains(&self, s: &str) -> bool {
        let components = s.split('.').collect::<Vec<&str>>();
        'try_matcher: for (exhaustive, pattern) in &self.matchers {
            if pattern.len() > components.len() {
                continue 'try_matcher;
            }
            if *exhaustive && pattern.len() != components.len() {
                continue 'try_matcher;
            }
            for i in 0..pattern.len() {
                let (prefix, ref pattern) = pattern[i];
                if prefix {
                    if !components[i].starts_with(pattern) {
                        continue 'try_matcher;
                    }
                } else {
                    if components[i] != pattern {
                        continue 'try_matcher;
                    }
                }
            }
            return true;
        }
        return false;
    }
}

pub fn expand_patterns(xs: &[String]) -> Result<Vec<Vec<(bool, String)>>> {
    if xs.len() == 0 {
        Ok(vec![vec![]])
    } else {
        let rest = expand_patterns(&xs[1..])?;
        let expanded = pattern::expand_element(&xs[0])?;
        let mut result = vec![];
        for e in expanded {
            for r in &rest {
                let is_prefix = e.ends_with('*');
                let text = if is_prefix {
                    e[..e.len() - 1].to_string()
                } else {
                    e.to_string()
                };
                let mut m = vec![(is_prefix, text)];
                m.extend_from_slice(&r);
                result.push(m);
            }
        }
        Ok(result)
    }
}

#[test]
fn test_hostfilter1() {
    let mut hf = HostFilter::new();
    hf.add_pattern(vec!["ml8".to_string()], false).unwrap();
    hf.add_pattern(
        vec![
            "ml4".to_string(),
            "hpc".to_string(),
            "uio".to_string(),
            "no".to_string(),
        ],
        true,
    )
    .unwrap();
    hf.add_pattern(vec!["ml3".to_string(), "hpc".to_string()], false)
        .unwrap();

    // Single-element prefix match against this
    assert!(hf.contains("ml8.hpc.uio.no"));

    // Multi-element prefix match against this
    assert!(hf.contains("ml3.hpc.uio.no"));

    // Exhaustive match against this
    assert!(hf.contains("ml4.hpc.uio.no"));
    assert!(!hf.contains("ml4.hpc.uio.no.yes"));
}

#[test]
fn test_hostfilter2() {
    let mut hf = HostFilter::new();
    hf.add_pattern(vec!["ml[1-3]*".to_string()], false).unwrap();
    assert!(hf.contains("ml1"));
    assert!(hf.contains("ml1x"));
    assert!(hf.contains("ml1.uio"));
}

#[test]
fn test_expansion() {
    assert!(
        expand_patterns(&vec!["hi[1-2]*".to_string(), "ho[3-4]".to_string()])
            .unwrap()
            .eq(&vec![
                vec![(true, "hi1".to_string()), (false, "ho3".to_string())],
                vec![(true, "hi1".to_string()), (false, "ho4".to_string())],
                vec![(true, "hi2".to_string()), (false, "ho3".to_string())],
                vec![(true, "hi2".to_string()), (false, "ho4".to_string())]
            ])
    )
}

/// The logic here is that we want to construct a comma-separated list of host name patterns where a
/// pattern has the form /prefix[numlist]/ (the [] are literal), where numlist is a comma-separated
/// list of decimal numbers or ranges of numbers a-b.  This fits the typical host naming on a
/// supercomputer, which is <name>-<number> or <name><number>.  But note the <name> may also contain
/// digits.
///
/// The input is a set and combine_hosts is a function on that set: that is, inputs [x,y] and [y,x]
/// will yield the same output (and this is important for some consumers).

pub fn combine_hosts(hosts: Vec<String>) -> String {
    // Split into groups of names a.b.c.d whose tails .b.c.d are the same, we will attempt to merge
    // their `a` elements.
    let mut splits = hosts
        .iter()
        .map(|s| s.split(".").collect::<Vec<&str>>())
        .collect::<Vec<Vec<&str>>>();

    // Sort lexicographically by tail first and then hostname second - this will allow us to group
    // by tail in a single pass and later group by prefix in another pass.
    splits.sort_by(|a, b| {
        let mut i = 1;
        while i < a.len() || i < b.len() {
            if i < a.len() && i < b.len() {
                if a[i] != b[i] {
                    return a[i].cmp(&b[i]);
                }
            } else if i < a.len() {
                return std::cmp::Ordering::Greater;
            } else {
                return std::cmp::Ordering::Less;
            }
            i += 1;
        }
        return a[0].cmp(&b[0]);
    });

    let mut groups: Vec<&[Vec<&str>]> = vec![];
    let mut i = 0;
    while i < splits.len() {
        let mut j = i + 1;
        while j < splits.len() && same_strs(&splits[i][1..], &splits[j][1..]) {
            j = j + 1
        }
        groups.push(&splits[i..j]);
        i = j
    }

    let mut results: Vec<String> = vec![];
    for g in groups.drain(0..) {
        // Each g has a common tail `.b.c.d`.  Group all the `a` elements from group `g` that can be
        // combined.  The `a` elements that can be combined have a common prefix and a set of
        // combinable suffixes, and since we started with a sorted list, the combinable elements of
        // g are consecutive.

        let lim = g.len();
        let mut i = 0;
        while i < lim {
            let mut j = i + 1;
            let mut suffixes = vec![];
            let mut prefix = None;
            while j < lim {
                if let Some(ix) = combinable(g[i][0], g[j][0]) {
                    if suffixes.is_empty() {
                        prefix = Some(g[i][0][..ix].to_string());
                        suffixes.push(g[i][0][ix..].parse::<usize>().unwrap());
                    }
                    suffixes.push(g[j][0][ix..].parse::<usize>().unwrap());
                    j += 1;
                } else {
                    break;
                }
            }
            if suffixes.is_empty() {
                results.push(g[i].join("."));
            } else {
                // combine several
                let s = prefix.unwrap() + &combine(suffixes);
                let mut elts = vec![s.as_str()];
                elts.extend_from_slice(&g[i][1..]);
                results.push(elts.join("."));
            }
            i = j;
        }
    }

    results.join(",")
}

fn same_strs(a: &[&str], b: &[&str]) -> bool {
    if a.len() != b.len() {
        return false;
    }
    for i in 0..a.len() {
        if a[i] != b[i] {
            return false;
        }
    }
    return true;
}

// It is known that the prefixes can be combined.
fn combine(mut suffixes: Vec<usize>) -> String {
    suffixes.sort();
    let mut s = "[".to_string();
    let mut k = 0;
    while k < suffixes.len() {
        let mut m = k + 1;
        while m < suffixes.len() && suffixes[m] == suffixes[k] + (m - k) {
            m += 1;
        }
        if k > 0 {
            s += ",";
        }
        if m == k + 1 {
            s += &suffixes[k].to_string();
        } else {
            s += &format!("{}-{}", suffixes[k], suffixes[m - 1]);
        }
        k = m;
    }
    s + "]"
}

#[test]
fn test_combine_hosts() {
    assert!(
        combine_hosts(vec![
            "a1".to_string(),
            "a3".to_string(),
            "a2".to_string(),
            "a5".to_string()
        ]) == "a[1-3,5]".to_string()
    );
    // Hosts are carefully ordered here to ensure that they are not sorted either by their first or
    // second elements.
    assert!(
        combine_hosts(vec![
            "a3.fox".to_string(),
            "a1.fox".to_string(),
            "a3.fum".to_string(),
            "a2.fox".to_string(),
            "a5.fox".to_string(),
        ]) == "a[1-3,5].fox,a3.fum".to_string()
    );
}

// Names can be merged if they both end with a digit string and there is a joint prefix before the
// digit string.  For now, we require this prefix to not end with a digit.  This returns None for
// "no" and Some(usize) for "yes" where usize is the byte index of the start of the digit string.
//
// Combinability must be reflexive, symmetric, and transitive.

fn combinable(a: &str, b: &str) -> Option<usize> {
    let xs = a.as_bytes();
    let mut i = xs.len();
    while i > 0 && xs[i - 1] >= b'0' && xs[i - 1] <= b'9' {
        i -= 1;
    }
    if i == 0 || i == xs.len() {
        return None;
    }
    let ys = b.as_bytes();
    let mut j = ys.len();
    while j > 0 && ys[j - 1] >= b'0' && ys[j - 1] <= b'9' {
        j -= 1;
    }
    if j == 0 || j == ys.len() {
        return None;
    }
    if i != j {
        return None;
    }
    if xs[..i] != ys[..j] {
        return None;
    }
    return Some(i);
}

#[test]
fn test_elements_can_be_merged() {
    assert!(combinable("", "") == None);
    assert!(combinable("a", "b") == None);
    assert!(combinable("a", "a") == None);
    assert!(combinable("a1", "a23") == Some(1));
    assert!(combinable("a1-1", "a1-23") == Some(3));
}
