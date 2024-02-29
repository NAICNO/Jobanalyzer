/// The GpuSet has three states:
///
///  - the set is known to be empty, this is Some({})
///  - the set is known to be nonempty and have only known gpus in the set, this is Some({a,b,..})
///  - the set is known to be nonempty but have (some) unknown members, this is None
///
/// During processing, the set starts out as Some({}).  If a device reports "unknown" GPUs then the
/// set can transition from Some({}) to None or from Some({a,b,..}) to None.  Once in the None
/// state, the set will stay in that state.  There is no representation for some known + some
/// unknown GPUs, it is not believed to be worthwhile.
///
/// To conserve space in the LogEntry we use a bitmap for cards rather than a HashSet.  The HashSet
/// is quite large.
use std::str::FromStr;

pub type GpuSet = Option<u32>;

pub fn empty_gpuset() -> GpuSet {
    Some(0)
}

pub fn is_empty_gpuset(s: &GpuSet) -> bool {
    if let Some(set) = s {
        *set == 0
    } else {
        false
    }
}

pub fn unknown_gpuset() -> GpuSet {
    None
}

pub fn is_unknown_gpuset(s: &GpuSet) -> bool {
    s.is_none()
}

pub fn singleton_gpuset(maybe_device: Option<u32>) -> GpuSet {
    if let Some(dev) = maybe_device {
        assert!(dev < 32);
        Some(1 << dev)
    } else {
        None
    }
}

pub fn adjoin_gpuset(lhs: &mut GpuSet, rhs: u32) {
    if let Some(gpus) = lhs {
        assert!(rhs < 32);
        *gpus |= 1 << rhs;
    }
}

pub fn union_gpuset(lhs: &mut GpuSet, rhs: &GpuSet) {
    if lhs.is_none() {
        // The result is also None
    } else if rhs.is_none() {
        *lhs = None;
    } else {
        *lhs.as_mut().unwrap() |= rhs.as_ref().unwrap();
    }
}

// For testing, we need a predictable order, so accumulate as numbers and sort
pub fn gpuset_to_string(gpus: &GpuSet) -> String {
    if let Some(gpus) = gpus {
        if *gpus == 0 {
            "none".to_string()
        } else {
            let mut cards = vec![];
            let mut g = *gpus;
            let mut i = 0;
            while g != 0 {
                if (g & 1) == 1 {
                    cards.push(i);
                }
                g >>= 1;
                i += 1;
            }
            cards.sort();
            let mut term = "";
            let mut s = String::new();
            for x in cards {
                s += term;
                term = ",";
                s += &x.to_string();
            }
            s
        }
    } else {
        "unknown".to_string()
    }
}

// The bool return value is "failed".

pub fn gpuset_from_bitvector(s: &str) -> (Option<GpuSet>, bool) {
    match u32::from_str_radix(s, 2) {
        Ok(bit_mask) => (Some(Some(bit_mask)), false),
        Err(_) => (None, true),
    }
}

// The bool return value is "failed".

pub fn gpuset_from_list(s: &str) -> (Option<GpuSet>, bool) {
    if s == "unknown" {
        (Some(unknown_gpuset()), false)
    } else if s == "none" {
        (Some(empty_gpuset()), false)
    } else {
        let mut set = empty_gpuset();
        let vs: std::result::Result<Vec<_>, _> = s.split(',').map(u32::from_str).collect();
        match vs {
            Err(_) => (None, true),
            Ok(vs) => {
                for v in vs {
                    adjoin_gpuset(&mut set, v);
                }
                (Some(set), false)
            }
        }
    }
}

#[test]
fn test_gpuset() {
    assert!(is_empty_gpuset(&empty_gpuset()));
    assert!(!is_empty_gpuset(&unknown_gpuset()));
    assert!(!is_empty_gpuset(&singleton_gpuset(Some(1))));
    let mut s = unknown_gpuset();
    adjoin_gpuset(&mut s, 1);
    assert!(is_unknown_gpuset(&s));
}

#[test]
fn test_get_gpus_from_list() {
    // Much more could be done here
    assert!(gpuset_from_list("unknownx") == (None, true));
    assert!(gpuset_from_list("unknown") == (Some(unknown_gpuset()), false));
    assert!(gpuset_from_list("none") == (Some(empty_gpuset()), false));
    assert!(gpuset_from_list("1") == (Some(singleton_gpuset(Some(1))), false));
    assert!(gpuset_from_list("1,1,1") == (Some(singleton_gpuset(Some(1))), false));
    let mut s1 = singleton_gpuset(Some(1));
    adjoin_gpuset(&mut s1, 2);
    assert!(gpuset_from_list("2,1") == (Some(s1), false));
    let mut s2 = unknown_gpuset();
    adjoin_gpuset(&mut s2, 1);
    assert!(s2 == unknown_gpuset());
}
// Other test cases are black-box, see ../../tests/sonarlog
