# Sonar and Jobanalyzer production setups

Every cluster has a *cluster name* that distinguishes it globally.  Frequently the cluster name is
the FQDN of the cluster's login node.  These clusters are defined at present:

* `mlx.hpc.uio.no` - abbreviated `ml` - UiO machine learning nodes
* `fox.educloud.no` - abbreviated `fox` - UiO Fox supercomputer

All clusters have individual setups (for now), owing partly to how the setups have evolved
independently and partly to real differences between the clusters.  There are however many
commonalities.

(The cluster name scheme is imperfectly implemented throughout the system but we are moving in the
direction of using it for everything.)

On every *compute node* in a cluster we run the analysis program `sonar` to obtain *sample data*.
The sample data are exfiltrated to the *analysis node* where the data are aggregated and analyzed.
The analyses generate *reports* that are uploaded to the *web node*.  The web node serves the
*dashboard* which delivers the reports to web clients that request them.  The web node can also
serve interactive queries, which it runs by contacting the analysis node.

Thus to set up Jobanalyzer, one must set up sonar on each of the compute nodes, the data management
and analysis framework on the analysis node, and the web server and dashboard framework on the web
node.

What follows is (unfortunately) partial and in flux.  It is cleaned up as the architecture
stabilizes.

## Sonar

We run `sonar` by means of cron on each node in the cluster.  This samples the node every few
minutes and exfiltrates the sample data to the analysis host (known as "the sonar VM" henceforth) by
means of the `exfiltrate` program.  (On some systems, the exfiltration is currently by means of
writing the data to a shared disk, but this will disappear.)

The cron script `sonar-runner.cron` in each cluster directory contains a cron setup for the cluster.
The shell script `sonar.sh` actually runs `sonar` with the appropriate options, which will vary from
cluster to cluster.

### ML nodes (cluster name: mlx.hpc.uio.no)

For the time being, `sonar` is run as user `larstha` and `~larstha/sonar` contains the cron script,
`sonar.sh` and the `sonar` and `exfiltrate` binaries.

### Fox (cluster name: fox.educloud.no)

`sonar` is run as the system user `sonar`.  `/cluster/var/sonar/bin` holds the cron script,
`sonar.sh` and the `sonar` and `exfiltrate` binaries.

## The Sonar VM and data storage

On the Sonar VM, data are ingested by means of the `infiltrate` program and stored in a local file
system under the UID that runs `infiltrate` (currently "ubuntu").

The root for Jobanalyzer is in `~ubuntu/sonar`.  In this directory, the subdirectory `data/` holds
the ingested data.  There is one directory for each cluster, it is named with the cluster name (eg,
`data/mlx.hpc.uio.no/` and `data/fox.educloud.no/`).  In each cluster directory there is a
subdirectory per year; in each year, one per month; in each month, one per day; and in each day, one
file per host name (the host name may be local to the cluster and not globally unique, this is
common for HPC systems).  The file name is `<hostname>.csv`.

## Sonalyze, naicreport, and analysis data

Separate cron jobs run on the Sonar VM to perform analyses of the ingested data and upload the files
to a web server.  These run as the same UID as `infiltrate`, ensuring that files can be shared among
ingestor and analyzers without being world-readable.

The analysis jobs have some intermediate products and produce some reports.  For this, they need two
data directories `$state_dir` and `$report_dir`.  Typically these are in subdirectories of a common
directory that also holds executables (`sonalyze` and `naicreport`), shell scripts (many), and
cluster configuration data (`ml-nodes.json`, `fox.json`, etc).  If this directory does not contain
the raw log data directory then it may have a symlink to that directory, so that the scripts don't
need to know where that directory is.

It is necessary for the analysis jobs for separate clusters to use separate state directories, as
files in the state are not tagged with cluster or host name.  The report directories should also be
separated by cluster.

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
