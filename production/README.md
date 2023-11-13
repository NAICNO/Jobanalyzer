# Sonar and Jobanalyzer production setups

All clusters have individual setups (for now), owing partly to how the setups have evolved
independently and partly to real differences between the clusters.  There are however many
commonalities, as follows.

## Sonar and sonar data

We run `sonar` by means of cron on each node in the cluster.  This samples the node every few
minutes and (currently) stores the sample data in files on a shared disk under the directory
`$data_dir`.  In this directory there is a subdirectory per year; in each year, one per month; in
each month, one per day; and in each day, one file per host name.  The file name is
`<hostname>.csv`.

It is possible for multiple clusters with disjoint host name sets to use the same shared file
system, as the file naming scheme ensures there will be no name collisions.

The cron script `sonar-runner.cron` in each cluster directory contains a cron setup for the cluster.
The shell script `sonar.sh` actually runs `sonar` with the appropriate options, which will vary from
cluster to cluster.

### ML nodes

For the time being, `sonar` is run as user `larstha` and `~larstha/sonar` contains the cron script,
`sonar.sh` and the `sonar` binary.  The value of `$data_dir` is `~larstha/sonar/data/`.

### Fox

`sonar` is run as the system user `sonar`.  `/cluster/var/sonar/bin` holds the cron script,
`sonar.sh` and the `sonar` binary.  The value of `$data_dir` is `/cluster/var/sonar/data`.

## Sonalyze, naicreport, and analysis data

A separate cron job running on a separate host, not necessarily in the cluster but (currently) with
access to the shared file system, runs a number of analysis jobs.

The analysis jobs have some intermediate products and produce some reports.  For this, they need two
data directories `$state_dir` and `$report_dir`.  Typically these are subdirectories of a common
directory that also holds executables (`sonalyze`, `naicreport`, and `loginfo`), shell scripts
(many), and cluster configuration data (`ml-nodes.json`, `fox.json`, etc).  If this directory does
not contain the raw log data directory then it may have a symlink to that directory, so that the
scripts don't need to know where that directory is.

It is necessary for the analysis jobs for separate clusters to use separate state directories, as
files in the state are not tagged with cluster or host name.

However, the reports are all tagged with cluster or host name, and the report directory can be
shared if desired.

In this directory there is also a key file (`.pem`) or a link to one; this is used for data
uploading.  The file `upload-config.sh` contains information about how to upload data to the web
server.

### ML nodes

Currently, the analysis job is run as user `larstha` on `moneypenny`.
`/itf-fi-ml/home/larstha/sonar` is the same directory as used by the sonar job.  It contains all the
analysis shell scripts, the analysis executables, and the analysis subdirectories `state` and
`output`.

### Fox

Currently, the analysis job is run as user `ec-larstha` on `login-1`.  `~ec-larstha/sonar` has the
analysis shell scripts and executables, along with subdirectories `state` and `output`, and a symlink
to the data directory.
