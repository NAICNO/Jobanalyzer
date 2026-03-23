# Data sources

Sonalyze always operates on a single data source containing Sonar data.  This source can be:

* a timescaledb database; it is set up and maintained by [slurm-monitor](https://github.com/2maz/slurm-monitor)
* a jobanalyzer database, which is a structured directory tree holding the data directories for one
  or more clusters and some other metadata; it is set up manually and maintained by Sonalyze
* a cluster data directory, which is a read-only directory tree with data for a single cluster, typically
  created by Sonar in its directory-tree-data-sink mode
* a list of read-only Sonar data files, all of the same file type, typically created by Sonar in its
  one-shot mode or by hand (for testing)

All sources are always for local Sonalyze operation, whether for a one-shot operation or for the
daemon.  Remote users of the Sonalyze daemon do not know what data source the remote daemon is
running on.

The range of possible sources reflects both the different use cases that exist for Sonalyze and also
its historical evolution.

## Timescaledb database

To take data from a Timescaledb database (which is just a PostgreSQL database with some special
sauce), use the `-database-uri` option with a `postgresql:` URI that names the database in the
normal way:

```
$ sonalyze daemon -database-uri postgresql:...
```

Typically, all data come from the database in this case.  However, it is possible to also provide a
`-jobanalyzer-dir` argument, in which case some cluster metadata are taken from files in that
directory.  See below.

A Timescaledb database is naturally a multi-user database, there can be independent readers and
writers, and the database can store not just Sonar timeseries data but also computed data, and data
in the database can be rewritten.


## Jobanalyzer directory

A jobanalyzer directory D has two notable subdirectories, `cluster-config` and `data`, that are in
correspondence with each other.  The former contains cluster metadata, which are JSON files with
names on the format `<cluster-name>-config.json`, and the latter contains cluster data
subdirectories, one subdirectory `<cluster-name>` for each cluster; see below.

The metadata used to be richer but are now very limited.  Notably they include user-friendly cluster
names and cluster aliases.  Metadata are always optional.  Sonalyze will infer some metadata from
the cluster directory name or command line arguments if metadata are not provided by a Jobanalyzer
directory.

To run sonalyze on a jobanalyzer directory:
```
$ sonalyze daemon -jobanalyzer-dir ...
```

A Jobanalyzer database is writable but append-only and stores only Sonar timeseries data, and there
can be only one Sonalyze user of such a database at a time.  The Jobanalyzer daemon facilitates
concurrent multi-user access.  Computed data are cached in the daemon's memory but are not persisted
to the Jobanalyzer database.


## Cluster data directory

A cluster data directory `D` is the root of a date-encoded tree: `D/yyyy/mm/dd/<filename>.json`
where the dates correspond to the time stamps of the data in the files and the filename encodes the
Sonar data type, as described in the next section.  For a directory name D, the cluster name is set
to `D.data` (with `D` folded to lower case and with non-hostname characters translated to `_`).

```
$ sonalyze daemon -data-dir ...
```

## Sonar data files

Sonalyze can be run on a set of individual Sonar data files, which must all be of the same type (see
below):

```
$ sonalyze daemon file-name ...
```

When sonalyze is run on a set of log files, the cluster name will be set to `anonymous-cluster.logfiles`.


## Cluster names in data and metadata

Sonar data always belong to a *cluster* (think HPC system), which has a canonical *cluster name*.
The cluster name usually takes the form of a host name (`fox.educloud.no`, `olivia.sigma2.no`) and
is in fact restricted to valid host name characters.  Sonalyze will generally replace invalid
characters in cluster names with underscores.

The cluster name may be present both in Sonar data (and always is present when Sonar is running in
daemon mode) and in metadata transmitted to the back-end with the Sonar data (the cluster name is
part of the message topic when data are transmitted via Kafka).

The back-ends are not required to check that the cluster name in the data and the cluster name in
the metadata are the same, and if they are not, the different back-ends may diverge in their
behavior.  Sonalyze ignores the embedded cluster name, always keying off the metadata, while
Slurm-monitor depends entirely on the embedded data.


## Data file types and formats

Sonalyze data files are just streams of Sonar-produced JSON data.  The data formats are described in
`NEW-FORMAT.md` in the Sonar documentation.

Sonalyze files in Sonalyze directory trees are named as they are by the Sonar daemon's "directory
data sink", see the section "The `[directory]` section" in `HOWTO-DAEMON.md` in the Sonar
documentation, and Sonalyze directory trees are also organized in the way described there.

Sonalyze can handle older-format Sonar JSON and CSV files, which are organized differently but have
a looser naming scheme; UTSL.
