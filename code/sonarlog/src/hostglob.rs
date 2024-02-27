/// Matcher, expander and compressor for host names.
///
/// This closely follows the grammar and semantics defined in the corresponding Go and JS code, see
/// go-utils/hostglob/hostglob.go and dashboard/hostglob.js.  They must all be kept in sync.

use crate::pattern;

use anyhow::{bail, Result};
use ustr::Ustr;
use regex::Regex;

/// A `HostGlobber` is a matcher of patterns against hostnames.
///
/// The matcher holds a number of patterns, added with `insert`.  Each `pattern` is a <pattern> in
/// the sense of the grammar referenced above.
///
/// The `match_hostname` method attempts to match its argument against the patterns in the matcher,
/// returning true if any of them match.

pub struct HostGlobber {
    // If true, then the patterns are all constructed to match a prefix of a <hostname>.
    is_prefix_matcher: bool,

    // Matcher + source pattern, for posterity.
    matchers: Vec<(regex::Regex, String)>,
}

impl HostGlobber {
    /// Create a new, empty filter.

    pub fn new(is_prefix_matcher: bool) -> HostGlobber {
        HostGlobber {
            is_prefix_matcher,
            matchers: vec![],
        }
    }

    /// Add the pattern to the set of patterns in the matcher.

    pub fn insert(&mut self, pattern: &str) -> Result<()> {
        self.matchers.push(compile_globber(pattern, self.is_prefix_matcher)?);
        Ok(())
    }

    /// Return true iff the filter has no patterns.

    pub fn is_empty(&self) -> bool {
        self.matchers.len() == 0
    }

    /// Match s against the patterns and return true iff it matches at least one pattern.

    pub fn match_hostname(&self, s: &str) -> bool {
        for m in &self.matchers {
            if m.0.is_match(s) {
                return true
            }
        }
        return false
    }
}

fn compile_globber(p: &str, prefix: bool) -> Result<(Regex, String)> {
    let cs = p.chars().collect::<Vec<char>>();
    let mut i = 0usize;
    let mut r = "^".to_string();
    while i < cs.len() {
        if r.len() > 50000 {
            bail!("Expression too large, use more '*'")
        }
        match cs[i] {
            '*' => {
                i += 1;
                r += "[^.]*";
            }
            '[' => {
                i += 1;
                let mut set = vec![];
                loop {
                    let mut n0;
                    (n0, i) = read_int(&cs, i)?;
                    if i < cs.len() && cs[i] == '-' {
                        i += 1;
                        let n1;
                        (n1, i) = read_int(&cs, i)?;
                        if n0 > n1 {
                            bail!("Invalid range");
                        }
                        while n0 <= n1 {
                            set.push(n0.to_string());
                            n0 += 1;
                            if set.len() > 10000 {
                                bail!("Range too large, use more '*'");
                            }
                        }
                    } else {
                        set.push(n0.to_string());
                    }
                    if i < cs.len() && cs[i] == ']' {
                        i += 1;
                        break;
                    }
                    if i >= cs.len() || cs[i] != ',' {
                        bail!("Expected ','");
                    }
                    i += 1;
                }
                r += "(?:";
                r += &set.join("|");
                r += ")";
            }
            '.' | '$' | '^' | '?' | '\\' | '(' | ')' | '{' | '}' | ']' => {
                // Mostly these special chars are not allowed in hostnames, but it doesn't hurt to
                // try to avoid disasters.
                r += "\\";
                r.push(cs[i]);
                i += 1;
            }
            _ => {
                r.push(cs[i]);
                i += 1;
            }
        }
    }
    if prefix {
        // Matching a prefix: we need to match entire host-elements, so following a prefix there
        // should either be EOS or `.` followed by whatever until EOS.
        r += "(?:\\..*)?$"
    } else {
        r += "$";
    }
    Ok((regex::Regex::new(&r)?, r))
}

fn read_int(cs: &[char], mut i: usize) -> Result<(usize, usize)> {
    let first = i;
    let mut n = 0u64;
    while i < cs.len() && cs[i].is_digit(10) {
        n = n*10 + (u32::from(cs[i]) - 48) as u64;
        if n > 0xFFFFFFFF {
            bail!("Number out of range in glob set");
        }
        i += 1;
    }
    if i == first {
        bail!("Invalid number in glob set");
    }
    Ok((n as usize, i))
}

