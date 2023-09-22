/// `sonalize` -- Analyze `sonar` log files
///
/// See MANUAL.md for a manual, or run with --help for brief help.
///
/// Quirks
///
/// Having the absence of --user mean "only $LOGNAME" can be confusing -- though it's the right
/// thing for a use case where somebody is looking only at her own jobs.
///
/// The --from and --to values are used *both* for filtering files in the directory tree of logs
/// (where it is used to generate directory names to search) *and* for filtering individual records
/// in the log files.  Things can become a confusing if the log records do not have dates
/// corresponding to the directories they are located in.  This is mostly a concern for testing;
/// production data will have a sane mapping.
///
/// Some filtering options select *records* (from, to, host, user, exclude) and some select *jobs*
/// (the rest of them), and this can be confusing.  For user and exclude this does not matter
/// (modulo setuid or similar personality changes).  The user might expect that from/to/host would
/// select jobs instead of records, s.t. if a job ran in the time interval (had samples in the
/// interval) then the entire job should be displayed, including data about it outside the interval.
/// Ditto, that if a job ran on a selected host then its work on all hosts should be displayed.  But
/// it just ain't so.

mod format;
mod jobs;
mod load;
mod prjobs;

use anyhow::{bail, Result};
use chrono::{Datelike, NaiveDate};
use clap::{Args, Parser, Subcommand};
use sonarlog::{self, HostFilter, LogEntry, Timestamp};
use std::collections::HashSet;
use std::env;
use std::io;
use std::num::ParseIntError;
use std::ops::Add;
use std::process;
use std::str::FromStr;
use std::time;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
pub struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// Print information about jobs
    Jobs(JobArgs),

    /// Print information about system load
    Load(LoadArgs),
}

#[derive(Args, Debug)]
pub struct JobArgs {
    #[command(flatten)]
    input_args: InputArgs,

    #[command(flatten)]
    filter_args: JobFilterAndAggregationArgs,

    #[command(flatten)]
    print_args: JobPrintArgs,

    #[command(flatten)]
    meta_args: MetaArgs,
}

#[derive(Args, Debug)]
pub struct LoadArgs {
    #[command(flatten)]
    input_args: InputArgs,

    #[command(flatten)]
    filter_args: LoadFilterAndAggregationArgs,

    #[command(flatten)]
    print_args: LoadPrintArgs,

    #[command(flatten)]
    meta_args: MetaArgs,
}

#[derive(Args, Debug)]
pub struct InputArgs {
    /// Select the root directory for log files [default: $SONAR_ROOT]
    #[arg(long)]
    data_path: Option<String>,

    /// Include this user, "-" for all (repeatable) [default: command dependent]
    #[arg(long, short)]
    user: Vec<String>,

    /// Exclude records where the user name matches this string (repeatable) [default: none]
    #[arg(long)]
    exclude_user: Vec<String>,

    /// Select records with this command name matches this string (repeatable) [default: all]
    #[arg(long)]
    command: Vec<String>,

    /// Exclude records where the command name matches this string (repeatable) [default: none]
    #[arg(long)]
    exclude_command: Vec<String>,

    /// Select records for this job (repeatable) [default: all]
    #[arg(long, short)]
    job: Vec<String>,

    /// Select records by this time and later.  Format can be YYYY-MM-DD, or Nd or Nw
    /// signifying N days or weeks ago [default: 1d, ie 1 day ago]
    #[arg(long, short, value_parser = parse_time_start_of_day)]
    from: Option<Timestamp>,

    /// Select records by this time and earlier.  Format can be YYYY-MM-DD, or Nd or Nw
    /// signifying N days or weeks ago [default: now]
    #[arg(long, short, value_parser = parse_time_end_of_day)]
    to: Option<Timestamp>,

    /// Select records for this host name (repeatable) [default: all]
    #[arg(long)]
    host: Vec<String>,

    /// File containing JSON data with system information, for when we want to print or use system-relative values [default: none]
    #[arg(long)]
    config_file: Option<String>,

    /// Log file names (overrides --data-path)
    #[arg(last = true)]
    logfiles: Vec<String>,
}

#[derive(Args, Debug)]
pub struct LoadFilterAndAggregationArgs {
    /// Bucket and average records hourly, cf --daily and --none [default]
    #[arg(long)]
    hourly: bool,

