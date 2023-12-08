// Execute a command with a timeout and safe handling of the communcation.
//
// Heavily inspired by code in Sonar that I wrote, but which is not copied here due to Sonar having
// an incompatible license.

use std::io;
use std::time::Duration;
use subprocess::{Exec, ExitStatus, Redirection};

pub fn run_with_timeout(command: &str, timeout_seconds: u64) -> Result<String, String> {
    let mut p = match Exec::shell(command)
        .stdout(Redirection::Pipe)
        .stderr(Redirection::Pipe)
        .popen()
    {
        Ok(p) => p,
        Err(_) => {
            return Err(command.to_string());
        }
    };

    // This is an elaborate workaround for a Rust bug.  There is a limited capacity in the pipe.
    // When the pipe fills up the child stops, which means that we'll time out if we use a timeout
    // or will hang indefinitely if not; this is a problem for subprocesses that produce a lot of
    // output, as they sometimes will.
    //
    // The solution for this problem is to be sure to drain the pipe while we are also waiting for
    // the termination of the child.  See
    //
    //   https://github.com/rust-lang/rust/issues/45572,
    //   https://github.com/rust-lang/rust/issues/45572#issuecomment-860134955
    //
    // See also https://doc.rust-lang.org/std/process/index.html ("Handling I/O").
    //
    // Handle it by limiting the amount of time we're willing to wait for output to become
    // available.

    let mut comm = p
        .communicate_start(None)
        .limit_time(Duration::new(timeout_seconds, 0));
    let mut stdout_result = "".to_string();
    let failed = loop {
        match comm.read_string() {
            Ok((Some(stdout), Some(stderr))) => {
                if !stderr.is_empty() {
                    stdout_result += &stderr;
                    // Command produced error output
                    break true;
                } else if stdout.is_empty() {
                    // Command terminated normally, probably
                    // This is always EOF because timeouts are signaled as Err()
                    break false;
                } else {
                    stdout_result += &stdout;
                }
            }
            Ok((_, _)) => {
                // Some type of internal error
                stdout_result = "Internal error".to_string();
                break true
            }
            Err(e) => {
                if e.error.kind() == io::ErrorKind::TimedOut {
                    match p.terminate() {
                        Ok(_) => {
                            // Command is hung
                            stdout_result = "Timed out".to_string();
                            break true
                        }
                        Err(_) => {
                            // Some type of internal error
                            stdout_result = "Timed out / internal error".to_string();
                            break true
                        }
                    }
                }
                // Some type of internal error
                break true
            }
        }
    };

    match p.wait() {
        Ok(ExitStatus::Exited(0)) => {
            if failed {
                Err(stdout_result)
            } else {
                Ok(stdout_result)
            }
        }
        _ => {
            // Various error conditions, we don't care.
            Err(stdout_result)
        }
    }
}