#[test]
fn test_hostfilter1() {
    let mut hf = HostGlobber::new(true);
    hf.insert("ml8").unwrap();
    hf.insert("ml3.hpc").unwrap();

    // Single-element prefix match against this
    assert!(hf.match_hostname("ml8.hpc.uio.no"));

    // Multi-element prefix match against this
    assert!(hf.match_hostname("ml3.hpc.uio.no"));

    let mut hf = HostGlobber::new(false);
    hf.insert("ml4.hpc.uio.no").unwrap();

    // Exhaustive match against this
    assert!(hf.match_hostname("ml4.hpc.uio.no"));
    assert!(!hf.match_hostname("ml4.hpc.uio.no.yes"));
}

#[test]
fn test_hostfilter2() {
    let mut hf = HostGlobber::new(true);
    hf.insert("ml[1-3]*").unwrap();
    assert!(hf.match_hostname("ml1"));
    assert!(hf.match_hostname("ml1x"));
    assert!(hf.match_hostname("ml1.uio"));
}

#[test]
fn test_hostfilter3() {
    let mut hf = HostGlobber::new(false);
    hf.insert("c[1-3]-[2,4]").unwrap();
    assert!(hf.match_hostname("c1-2"));
    assert!(hf.match_hostname("c2-2"));
    assert!(!hf.match_hostname("c2-3"));
}

// TODO: Less trivial test cases

/// `expand_pattern()` takes a single <pattern> from the grammar and expands it into a set of host
/// names without wildcards or ranges.
///
/// Restriction: The pattern must contain no "*" wildcards.

pub fn expand_pattern(p: &str) -> Result<Vec<String>> {
    let elements = p
        .split('.')
        .map(|x| x.to_string())
        .collect::<Vec<String>>();
    let expansions = expand_elements(&elements)?;
    let mut result = vec![];
    for mut exps in expansions {
        if exps.iter().any(|(prefix, _)| *prefix) {
            bail!("Suffix wildcard not allowed in expandable hostname in config file")
        }
        result.push(
            exps.drain(0..)
                .map(|(_, elt)| elt)
                .collect::<Vec<String>>()
                .join("."),
        )
    }
    Ok(result)
}

// TODO: Test cases

fn expand_elements(xs: &[String]) -> Result<Vec<Vec<(bool, String)>>> {
    if xs.len() == 0 {
        Ok(vec![vec![]])
    } else {
        let rest = expand_elements(&xs[1..])?;
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
fn test_expansion() {
    assert!(
        expand_elements(&vec!["hi[1-2]*".to_string(), "ho[3-4]".to_string()])
            .unwrap()
            .eq(&vec![
                vec![(true, "hi1".to_string()), (false, "ho3".to_string())],
                vec![(true, "hi1".to_string()), (false, "ho4".to_string())],
                vec![(true, "hi2".to_string()), (false, "ho3".to_string())],
                vec![(true, "hi2".to_string()), (false, "ho4".to_string())]
            ])
    )
}

/// `compress_hostnames()` takes a list of <hostname> and returns a vector of <pattern> s.t. the
/// expansion of those patterns results in the original input list.

pub fn compress_hostnames(hosts: &[Ustr]) -> Vec<String> {
    // Given names a1.b.c.d and a2.b.c.d this will return a[1,2].b.c.d, ie, it compresses numbers at
    // the end of the first <host-element> for equal prefixes of those elements and equal tails of
    // elements.  This fits the typical host naming on a supercomputer, which is <name>-<number> or
    // <name><number>.  Note the <name> may also contain digits in the former case.
    //
    // TODO: This code is from an earlier iteration of the problem.  It probably wants to use the
    // same logic as the Go code, which can also compress "a1b" and "a2b" into "a[1,2]b".  This code
    // is still correct, though, and effective in most cases.

    // Split into groups of names a.b.c.d whose tails .b.c.d are the same, we will attempt to merge
    // their `a` elements.
    let mut splits = hosts
        .iter()
        .map(|s| s.as_str().split(".").collect::<Vec<&str>>())
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

    results
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
fn test_compress_hostnames() {
    assert!(
        compress_hostnames(&vec![
            Ustr::from("a1"),
            Ustr::from("a3"),
            Ustr::from("a2"),
            Ustr::from("a5")
        ]).join(",") == "a[1-3,5]"
    );
    // Hosts are carefully ordered here to ensure that they are not sorted either by their first or
    // second elements.
    assert!(
        compress_hostnames(&vec![
            Ustr::from("a3.fox"),
            Ustr::from("a1.fox"),
            Ustr::from("a3.fum"),
            Ustr::from("a2.fox"),
            Ustr::from("a5.fox"),
        ]).join(",") == "a[1-3,5].fox,a3.fum"
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