    /// Bucket and average records daily
    #[arg(long)]
    daily: bool,

    /// Do not bucket and average records
    #[arg(long)]
    none: bool,
}

#[derive(Args, Debug, Default)]
pub struct JobFilterAndAggregationArgs {
    /// Select only jobs with at least this many samples [default: 2]
    #[arg(long)]
    min_samples: Option<usize>,

    /// Select only jobs with at least this much average CPU use (100=1 full CPU)
    #[arg(long, default_value_t = 0)]
    min_cpu_avg: usize,

    /// Select only jobs with at least this much peak CPU use (100=1 full CPU)
    #[arg(long, default_value_t = 0)]
    min_cpu_peak: usize,

    /// Select only jobs with at most this much average CPU use (100=1 full CPU)
    #[arg(long, default_value_t = 100000000)]
    max_cpu_avg: usize,

    /// Select only jobs with at most this much peak CPU use (100=1 full CPU)
    #[arg(long, default_value_t = 100000000)]
    max_cpu_peak: usize,

    /// Select only jobs with at least this much relative average CPU use (100=all cpus)
    #[arg(long, default_value_t = 0)]
    min_rcpu_avg: usize,

    /// Select only jobs with at least this much relative peak CPU use (100=all cpus)
    #[arg(long, default_value_t = 0)]
    min_rcpu_peak: usize,

    /// Select only jobs with at most this much relative average CPU use (100=all cpus)
    #[arg(long, default_value_t = 100)]
    max_rcpu_avg: usize,

    /// Select only jobs with at most this much relative peak CPU use (100=all cpus)
    #[arg(long, default_value_t = 100)]
    max_rcpu_peak: usize,

    /// Select only jobs with at least this much average main memory use (GB)
    #[arg(long, default_value_t = 0)]
    min_mem_avg: usize,

    /// Select only jobs with at least this much peak main memory use (GB)
    #[arg(long, default_value_t = 0)]
    min_mem_peak: usize,

    /// Select only jobs with at least this much relative average main memory use (100=all memory)
    #[arg(long, default_value_t = 0)]
    min_rmem_avg: usize,

    /// Select only jobs with at least this much relative peak main memory use (100=all memory)
    #[arg(long, default_value_t = 0)]
    min_rmem_peak: usize,

    /// Select only jobs with at least this much average GPU use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_gpu_avg: usize,

    /// Select only jobs with at least this much peak GPU use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_gpu_peak: usize,

    /// Select only jobs with at most this much average GPU use (100=1 full GPU card)
    #[arg(long, default_value_t = 100000000)]
    max_gpu_avg: usize,

    /// Select only jobs with at most this much peak GPU use (100=1 full GPU card)
    #[arg(long, default_value_t = 100000000)]
    max_gpu_peak: usize,

    /// Select only jobs with at least this much relative average GPU use (100=all cards)
    #[arg(long, default_value_t = 0)]
    min_rgpu_avg: usize,

    /// Select only jobs with at least this much relative peak GPU use (100=all cards)
    #[arg(long, default_value_t = 0)]
    min_rgpu_peak: usize,

    /// Select only jobs with at most this much relative average GPU use (100=all cards)
    #[arg(long, default_value_t = 100)]
    max_rgpu_avg: usize,

    /// Select only jobs with at most this much relative peak GPU use (100=all cards)
    #[arg(long, default_value_t = 100)]
    max_rgpu_peak: usize,

    /// Select only jobs with at least this much average GPU memory use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_gpumem_avg: usize,

    /// Select only jobs with at least this much peak GPU memory use (100=1 full GPU card)
    #[arg(long, default_value_t = 0)]
    min_gpumem_peak: usize,

    /// Select only jobs with at least this much relative average GPU memory use (100=all cards)
    #[arg(long, default_value_t = 0)]
    min_rgpumem_avg: usize,

    /// Select only jobs with at least this much relative peak GPU memory use (100=all cards)
    #[arg(long, default_value_t = 0)]
    min_rgpumem_peak: usize,

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

    /// Aggregate data across hosts (appropriate for batch systems)
    #[arg(long, short, default_value_t = false)]
    batch: bool,
}

