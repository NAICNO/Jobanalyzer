// `lsjobs` -- process Sonar log files and list jobs, with optional filtering and details.
//
// See MANUAL.md for a manual, or run with --help for brief help.

// TODO - High pri
//
// (Nothing)
//
//
// TODO - Normal pri
//
// There's a fairly benign bug below in how earliest and latest are computed.
//
// Hostname filtering (beyond FQDN matching) in logtree.md.
//
// Figure out how to show hosts / node names for a job.  (This is something that only matters when
// integrating with SLURM or other job queues, it can't be tested on the ML or light-HPC nodes.  So
// test on Fox.)  I think maybe an option --show-hosts would be appropriate, and in this case the
// list of hosts would be printed after the command?  Or instead of the command?
//
//
// TODO - Backlog / discussion
//
// Bug: For zombies, the "user name" can be longer than 8 chars and may need to be truncated or
// somehow managed, I think.  It's possible it shouldn't be printed if --zombie, but that's not
// the only case.
//
// Feature ("manual monitoring" use case): Figure out how to show load.
//
//   Definition: the "load at time t on a host" is the sum across all jobs at time t of
//   cpu/gpu/mem/vmem, with the same meanings as those fields have.  (This can then be related to
//   the configuration of that host but that's for later.)
//
//   This presupposes that the sonar log uses the same time stamp for all records captured at a
//   given time (it currently does this) or that we establish a time window for observations that
//   are to be summed.  For now, there's no reason to establish such a time window.
//
//   The "historical load" of a host is then a table of the load at times through history, computed
//   every time sonar has a sample for the host.
//
//   There is a complication if we want the *printed* historical load to be extracted from the full
//   table; for example, if we sample every five minutes but want to print the load hourly.  In this
//   case, some kind of average of the load values over a time period would be the printed load.
//
//   Thus we have --load=<something> which specifies how to compute and display the load.  This
//   implies --user=- instead of --user=$LOGNAME (if not specified) and requires --host=<hostname> (but why?).
//
//   The <something> specifies what to print: `last` implies the last sample time; `all` is the
//   full log for the time window; `hourly` and `daily` are averages within the time window.
//
// Feature: One could imagine other sort orders for the output than least-recently-started-first.
// This only matters for the --numjobs switch.
//
// Tweak: A number of minor TODO items in logtree.rs when accessing a directory or file fails.
//
// Tweak: We allow for at most a two-digit number of days of running time in the output but in
// practice we're going to see some three-digit number of days, make room for that.
//
// Perf: Performance and memory use will become an issue with a large number of records?  Probably want to
// profile before we hack too much, but there are obvious inefficiencies in representations and the
// number of passes across data structures, and probably in the number of copies made (and thus the
// amount of memory allocation).
//
// Testing: Selftest cases everywhere, but esp for the argument parsers and filterers.
//
// Structure: Maybe refactor the argument processing into a separate file, it's becoming complex
// enough.  Wait until output filtering logic is in order.
//
//
//
// Quirks
//
// Having the absence of --user mean "only $LOGNAME" can be confusing -- though it's the right thing
// for a use case where somebody is looking only at her own jobs.
//
// The --from and --to values are used *both* for filtering files in the directory tree of logs
// (where it is used to generate directory names to search) *and* for filtering individual records
// in the log files.  Things can become a confusing if the log records do not have dates
// corresponding to the directories they are located in.  This is mostly a concern for testing.
//
// Some filtering options select *records* (from, to, host, user, exclude) and some select *jobs*
// (the rest of them), and this can be confusing.  For user and exclude this does not matter (modulo
// setuid or similar personality changes).  The user might expect that from/to/host would select
// jobs instead of records, s.t. if a job ran in the time interval (had samples in the interval)
// then the entire job should be displayed, including data about it outside the interval.  Ditto,
// that if a job ran on a selected host then its work on all hosts should be displayed.  But it just
// ain't so.

mod logfile;
mod logtree;
mod dates;

