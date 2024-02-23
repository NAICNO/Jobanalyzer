SoS - "Son of Sonalyze(d)"

This is a rethinking of the data store architecture, see bug 379.

## What it is

In this rethinking,

- the data store on disk is owned by a process, SoS, that is always running
- ingestion and query are clients that talk to this process
- the process caches data transparently in memory and manages disk, runs queries, produces
  results on some structured format
- probably the interface is a network socket or a unix-domain socket, I'm not sure which,
  taking queries in structured form and returning results in structured form

It would still have to work for the `sonard` use case where a local user creates local log data.  In
this case, SoS would be started on a fresh directory for data ingest, and then could later be
started on the directory for analysis.  The only thing that really changes is that `sonard` must
start SoS for ingest before starting sonar, and sonar output is piped into some utility that passes
the data to SoS.  It would be very transparent.

In the event that we need to change data formats from what we have now, there would be a utility to
replay existing log data into SoS.

## How it does it

The process is opened against a data store, which is (in current terms):

- the data/ tree with sonar data
- the state/ tree with cpuhog/deadweight state
- any config files
- ...

Then consider an obvious query:
```
   {op: "jobs",
    cluster: "ml",
    from: "2023-01-10",
    host:"ml[1-3]",
    user:"-",
    "no-gpu":"true"}
```

Here, "op" is mandatory, everything else is interpreted in the context of that.  For "jobs", there
is a mandatory "cluster" and then everything else is optional (though the defaults may not be what
they currently are for sonalyze).

This would slurp the data for the date range into memory and cache it.  It could slurp from raw CSV
files, or the data could be cooked on ingest.  Mostly it would already have the data.  Caching would
be a thing.  During slurping we might "lock in" records for the cluster/host/date range in question
(these are the only criteria that matter for slurping) and consider all non-locked records to be
fair game.  Caching would be per-host per-day so that re-reading an entire file is the primitive
operation.  Caching does *not* have to take into account the possibility that a file may have been
updated since this process also handles ingestion and would already know that; we can assume that
ingested records have been both appended to the file and appended to the cache.

Then it would create a view on those data by filtering by other criteria: host, user, no-gpu.

Then it would process those data by normal means and create the result set.

The main parameter is "size of cache" and how to keep track of how much we are using.  For the
naic-monitor node, a reasonable cache size may be around 24GB.

## Implementation

Internally there might be a tree that mirrors the directory tree of the data store: for sonar data
we have cluster/year/month/day/host-data-of-some-type, the leaf holds cached read-only data,
metadata, and even state.

Importantly for each file (probably) there is a mutex so that we can have multi-threaded access but
cache eviction, ingest, and use can all happen concurrently.

If we consider the data read-only and rely on GC, there need be no further coordination: once data
are obtained (probably as a slice of pointers to records) and the lock is relinquished, the cache
can be cleared for that day and host and we'll still have the data; new data can be added for that
day and host but we'll only have the data we got initially; others can also obtain the data.

The storage is decoupled from this logical structure.  The "state" data in a leaf could be in the
state/ subtree, and the event data could be in data/, and of course at the root of the tree there
would be per-cluster information that could be stored in a single file but exposed underneath
cluster/ in the tree.  (Maybe.)

A different view is that there are different databases, and where cluster/year/month/day/host is a
key into each database.  Maybe, `LOGFILE/<cluster>/<year>/<month>/<day>/<host>` is the shape of that
key.  This yields a flat keyspace (Redis).  Then the associated data, and how they are managed, are
application-specific.

Hard to know how much structure to impose.  It should be driven by two concerns:

- the needs of the program logic
- the possibility of using external software (eg Redis) to manage the data

I think it's probably good to have a specialized API and to consider how to map this onto a generalized
store behind the scenes.

```
type Store interface {

  // Purge everything, wait for all writing to finish, then mark the store as closed.

  func (s *Store) Close()

  // cluster must be a known cluster
  //
  // from must be !after to
  //
  // hostFilter can have bracketed ranges and embedded wildcards, if it is "" no host filtering
  // is performed

  func (s *Store) EventRecords(
      cluster string,
      from, to time.Time,
      hostFilter string,
  ) [][]*EventRecord
}

// This is the Store that uses the existing directory tree
type ClassicStore struct {
}

// rootDir must be the directory that holds data/, state/, cluster-config.json, and so on.
// The cache size 
func Open(rootDir path.Path, cacheSizeGiB uint) (*Store, error) {}

```