#[derive(Args, Debug)]
pub struct LoadPrintArgs {
    /// Print records for all times (after bucketing), cf --last [default]
    #[arg(long)]
    all: bool,

    /// Print records for the last time instant (after bucketing)
    #[arg(long)]
    last: bool,

    /// Select fields for the output [default: see MANUAL.md]
    #[arg(long)]
    fmt: Option<String>,

    /// After bucketing, do not print anything for time slots that are empty
    #[arg(long, default_value_t = false)]
    compact: bool,
}

#[derive(Args, Debug, Default)]
pub struct JobPrintArgs {
    /// Break down job by host, by command, or both [default: neither]
    #[arg(long)]
    breakdown: Option<String>,

    /// Print at most these many most recent jobs per user [default: all]
    #[arg(long, short)]
    numjobs: Option<usize>,

    /// Select fields for the output [default: see MANUAL.md]
    #[arg(long)]
    fmt: Option<String>,
}

#[derive(Args, Debug, Default)]
pub struct MetaArgs {
    /// Print useful statistics about the input to stderr, then terminate
    #[arg(long, short, default_value_t = false)]
    verbose: bool,

    /// Turn off default filtering, and print unformatted data (for developers)
    #[arg(long, default_value_t = false)]
    raw: bool,
}

// The command arg parsers don't need to include the string being parsed because the error generated
// by clap includes that.

// YYYY-MM-DD, but with a little (too much?) flexibility.  Or Nd, Nw.
fn parse_time(s: &str, end_of_day: bool) -> Result<Timestamp> {
    if let Some(n) = s.strip_suffix('d') {
        if let Ok(k) = usize::from_str(n) {
            Ok(sonarlog::now() - chrono::Duration::days(k as i64))
        } else {
            bail!("Invalid date")
        }
    } else if let Some(n) = s.strip_suffix('w') {
        if let Ok(k) = usize::from_str(n) {
            Ok(sonarlog::now() - chrono::Duration::weeks(k as i64))
        } else {
            bail!("Invalid date")
        }
    } else {
        let parts = s
            .split('-')
            .map(|x| usize::from_str(x))
            .collect::<Vec<Result<usize, ParseIntError>>>();
        if !parts.iter().all(|x| x.is_ok()) || parts.len() != 3 {
            bail!("Invalid date syntax");
        }
        let vals = parts
            .iter()
            .map(|x| *x.as_ref().unwrap())
            .collect::<Vec<usize>>();
        let d = NaiveDate::from_ymd_opt(vals[0] as i32, vals[1] as u32, vals[2] as u32);
        if !d.is_some() {
            bail!("Invalid date");
        }
        // TODO: This is roughly right, but what we want here is the day + 1 day and then use `<`
        // rather than `<=` in the filter.
        let (h, m, s) = if end_of_day { (23, 59, 59) } else { (0, 0, 0) };
        Ok(sonarlog::timestamp_from_ymdhms(
            d.unwrap().year(),
            d.unwrap().month(),
            d.unwrap().day(),
            h,
            m,
            s,
        ))
    }
}

fn parse_time_start_of_day(s: &str) -> Result<Timestamp> {
    parse_time(s, false)
}

fn parse_time_end_of_day(s: &str) -> Result<Timestamp> {
    parse_time(s, true)
}

// This is WwDdHhMm with all parts optional but at least one part required.  There is possibly too
// much flexibility here, as the parts can be in any order.
fn run_time(s: &str) -> Result<chrono::Duration> {
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
            if ds == ""
                || (ch != 'd' && ch != 'h' && ch != 'm' && ch != 'w')
                || (ch == 'd' && have_days)
                || (ch == 'h' && have_hours)
                || (ch == 'm' && have_minutes)
                || (ch == 'w' && have_weeks)
            {
                bail!("Bad suffix")
            }
            let v = u64::from_str(&ds);
            if !v.is_ok() {
                bail!("Bad number")
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
        bail!("Inconsistent")
    }

    days += weeks * 7;
    hours += days * 24;
    minutes += hours * 60;
    let seconds = minutes * 60;
    Ok(chrono::Duration::from_std(time::Duration::from_secs(
        seconds,
    ))?)
}