use chrono::prelude::{DateTime,NaiveDate};
use chrono::Utc;
use clap::Parser;
use core::cmp::{min,max};
use std::cell::RefCell;
use std::collections::{HashSet,HashMap};
use std::env;
use std::num::ParseIntError;
use std::ops::Add;
use std::process;
use std::str::FromStr;
use std::time;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    /// Select the root directory for log files [default: $SONAR_ROOT]
    #[arg(long)]
    data_path: Option<String>,

    /// Select these user name(s), comma-separated, "-" for all [default: $LOGNAME]
    #[arg(long, short)]
    user: Option<String>,

    /// Exclude these user name(s), comma-separated [default: none]
    #[arg(long)]
    exclude: Option<String>,

    /// Select these job number(s), comma-separated [default: all]
    #[arg(long, short, value_parser = job_numbers)]
    job: Option<Vec<usize>>,
    
    /// Select only jobs with this command name (case-sensitive substring) [default: all]
    #[arg(long)]
    command: Option<String>,

    /// Select records by this time and later.  Format can be YYYY-MM-DD, or Nd or Nw
    /// signifying N days or weeks ago [default: 1d, ie 1 day ago]
    #[arg(long, short, value_parser = parse_time)]
    from: Option<DateTime<Utc>>,

    /// Select records by this time and earlier.  Format can be YYYY-MM-DD, or Nd or Nw
    /// signifying N days or weeks ago [default: now]
    #[arg(long, short, value_parser = parse_time)]
    to: Option<DateTime<Utc>>,

    /// Select records and logs by these host name(s), comma-separated [default: all]
    #[arg(long)]
    host: Option<String>,

    /// Select only jobs with at least this many observations
    #[arg(long, default_value_t = 2)]
    min_observations: usize,

    /// Select only jobs with at least this much average CPU use (100=1 full CPU)
    #[arg(long, default_value_t = 0)]
    min_avg_cpu: usize,

    /// Select only jobs with at least this much peak CPU use (100=1 full CPU)
    #[arg(long, default_value_t = 0)]
    min_peak_cpu: usize,

    /// Select only jobs with at least this much average main memory use (GB)
    #[arg(long, default_value_t = 0)]
    min_avg_mem: usize,

    /// Select only jobs with at least this much peak main memory use (GB)
    #[arg(long, default_value_t = 0)]
    min_peak_mem: usize, 

    /// Select only jobs with at least this much average GPU use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_avg_gpu: usize, 

    /// Select only jobs with at least this much peak GPU use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_peak_gpu: usize, 

    /// Select only jobs with at least this much average GPU memory use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_avg_vmem: usize, 

    /// Select only jobs with at least this much peak GPU memory use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_peak_vmem: usize, 

    /// Select only jobs with at least this much runtime, format `WwDdHhMm`, all parts optional [default: 0m]
    #[arg(long, value_parser = run_time)]
    min_runtime: Option<chrono::Duration>,

    /// Select only jobs with no GPU use
    #[arg(long, default_value_t = false)]
    no_gpu: bool,

    /// Select only jobs with some GPU use
    #[arg(long, default_value_t = false)]
    some_gpu: bool,

    /// Select only jobs that have run to completion
    #[arg(long, default_value_t = false)]
    completed: bool,

    /// Select only jobs that are still running
    #[arg(long, default_value_t = false)]
    running: bool,

    /// Select only zombie jobs (usually these are still running)
    #[arg(long, default_value_t = false)]
    zombie: bool,

    /// Print system load instead of jobs, argument is `last`,`hourly`,`daily` [default: none]
    #[arg(long)]
    load: Option<String>,

    /// Print at most these many most recent jobs per user [default: all]
    #[arg(long, short)]
    numjobs: Option<usize>,

    /// Print useful(?) statistics about the input and output
    #[arg(long, short, default_value_t = false)]
    verbose: bool,
    
    /// Print unformatted data (for developers)
    #[arg(long, default_value_t = false)]
    raw: bool,

    /// Log file names (overrides --data-path)
    #[arg(last = true)]
    logfiles: Vec<String>,
}

// Comma-separated job numbers.
fn job_numbers(s: &str) -> Result<Vec<usize>, String> {
    let candidates = s.split(',').map(|x| usize::from_str(x)).collect::<Vec<Result<usize, ParseIntError>>>();
    if candidates.iter().all(|x| x.is_ok()) {
        Ok(candidates.iter().map(|x| *x.as_ref().unwrap()).collect::<Vec<usize>>())
    } else {
        Err("Illegal job numbers: ".to_string() + s)
    }
}

