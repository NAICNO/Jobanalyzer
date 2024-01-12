package main

/*
#[cfg(test)]
fn mkusers() -> UserTable<'static> {
    map! {
        447153 => ("bob", 1001),
        447160 => ("bob", 1001),
        1864615 => ("alice", 1002),
        2233095 => ("charlie", 1003),
        2233469 => ("charlie", 1003)
    }
}

#[cfg(test)]
pub fn parsed_pmon_output() -> Vec<Process> {
    let text = "# gpu        pid  type    sm   mem   enc   dec   command
# Idx          #   C/G     %     %     %     %   name
# gpu        pid  type    fb    sm   mem   enc   dec   command
# Idx          #   C/G    MB     %     %     %     %   name
    0     447153     C  7669     -     -     -     -   python3.9
    0     447160     C 11057     -     -     -     -   python3.9
    0     506826     C 11057     -     -     -     -   python3.9
    0    1864615     C  1635    40     0     -     -   python
    1    1864615     C   535     -     -     -     -   python
    1    2233095     C 24395    84    23     -     -   python3
    2    1864615     C   535     -     -     -     -   python
    2    1448150     C  9383     -     -     -     -   python3
    3    1864615     C   535     -     -     -     -   python
    3    2233469     C 15771    90    23     -     -   python3
";
    parse_pmon_output(text, &mkusers())
}

#[cfg(test)]
macro_rules! proc(
    { $a:expr, $b:expr, $c:expr, $d:expr, $e: expr, $f:expr, $g:expr, $h:expr } => {
	Process { device: $a,
		  pid: $b,
		  user: $c.to_string(),
                  uid: $d,
		  gpu_pct: $e,
		  mem_pct: $f,
		  mem_size_kib: $g,
		  command: $h.to_string()
	}
    });

#[test]
fn test_parse_pmon_output() {
    assert!(parsed_pmon_output().into_iter().eq(vec![
        proc! { Some(0),  447153, "bob",            1001, 0.0,  0.0,  7669 * 1024, "python3.9" },
        proc! { Some(0),  447160, "bob",            1001, 0.0,  0.0, 11057 * 1024, "python3.9" },
        proc! { Some(0),  506826, "_zombie_506826", ZOMBIE_UID, 0.0,  0.0, 11057 * 1024, "python3.9" },
        proc! { Some(0), 1864615, "alice",          1002, 40.0,  0.0,  1635 * 1024, "python" },
        proc! { Some(1), 1864615, "alice",          1002,  0.0,  0.0,   535 * 1024, "python" },
        proc! { Some(1), 2233095, "charlie",        1003, 84.0, 23.0, 24395 * 1024, "python3" },
        proc! { Some(2), 1864615, "alice",          1002, 0.0,  0.0,   535 * 1024, "python" },
        proc! { Some(2), 1448150, "_zombie_1448150", ZOMBIE_UID, 0.0,  0.0,  9383 * 1024, "python3"},
        proc! { Some(3), 1864615, "alice",          1002,  0.0,  0.0,   535 * 1024, "python" },
        proc! { Some(3), 2233469, "charlie",        1003, 90.0, 23.0, 15771 * 1024, "python3" }
    ]))
}

#[cfg(test)]
pub fn parsed_query_output() -> Vec<Process> {
    let text = "2233095, 1190
3079002, 2350
1864615, 1426";
    parse_query_output(text, &mkusers())
}

#[test]
fn test_parse_query_output() {
    assert!(parsed_query_output().into_iter().eq(vec![
        proc! { None, 3079002, "_zombie_3079002", ZOMBIE_UID, 0.0, 0.0, 2350 * 1024, "_unknown_" }
    ]))
}
*/