#[test]
fn test_run_time() {
    // This is illegal as of now, we might want to change this?
    assert!(run_time("3").is_err());

    // Years (and other things) are not supported
    assert!(run_time("3y").is_err());
    assert!(run_time("d").is_err());

    let x = run_time("3m").unwrap();
    assert!(x.num_minutes() == 3);
    assert!(x.num_minutes() == x.num_seconds() / 60);
    assert!(x.num_hours() == 0);

    let x = run_time("4h7m").unwrap();
    assert!(x.num_minutes() == 4*60+7);
    assert!(x.num_minutes() == x.num_seconds() / 60);
    assert!(x.num_hours() == 4);
    assert!(x.num_hours() == x.num_minutes() / 60);

    let x = run_time("4h").unwrap();
    assert!(x.num_minutes() == 4*60);
    assert!(x.num_seconds() == 4*60*60);

    let x = run_time("2d4h7m").unwrap();
    assert!(x.num_minutes() == (2*24 + 4)*60 + 7);

    let x = run_time("2d").unwrap();
    assert!(x.num_minutes() == (2*24)*60);
    assert!(x.num_seconds() == (2*24)*60*60);
}

fn main() {
    match sonalyze() {
        Ok(()) => {}
        Err(msg) => {
            eprintln!("ERROR: {}", msg);
            process::exit(1);
        }
    }
}

