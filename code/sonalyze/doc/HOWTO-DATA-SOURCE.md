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

Sonar data can be self-identifying: they can contain their cluster name.  This datum, if present,
should be considered advisory at best; it is uniformly ignored by Sonalyze, which always takes the
cluster name from metadata.  TODO: it is not yet clear to me if the slurm-monitor ingestor uses the
embedded cluster name or if it keys off the Kafka topic.


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
Sonar data type, as described in the next section.  The name D is always taken to be the cluster
name.

```
$ sonalyze daemon -data-dir ...
```

TODO: As of now, the directory-name to cluster-name inference is only implemented when the directory
is inside a jobanalyzer directory, see above, but this will change.

TODO: We will allow for `-cluster-name` and `-cluster-aliases` to name the cluster, overriding `D`.


## Sonar data files

Sonalyze can run on a set of individual Sonar data files (current-format JSON files as well as
older-format JSON and CSV files).  The format of a file is inferred from its name: current-format
JSON files have structured names on the form `<version>+<type>-<host>.json`, while older JSON files
have extension `.json` and older CSV files have names of the form `.csv`.  Since current Sonar only
produces the new-format JSON, older file types will not be discussed further.

```
$ sonalyze daemon file-name ...
```

In a Sonar data file, the embedded cluster name is normally ignored by Sonalyze, as the cluster name
is imposed externally by metadata or command-line switches.

TODO: The above description of file names is not comprehensive, and it is questionable whether we
are that structured for the data-files case, or only in the data-directory case.

TODO: We will allow for `-cluster-name` and `-cluster-aliases` to name the cluster.
