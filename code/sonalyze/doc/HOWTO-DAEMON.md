# Running the sonalyze daemon

The sonalyze daemon sits on top of a data store and responds to queries or requests to insert data.
Alternatively to receiving insertion requests, it can subscribe to a Kafka broker to receive data to
insert.  Insertion (from any source) is only enabled when the data store is a "jobanalyzer
directory", see [HOWTO-DATA-SOURCE.md](HOWTO-DATA-SOURCE.md) and below.  When the data store is a
timescaledb database, insertion must be performed by the listeners of
[slurm-monitor](https://github.com/2maz/slurm-monitor), and in the remaining cases the data store is
read-only and must not be modified while sonalyze is using it.

To create a Jobanalyzer directory `D`, create `D` and `D/data`, and use `D` as the argument to the
`-jobanalyzer-dir` argument:

```
sonalyze daemon -jobanalyzer-dir D ... &
```

When sonalyze receives data to insert for a cluster `my.cluster`, it will store the data under
`D/data/my.cluster`, in a structure described in [HOWTO-DATA-SOURCE.md](HOWTO-DATA-SOURCE.md).  It
will create directories as needed.

Optionally, there can also be a directory `D/cluster-config` with JSON files in it adding some
metadata to the cluster data.  A file `D/cluster-config/my.cluster-config.json` has config data for
the `my.cluster` cluster.  The JSON format is an object with keys `name` for the cluster name
(`my.cluster`), `description` for a human-readable description of the cluster, and `aliases` for a
list of aliases for the cluster.  Here's a real one:

```
{
    "name":"fox.educloud.no",
    "aliases":["fox"],
    "description":"UiO 'Fox' supercomputer",
}
```

## Data acquisition by kafka

The Sonalyze daemon can be told to access a Kafka broker to acquire data.  Run the daemon with the
`--kafka` option to provide a broker address (plaintext only, so the broker should be running on the
local system).  The acquisition will happen for every cluster known to the daemon, for all four data
types - `cluzter`, `sacct`, `sample`, and `sysinfo`.

The Kafka acquisition service is currently considered experimental, but works well enough.

