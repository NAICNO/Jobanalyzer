# `sonalyze` manual

## Typical usage patterns

`sonalyze` is a database manager and query interface for HPC cluster time series data coming from
[Sonar](https://github.com/NordicHPC/Sonar).  Use it to add data and to query, aggregate, and export
data.

Sonalyze can operate on directory trees of Sonar data, either read-only or in an appending mode.
This is its primary historial function.  But it can also operate on read-only file lists (for
testing) or on a timescaledb database maintained by a sibling tool,
[Slurm-monitor](https://github.com/2maz/slurm-monitor).

Sonalyze can be used in one-shot / command-line mode against a data store, and in this case it
performs a single operation and exits.  More commonly it will run in a daemon mode where it provides
a REST API for queries.  In the daemon mode case it can amortize the cost of data access over
multiple queries.

The line between one-shot mode and the daemon mode is barely visible: When a sonalyze daemon is
running on a server, sonalyze can be run on a remote workstation in its one-shot mode to build and
submit REST queries to the server and to format and present the output.

This manual presents the one-shot / command-line mode as it is used to access local data or to build
remote queries transparently.

For information about how to set up and run the daemon, start in [HOWTO-DAEMON.md](HOWTO-DAEMON.md).

For information about how to access the REST API directly, start in
[HOWTO-RESTAPI.md](HOWTO-RESTAPI.md).

## Input source

In one-shot mode, running on local data, the data source can be specified in various ways, and apart
from restrictions on inserting data, all operations work the same way on all the sources.  See
[HOWTO-DATA-SOURCE.md](HOWTO-DATA-SOURCE.md) for more about data sources.

## Data model at a glance

Notionally there are these tables:

* `cluster` is the table of clusters known to the DB manager, with their aliases and descriptions
* `cluzter` is the table of Slurm partition information
* `config` is the table of per-node configuration information
* `sysinfo` is the table of per-node system information extracted by Sonar on each node every day
* `sacct` is the table of completed Slurm jobs on the cluster
* `sample` is the table of Sonar samples, collected by Sonar on each node

The sonalyze operations (next section) insert data into a table, extract raw data from a table, or
generate cooked data from one or more tables.

## Summary of operations

In summary,

```
sonalyze <command> [options]
```

It is easiest to just run `sonalyze help`, the on-line help is supposed to sufficient (and current).
Every operation accepts options to request help (`sonalyze jobs -h`) or output formatting help
(`sonalyze jobs -fmt=help`).

Command line parsing allows for `--option=value`, `--option value`, and also to spell `-option` with
a single dash in either case.  It is however not possible to run single-letter options together with
values, in the manner of `-f1w`; it must be `-f 1w`.

(Older prose about operations is in [STALE.md](STALE.md), and may be helpful, or may be wrong.)

## Remote access

Sonalyze can perform transparent REST calls to a remote Sonalyze daemon.

To do so, use the `--remote` argument to provide a URL for the remote host, the `--cluster` argument
to name the cluster for which we want data, and the `--auth-file` argument to authenticate yourself:

```
$ sonalyze jobs --remote http://some.host.no:8087 \
                --cluster ml \
                --auth-file some-file.netrc \
                -f 20w \
                -u - \
                --some-gpu \
                --host ml8
```

The auth-file should be kept secret; the identity must be provided by an admin.  See the
[HOWTO-DAEMON.md](HOWTO-DAEMON.md) and [HOWTO-AUTH.md](HOWTO-AUTH.md) for how to set that up.

The auth-file must be on the .netrc format (but must have a single line of text).  The .netrc format
is `machine <machine> login <username> password <password>`.

In the case of an `add` operation you need to supply an identity that is allowed to insert data,
typically these are different from identities that can query.

## Ini file

The file `$HOME/.sonalyze` can contain defaults for some arguments.  The file is a line-oriented
sequence of sections starting with a header and containing definitions, environment variables are
expanded.  Comment lines are allowed.  A typical file:

```
# Standard setup

[data-source]
remote=https://naic-monitor.uio.no
auth-file=$HOME/.ssh/sonalyzed-auth.netrc
```

Currently the only section is `[data-source]` and valid keys are `remote`, `auth-file`, `cluster`,
`data-dir`, `from`, and `to`.

## Environment variables

Remote credentials can be provided in the `SONALYZE_AUTH` variable, this takes a `username:password` and
overrides whatever values are used with `-auth-file` or in `$HOME/.sonalyze`.

