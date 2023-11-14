# ML-nodes sonar and sonalyze setup

Be sure to read `../README.md` first.

## Files

These are files that are used to drive sonar and the analysis of sonar
logs during production.

The work is driven by cron, so there are two crontabs:

- `jobanalyzer.cron` is a user crontab to run on each host other than
  ML4.  It just runs `sonar`.

- `jobanalyzer-moneypenny.cron` is a user crontab to run on Moneypenny, the analysis host.  It runs
  all the analysis jobs, and it runs on Moneypenny so as not to burden the ML nodes with this task.

The crontabs just run a bunch of shell scripts:

- `sonar.sh` is a script that runs sonar with a set of predetermined
  command line switches and with stdout piped to a predetermined
  location.

- `cpuhog.sh` and `deadweight.sh` are analysis jobs that process the sonar
  logs and look for jobs that either should not be on the ML nodes or
  are stuck and indicate system problems.

- `cpuhog-report.sh` and `deadweight-report.sh` are meta-analysis jobs that process the output from
  `cpuhog.sh` and `deadweight.sh` and produce alerts for problem jobs that have not been seen
  previously.

- `webload-1h.sh` and `webload-24h.sh` produce load graphs every 1h (for hourly breakdowns over the
  last day or week) and 24h (for daily breakdowns over the last month and quarter)

- `upload-data.sh` uploads all produced json reports to the web server

- `webload-5m-and-upload.sh` produces and uploads a moment-to-moment load graph for the last 24h.
  It runs every time sonar runs (roughly) and is its own thing so that it can move less data.

The analyses needs to know what the systems look like, so there are files for that:

- `ml-nodes.json` describes the hardware of the ML nodes, its format
  is documented in `../../sonalyze/MANUAL.md`.

## Production

The typical case in production will be that all of these files are
manually copied into a directory shared among all the ML nodes, called
`$HOME/sonar`, for whatever user is running these jobs.  Also in that
directory must be binaries for `sonar` and `sonalyze`.

If your case is not typical you will need to edit the shell scripts
(to get the paths right and make any other adjustments).

`sonar` runs every 5 minutes and logs data in $HOME/sonar/data/, under
which there is a tree with a directory for each year, under that
directories for each month, and under each month directories for each
day.  Directories are created as necessary.  In each leaf directory
there are csv files named by hosts (eg, `ml8.hpc.uio.no.csv`),
containing the data logged by sonar on that host on that day.

The analysis jobs `cpuhog` and `deadweight` run every two hours now and
log data exactly as `sonar`, except that the per-day log files are
named `cpuhog.csv` and `deadweight.csv`.

(The analysis log files are then further postprocessed off-node by the
`naicreport` system; the latter also sometimes uses the raw logs to
produce reports.)

**TODO: Document the upload script here!**

**TODO: Document the naicreport state files here!**

## UiO setup

The production web server is an nginx instance that runs on a vm on the open internet with static
address `http://158.39.48.160`.  The user is `ubuntu`; this user has password-less sudo capability
on the system.

There must not be any secrets on the system that lets anyone on it into other systems.  It is an
endpoint system.

We use the default nginx setup, so files are in `/var/www/html`.  The files in that directory are
the files from the `dashboard` directory in this repository, plus a subdirectory called `output`.
All files and directories should have user.group=`ubuntu.ubuntu`, but this is probably only
important for the `output` directory.

The access key is in the file `ubuntu-vm.pem` which I'm not going to be including here.  The key
needs to be located in the `~/.ssh` directory of any user account that will run the `upload-data.sh`
script documented above.

## Staging

(This describes a future architecture, once the system is properly set up for production.)

Recall that production has multiple aspects

- running sonar and producing logs
- running sonalyze and producing cpuhog and deadweight state
- running naicreport and producing cpuhog, deadweight, and load reports, and maintaining state
- uploading data to a data directory on the sever
- on the server, serving html, js and json

For staging, we use the same web server as for production, but all the files are in
`/var/www/staging`.

On the ML nodes, there will be a separate user and separate cron jobs producing separate data, and
everything will be uploaded to the staging directory.  Mostly this will "just work", except:

- jobanalyzer-moneypenny.cron will need to not hardcode the work directory
- the upload scripts will need to not hardcode the upload path

If there are several developers working independently then they could use multiple staging
directories.