// YYYY-MM-DD, but with a little (too much?) flexibility.  Or Nd, Nw.
fn parse_time(s: &str) -> Result<DateTime<Utc>, String> {
    if let Some(n) = s.strip_suffix('d') {
        if let Ok(k) = usize::from_str(n) {
            Ok(now() - chrono::Duration::days(k as i64))
        } else {
            Err(format!("Invalid date: {}", s))
        }
    } else if let Some(n) = s.strip_suffix('w') {
        if let Ok(k) = usize::from_str(n) {
            Ok(now() - chrono::Duration::weeks(k as i64))
        } else {
            Err(format!("Invalid date: {}", s))
        }
    } else {
        let parts = s.split('-').map(|x| usize::from_str(x)).collect::<Vec<Result<usize, ParseIntError>>>();
        if !parts.iter().all(|x| x.is_ok()) || parts.len() != 3 {
            return Err(format!("Invalid date syntax: {}", s));
        }
        let vals = parts.iter().map(|x| *x.as_ref().unwrap()).collect::<Vec<usize>>();
        let d = NaiveDate::from_ymd_opt(vals[0] as i32, vals[1] as u32, vals[2] as u32);
        if !d.is_some() {
            return Err(format!("Invalid date: {}", s));
        }
        Ok(DateTime::from_utc(d.unwrap().and_hms_opt(0,0,0).unwrap(), Utc))
    }
}

// This is DdHhMm with all parts optional but at least one part required.  There is possibly too
// much flexibility here, as the parts can be in any order.
fn run_time(s: &str) -> Result<chrono::Duration, String> {
    let bad = format!("Bad time duration syntax: {}", s);
    let mut weeks = 0u64;
    let mut days = 0u64;
    let mut hours = 0u64;
    let mut minutes = 0u64;
    let mut have_weeks = false;
    let mut have_days = false;
    let mut have_hours = false;
    let mut have_minutes = false;
    let mut ds = "".to_string();
    for ch in s.chars() {
        if ch.is_digit(10) {
            ds = ds + &ch.to_string();
        } else {
            if ds == "" ||
                (ch != 'd' && ch != 'h' && ch != 'm' && ch != 'w') ||
                (ch == 'd' && have_days) || (ch == 'h' && have_hours) || (ch == 'm' && have_minutes) || (ch == 'w' && have_weeks) {
                    return Err(bad)
                }
            let v = u64::from_str(&ds);
            if !v.is_ok() {
                return Err(bad);
            }
            let val = v.unwrap();
            ds = "".to_string();
            if ch == 'd' {
                have_days = true;
                days = val;
            } else if ch == 'h' {
                have_hours = true;
                hours = val;
            } else if ch == 'm' {
                have_minutes = true;
                minutes = val;
            } else if ch == 'w' {
                have_weeks = true;
                weeks = val;
            }
        }
    }
    if ds != "" || (!have_days && !have_hours && !have_minutes && !have_weeks) {
        return Err(bad);
    }

    days += weeks * 7;
    hours += days * 24;
    minutes += hours * 60;
    let seconds = minutes * 60;
    match chrono::Duration::from_std(time::Duration::from_secs(seconds)) {
        Ok(e) => Ok(e),
        Err(_) => Err("Bad running time".to_string())
    }
}

