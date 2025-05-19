# The database

## Data model at a glance

The database exposes six tables:

* `cluster` is the table of clusters known to the DB, with their aliases and descriptions
* `cluzter` is the table of Slurm partitions on the cluster, with their nodes
* `config` is the table of per-node manually-maintained configuration information
* `sysinfo` is the table of per-node system information extracted by Sonar on each node every day
* `sacct` is the table of completed Slurm jobs on the cluster
* `sample` is the table of Sonar samples, collected by Sonar on each node

Except for `config` the tables are append-only.  The `cluzter`, `sysinfo`, `sacct` and `sample`
tables are collectively known as the "data tables" and these are organized by time.

Every query against the data tables must specify a time period within which to run the query.

Every query against the table `config` and the data tables must specify a cluster.

It is a bug that `config` is not append-only and organized by time; this will have to change, as it
is sometimes used to describe the time-relative data in the data tables.

It is a bug that cluster configuration information is split between `cluster`, `cluzter`, `config`,
`sysinfo`, and additional "background" files with augmenting information that are not stored in the
database.

It is (sort of) a bug that we have both `cluster` and `cluzter` (as well as a database-internal
concept called `Cluster` that is not closely related to the two), it's the result of a series of
unfortunate events.

## Implementation

### Representation

In the implementation (here the `v` prefix is currently always `0`):

* `cluster` is a table constructed from the names of the subdirectories of the Sonalyze daemon's
  data directory, the top level `cluster-aliases.json` file, and the cluster information in the
  hand-maintained per-cluster configuration files, `<cluster-name>-config.json`, stored within the
  daemon's directories
* `cluzter` is constructed from the individual `v+cluzter-slurm.json` files in the cluster's data
  directories
* `config` exposes the per-node information in those per-cluster configuration files
* `sysinfo` is constructed from the individual `sysinfo-<nodename>.json` and
  `v+sysinfo-<nodename>.json` files in the cluster's data directories
* `sacct` constructed from the individual `slurm-sacct.csv` and `v+job-slurm.json` files in the
  cluster's data directories
* `sample` is constructed from the individual `<nodename>.csv` and `v+sample-<nodename>.json` files
  in the cluster's data directories

Jobanalyzer's data directory has a subdirectory for each cluster (the subdirectory's name is the
canonical cluster name), and within each cluster the data are organized in directory trees by year,
month, and day.  UTC is used throughout, so the 00:30am data from some node in Norway (in time zone
UTC+1 or UTC+2) on some date normally ends up in the directory for the previous date.

It is a bug that `cluster-aliases.json` exists at all, as the per-cluster configuration files also
carry alias information.  However, currently the alias information in the per-cluster configuration
files is ignored by everyone, only the global aliases file is consulted.

Note that there are no index files or files of computed data - everything currently operates on raw
data, recomputing what is needed when it is needed, and relying on in-memory caching to make this
fast.  There are good (resilience, simplicity) and bad (performance, complexity) consequences of
this that I won't go into here.

### Caching

Data are broadly cached.  For `cluster` and `config` these caches are never purged except by a
restart.  For the data tables the caches are capacity-based and are cleaned by a 2-random LRU
method, but are never fully purged.  Also, for the data tables, a file's cache is cleared when new
data arrives for the file - the file is removed from the cache and then the data are appended to the
file on disk.  Subsequent read access to the file will cache it again.

It is a bug that a restart is required to purge the caches.  Caches should all be purged if the
daemon receives SIGHUP.

### Code

In this directory:

- the `cluster.go` file implements the `cluster` table
- the `cluzter*.go` files implement the `cluzter` table
- the `config.go` file implements the `config` table
- the `sample*.go` files implement the `sample` table
- the `sysinfo*.go` files implement the `sysinfo` table
- the `sacct*.go` files implement the `sacct` table

Below that:

- `clusterstore.go` is an interface to the loading and caching subsystem for `sample`, `sysinfo`,
  and `sacct` data
- `persistentcluster.go` and `transientcluster.go` are implementations of that interface for a
  disk-based directory tree and a list of files, respectively
- `logfile.go` handles individual files in all cases
- `cache.go` is the cache logic used by all this

And at the bottom:

- `csvtokens.go` and `csvutil.go` support fast, non-allocating CSV parsers