fn sonalyze() -> Result<()> {
    let cli = Cli::parse();

    let input_args = match cli.command {
        Commands::Jobs(ref jobs_args) => &jobs_args.input_args,
        Commands::Load(ref load_args) => &load_args.input_args,
    };

    let meta_args = match cli.command {
        Commands::Jobs(ref jobs_args) => &jobs_args.meta_args,
        Commands::Load(ref load_args) => &load_args.meta_args,
    };

    // Validate and regularize input parameters from switches and defaults.

    let (
        from,
        to,
        include_hosts,
        include_jobs,
        include_users,
        exclude_users,
        include_commands,
        exclude_commands,
        system_config,
        logfiles,
    ) = {
        // Included date range.  These are used both for file names and for records.

        let from = if let Some(x) = input_args.from {
            x
        } else {
            sonarlog::now() - chrono::Duration::days(1)
        };
        let to = if let Some(x) = input_args.to {
            x
        } else {
            sonarlog::now()
        };
        if from > to {
            bail!("The --from time is greater than the --to time");
        }

        // Included host set, empty means "all"

        let include_hosts = {
            let mut hosts = HostFilter::new();
            for host in &input_args.host {
                hosts.insert(host)?;
            }
            hosts
        };

        // Included job numbers, empty means "all"

        let include_jobs = {
            let mut jobs = HashSet::<usize>::new();
            for job in &input_args.job {
                jobs.insert(usize::from_str(job)?);
            }
            jobs
        };

        // Included users.  The default depends on various other switches.

        let all_users =
            if let Commands::Load(_) = cli.command {
                // `load` implies `--user=-` b/c we're interested in system effects.
                true
            } else if !input_args.job.is_empty() {
                // `jobs --job=...` implies `--user=-` b/c the job also implies a user.
                true
            } else if !input_args.exclude_user.is_empty() {
                // `jobs --exclude-user=...` implies `--user=-` b/c the only sane way to include
                // many users so that some can be excluded is by also specifying `--users=-`.
                true
            } else if let Commands::Jobs(ref jobs_args) = cli.command {
                // `jobs --zombie` implies `--user=-` because the use case for `--zombie` is to hunt
                // across all users.
                jobs_args.filter_args.zombie
            } else {
                false
            };

        let include_users = {
            let mut users = HashSet::<String>::new();
            if input_args.user.len() > 0 {
                // Not the default value
                if input_args.user.iter().any(|user| user == "-") {
                    // Everyone, so do nothing
                } else {
                    for user in &input_args.user {
                        users.insert(user.to_string());
                    }
                }
            } else if all_users {
                // Everyone, so do nothing
            } else {
                if let Ok(u) = env::var("LOGNAME") {
                    users.insert(u);
                };
            }
            users
        };

        // Excluded users.

        let mut exclude_users = {
            let mut excluded = HashSet::<String>::new();
            if input_args.exclude_user.len() > 0 {
                // Not the default value
                for user in &input_args.exclude_user {
                    excluded.insert(user.to_string());
                }
            } else {
                // Nobody
            }
            excluded
        };
        exclude_users.insert("root".to_string());
        exclude_users.insert("zabbix".to_string());

        // Included commands.

        let include_commands = {
            let mut included = HashSet::<String>::new();
            if input_args.command.len() > 0 {
                for command in &input_args.command {
                    included.insert(command.to_string());
                }
            } else {
                // Every command
            }
            included
        };

        // Excluded commands.
        
        let mut exclude_commands = {
            let mut excluded = HashSet::<String>::new();
            if input_args.exclude_command.len() > 0 {
                // Not the default value
                for command in &input_args.exclude_command {
                    excluded.insert(command.to_string());
                }
            } else {
                // Nobody
            }
            excluded
        };
        exclude_commands.insert("bash".to_string());
        exclude_commands.insert("zsh".to_string());
        exclude_commands.insert("sshd".to_string());
        exclude_commands.insert("tmux".to_string());
        exclude_commands.insert("systemd".to_string());

        // Data path, if present.

        let data_path = if input_args.data_path.is_some() {
            input_args.data_path.clone()
        } else if let Ok(val) = env::var("SONAR_ROOT") {
            Some(val)
        } else if let Ok(val) = env::var("HOME") {
            Some(val + "/sonar_logs")
        } else {
            None
        };

        // System configuration, if specified.

        let system_config = if let Some(ref config_filename) = input_args.config_file {
            Some(sonarlog::read_from_json(&config_filename)?)
        } else {
            None
        };

        // Log files, filtered by host and time range.
        //
        // If the log files are provided on the command line then there will be no filtering by host
        // name on the file name.  This is by design.

        let logfiles = if input_args.logfiles.len() > 0 {
            input_args.logfiles.clone()
        } else {
            if meta_args.verbose {
                eprintln!("Data path: {:?}", data_path);
            }
            if data_path.is_none() {
                bail!("No data path");
            }
            let maybe_logfiles =
                sonarlog::find_logfiles(&data_path.unwrap(), &include_hosts, from, to);
            if let Err(ref msg) = maybe_logfiles {
                bail!("{msg}");
            }
            maybe_logfiles.unwrap()
        };

        if meta_args.verbose {
            eprintln!("Log files: {:?}", logfiles);
        }

        (
            from,
            to,
            include_hosts,
            include_jobs,
            include_users,
            exclude_users,
            include_commands,
            exclude_commands,
            system_config,
            logfiles,
        )
    };

    // Input filtering logic is the same for both job and load listing, the only material
    // difference (handled above) is that the default user set for load listing is "all".

    let filter = |e:&LogEntry| {
        ((&include_users).is_empty() || (&include_users).contains(&e.user))
            && ((&include_hosts).is_empty() || (&include_hosts).contains(&e.hostname))
            && ((&include_jobs).is_empty() || (&include_jobs).contains(&(e.job_id as usize)))
            && !(&exclude_users).contains(&e.user)
            && ((&include_commands).is_empty() || (&include_commands).contains(&e.command))
            && !(&exclude_commands).contains(&e.command)
            && from <= e.timestamp
            && e.timestamp <= to
    };

    let (entries, earliest, latest, records_read) = sonarlog::read_logfiles(&logfiles)?;
    let streams = sonarlog::postprocess_log(entries, &filter, &system_config);

    match cli.command {
        Commands::Load(ref load_args) => {
            load::aggregate_and_print_load(
                &mut io::stdout(),
                &system_config,
                &include_hosts,
                from,
                to,
                &load_args.filter_args,
                &load_args.print_args,
                meta_args,
                streams,
            )
        }
        Commands::Jobs(ref job_args) => {
            if meta_args.verbose {
                eprintln!("Number of samples read: {}", records_read);
                let numrec = streams
                    .iter()
                    .map(|(_, recs)| recs.len())
                    .reduce(usize::add)
                    .unwrap_or_default();
                eprintln!("Number of samples after input filtering: {}", numrec);
                eprintln!("Number of streams after input filtering: {}", streams.len());
            }
            jobs::aggregate_and_print_jobs(
                &mut io::stdout(),
                &system_config,
                &job_args.filter_args,
                &job_args.print_args,
                meta_args,
                streams,
                earliest,
                latest,
            )
        }
    }
}