fn main() {
    let mut cli = Cli::parse();

    // Perform some ad-hoc validation.

    if let Some(ref l) = cli.load {
        if !cli.host.is_some() {
            fail("--load requires --host to select host(s)");
        }
        match l.as_str() {
            "last" | "hourly" | "daily" => {},
            _ => fail("--load requires a value `last`, `hourly`, `daily`")
        }
    }

    // Figure out the data path from switches and defaults.

    let data_path = if cli.data_path.is_some() {
        cli.data_path.clone()
    } else if let Ok(val) = env::var("SONAR_ROOT") {
        Some(val)
    } else if let Ok(val) = env::var("HOME") {
        Some(val + "/sonar_logs")
    } else {
        None
    };

    // Convert the input filtering options to a useful form.

    let from = if let Some(x) = cli.from { x } else { one_day_ago() };
    let to = if let Some(x) = cli.to { x } else { now() };
    if from > to {
        fail("The --from time is greater than the --to time");
    }

    let include_hosts = if let Some(ref hosts) = cli.host {
        hosts.split(',').map(|x| x.to_string()).collect::<HashSet<String>>()
    } else {
        HashSet::new()
    };

    let include_jobs = if let Some(ref jobs) = cli.job {
        jobs.iter().map(|x| *x).collect::<HashSet<usize>>()
    } else {
        HashSet::new()
    };

    let include_users = if let Some(ref users) = cli.user {
        if users == "-" {
            HashSet::new()
        } else {
            users.split(',').map(|x| x.to_string()).collect::<HashSet<String>>()
        }
    } else if cli.zombie || cli.load.is_some() {
        HashSet::new()
    } else {
        let mut users = HashSet::new();
        if let Ok(u) = env::var("LOGNAME") {
            users.insert(u);
        };
        users
    };

    let mut exclude_users = if let Some(ref excl) = cli.exclude {
        excl.split(',').map(|x| x.to_string()).collect::<HashSet<String>>()
    } else {
        HashSet::new()
    };
    exclude_users.insert("root".to_string());
    exclude_users.insert("zabbix".to_string());

    // The input filter.

    let record_counter = RefCell::new(0usize);
    let filter = |user:&str, host:&str, job: u32, t:&DateTime<Utc>| {
        *record_counter.borrow_mut() += 1;
        ((&include_users).is_empty() || (&include_users).contains(user)) &&
        ((&include_hosts).is_empty() || (&include_hosts).contains(host)) &&
        ((&include_jobs).is_empty() || (&include_jobs).contains(&(job as usize))) &&
            !(&exclude_users).contains(user) &&
            from <= *t &&
            *t <= to
    };

    // Logfiles, filtered by host and time range.

    let logfiles =
        if cli.logfiles.len() > 0 {
            cli.logfiles.split_off(0)
        } else {
            if cli.verbose {
                eprintln!("Data path: {:?}", data_path);
            }
            let maybe_logfiles = logtree::find_logfiles(data_path, &include_hosts, from, to);
            if let Err(ref msg) = maybe_logfiles {
                fail(&format!("{}", msg));
            }
            maybe_logfiles.unwrap()
        };

    if cli.verbose {
        eprintln!("Log files: {:?}", logfiles);
    }

    // Read the files, filter the records, build up a set of candidate log records.

    let mut joblog = HashMap::<u32, Vec<logfile::LogEntry>>::new();
    logfiles.iter().for_each(|file| {
        match logfile::parse(file, &filter) {
            Ok(mut log_entries) => {
                for entry in log_entries.drain(0..) {
                    if let Some(job) = joblog.get_mut(&entry.job_id) {
                        job.push(entry);
                    } else {
                        joblog.insert(entry.job_id, vec![entry]);
                    }
                }
            }
            Err(e) => {
                eprintln!("ERROR: {:?}", e);
                return;
            }
        }
    });

    if cli.verbose {
        eprintln!("Number of job records read: {}", *record_counter.borrow());
        eprintln!("Number of job records after input filtering: {}", joblog.len());
    }

    if let Some(_l) = cli.load {
        // TODO: Compute loads, l is the specifier
    } else {
        aggregate_and_print_jobs(cli, joblog);
    }
}

fn aggregate_and_print_jobs(cli: Cli, mut joblog: HashMap::<u32, Vec<logfile::LogEntry>>) {
    // The `joblog` is a map from job ID to a vector of all job records with that job ID. Sort each
    // vector by ascending timestamp to get an idea of the duration of the job.
    //
    // TODO: We currenly only care about the max and min timestamps per job, so optimize later if
    // that doesn't change.
    //
    // (I have no idea what `&mut ref mut` means.)
    joblog.iter_mut().for_each(|(_k, &mut ref mut job)| {
        job.sort_by_key(|j| j.timestamp);
    });

    // Compute the earliest and latest times observed across all the logs
    //
    // FIXME: This is wrong!  It considers only included records, thus leading to incorrect marks being computed.
    // To do better, the log reader must compute these values, or we compute it in the filter function.
    let (earliest, latest) = {
        let max_start = epoch();
        let min_start = now();
        joblog.iter().fold((min_start, max_start),
                           |(earliest, latest), (_k, r)| (min(earliest, r[0].timestamp), max(latest, r[r.len()-1].timestamp)))
    };

    // Convert the aggregation filter options to a useful form.

    let min_avg_cpu = cli.min_avg_cpu as f64;
    let min_peak_cpu = cli.min_peak_cpu as f64;
    let min_avg_mem = cli.min_avg_mem;
    let min_peak_mem = cli.min_peak_mem;
    let min_avg_gpu = cli.min_avg_gpu as f64;
    let min_peak_gpu = cli.min_peak_gpu as f64;
    let min_runtime = if let Some(n) = cli.min_runtime { n.num_seconds() } else { 0 };
    let min_avg_vmem = cli.min_avg_vmem as f64;
    let min_peak_vmem = cli.min_peak_vmem as f64;

    // Get the vectors of jobs back into a vector, aggregate data, and filter the jobs.
    const LIVE_AT_END : u32 = 1;
    const LIVE_AT_START : u32 = 2;

    #[derive(Debug)]
    struct Aggregate {
        first: DateTime<Utc>,
        last: DateTime<Utc>,
        duration: i64,
        minutes: i64,
        hours: i64,
        days: i64,
        uses_gpu: bool,
        avg_cpu: f64,
        peak_cpu: f64,
        avg_gpu: f64,
        peak_gpu: f64,
        avg_mem_gb: f64,
        peak_mem_gb: f64,
        avg_vmem_pct: f64,
        peak_vmem_pct: f64,
        selected: bool,
        classification: u32,
    }

    let mut jobvec = joblog
        .drain()
        .filter(|(_, job)| job.len() >= cli.min_observations)
        .map(|(_, job)| {
            let first = job[0].timestamp;
            let last = job[job.len()-1].timestamp;
            let duration = (last - first).num_seconds();
            let minutes = duration / 60;
            let mut classification = 0;
            if first == earliest {
                classification |= LIVE_AT_START;
            }
            if last == latest {
                classification |= LIVE_AT_END;
            }
            (Aggregate {
                first,
                last,
                duration: duration,                     // total number of seconds
                minutes: minutes % 60,                  // fractional hours
                hours: (minutes / 60) % 24,             // fractional days
                days: minutes / (60 * 24),              // full days
                uses_gpu: job.iter().any(|jr| jr.gpu_mask != 0),
                avg_cpu: (job.iter().fold(0.0, |acc, jr| acc + jr.cpu_pct) / (job.len() as f64) * 100.0).ceil(),
                peak_cpu: (job.iter().map(|jr| jr.cpu_pct).reduce(f64::max).unwrap() * 100.0).ceil(),
                avg_gpu: (job.iter().fold(0.0, |acc, jr| acc + jr.gpu_pct) / (job.len() as f64) * 100.0).ceil(),
                peak_gpu: (job.iter().map(|jr| jr.gpu_pct).reduce(f64::max).unwrap() * 100.0).ceil(),
                avg_mem_gb: (job.iter().fold(0.0, |acc, jr| acc + jr.mem_gb) /  (job.len() as f64)).ceil(),
                peak_mem_gb: (job.iter().map(|jr| jr.mem_gb).reduce(f64::max).unwrap()).ceil(),
                avg_vmem_pct: (job.iter().fold(0.0, |acc, jr| acc + jr.gpu_mem_pct) /  (job.len() as f64) * 100.0).ceil(),
                peak_vmem_pct: (job.iter().map(|jr| jr.gpu_mem_pct).reduce(f64::max).unwrap() * 100.0).ceil(),
                selected: true,
                classification,
             },
             job)
        })
        .filter(|(aggregate, job)| {
            aggregate.avg_cpu >= min_avg_cpu &&
                aggregate.peak_cpu >= min_peak_cpu &&
                aggregate.avg_mem_gb >= min_avg_mem as f64 &&
                aggregate.peak_mem_gb >= min_peak_mem as f64 &&
                aggregate.avg_gpu >= min_avg_gpu &&
                aggregate.peak_gpu >= min_peak_gpu &&
                aggregate.avg_vmem_pct >= min_avg_vmem &&
                aggregate.peak_vmem_pct >= min_peak_vmem &&
                aggregate.duration >= min_runtime &&
            { if cli.no_gpu { !aggregate.uses_gpu } else { true } } &&
            { if cli.some_gpu { aggregate.uses_gpu } else { true } } &&
            { if cli.completed { (aggregate.classification & LIVE_AT_END) == 0 } else { true } } &&
            { if cli.running { (aggregate.classification & LIVE_AT_END) == 1 } else { true } } &&
            { if cli.zombie { job[0].user.starts_with("_zombie_") } else { true } } &&
            { if let Some(ref cmd) = cli.command { job[0].command.contains(cmd) } else { true } }
        })
        .collect::<Vec<(Aggregate, Vec<logfile::LogEntry>)>>();

    if cli.verbose {
        eprintln!("Number of job records after aggregation filtering: {}", jobvec.len());
    }

    // And sort ascending by lowest beginning timestamp
    jobvec.sort_by(|a, b| a.0.first.cmp(&b.0.first));

    // Select a number of jobs per user, if applicable.  This means working from the bottom up
    // in the vector and marking the n first per user.  We need a hashmap user -> count.
    if let Some(n) = cli.numjobs {
        let mut counts: HashMap<&str,usize> = HashMap::new();
        jobvec.iter_mut().rev().for_each(|(aggregate, job)| {
            if let Some(c) = counts.get(&(*job[0].user)) {
                if *c < n {
                    counts.insert(&job[0].user, *c+1);
                } else {
                    aggregate.selected = false;
                }
            } else {
                counts.insert(&job[0].user, 1);
            }
        })
    }

    if cli.verbose {
        let numselected = jobvec.iter()
            .map(|(aggregate, _)| {
                if aggregate.selected { 1i32 } else { 0i32 }
            })
            .reduce(i32::add)
            .unwrap_or(0);
        eprintln!("Number of job records after output filtering: {}", numselected);
    }

    // Now print.
    //
    // Unix user names are max 8 chars.
    // Linux pids are max 7 decimal digits.
    // We don't care about seconds in the timestamp, nor timezone.

    if cli.raw {
        jobvec.iter().for_each(|(aggregate, job)| {
            println!("{:?}\n{:?}\n", job[0], aggregate);
        });
    } else {
        println!("{:8} {:8}   {:9}   {:16}   {:16}   {:9}  {:9}  {:9}  {:9}   {}",
                 "job#", "user", "time", "start?", "end?", "cpu", "mem gb", "gpu", "gpu mem", "command", );
        let tfmt = "%Y-%m-%d %H:%M";
        jobvec.iter().for_each(|(aggregate, job)| {
            if aggregate.selected {
                let dur = format!("{:2}d{:2}h{:2}m", aggregate.days, aggregate.hours, aggregate.minutes);
                println!("{:7}{} {:8}   {}   {}   {}   {:4}/{:4}  {:4}/{:4}  {:4}/{:4}  {:4}/{:4}   {:22}",
                         job[0].job_id,
                         if aggregate.classification & (LIVE_AT_START|LIVE_AT_END) == LIVE_AT_START|LIVE_AT_END {
                             "!"
                         } else if aggregate.classification & LIVE_AT_START != 0 {
                             "<"
                         } else if aggregate.classification & LIVE_AT_END != 0 {
                             ">"
                         } else {
                             " "
                         },
                         job[0].user,
                         dur,
                         aggregate.first.format(tfmt),
                         aggregate.last.format(tfmt),
                         aggregate.avg_cpu,
                         aggregate.peak_cpu,
                         aggregate.avg_mem_gb,
                         aggregate.peak_mem_gb,
                         aggregate.avg_gpu,
                         aggregate.peak_gpu,
                         aggregate.avg_vmem_pct,
                         aggregate.peak_vmem_pct,
                         job[0].command);
            }
        });
    }
}

fn epoch() -> DateTime<Utc> {
    // FIXME, but this is currently good enough for all our uses
    DateTime::from_utc(NaiveDate::from_ymd_opt(2000,1,1).unwrap().and_hms_opt(0,0,0).unwrap(), Utc)
}

fn now() -> DateTime<Utc> {
    Utc::now()
}

fn one_day_ago() -> DateTime<Utc> {
    now() - chrono::Duration::days(1)
}

fn fail(msg: &str) {
    eprintln!("ERROR: {}", msg);
    process::exit(1);
}
