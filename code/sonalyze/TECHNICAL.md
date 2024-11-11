# `sonalyze` technical notes (internal use mostly)

## Authentication and authorization

The following pertains to sonalyze running as a daemon.


### Query authentication

#### Design

Jobanalyzer user identities U are global as opposed to cluster-local, this is a necessity for the
web dashboard that provides views to several clusters' data on the same domain.

Each user has one password P associated with it. Passwords are stored in the server's secrets - any
kind of identity management can be used can in principle be used.

A request comes in to the HTTP server with username/password U/P (encoded as an HTTP header asking
for the simple HTTP authentication).  Identity checking is performed by the daemon before the
request is acted on.  If the check fails the request is rejected.

#### Implementation

The file `$JOBANALYZER/secrets/sonalyzed-auth.txt` is a list of `username:password` pairs, passwords
are stored in plaintext.  New users are added at the end.  The file is read once on daemon startup.
The directory, and indeed the file, should be umask 0600.

(The design dates back to the earliest days but works well enough for now.  Plaintext passwords are
OK-ish since the passwords only provide query authorization for sonalyze, and not access to the
system in general, but something hashed would be better.)

Any other kind of identity management would work OK here; for authorization to work (below) we just
need the user name.


### Query authorization

#### Design

**NOTE:** This has not been implemented.  At the moment, every user has superuser privileges on all
clusters, ie, they can see all data for all users.

Jobanalyzer user identities U are global, but users may have different identities on different
clusters (at UiO, users may be ec-robert@fox.educloud.no on Fox and robert@uio.no elsewhere), and
may have different identities still on the Sigma2 supercomputers.  Most users should only be
authorized to see their own records, but from all the clusters they have accounts on.  This
necessitates a mapping from jobanalyzer user ID to a cluster user ID to use when filtering records,
and the enforcement of scrubbing data by privileges.

Additionally, sysadmins and user support staff are different, they should be able to see all users'
records.

Jobanalyzer maintains a mapping for each cluster from the global ID U to the cluster's local
identity L or to a flag indicating a superuser.  The mapping can be thought of as a table of
(cluster, user, local-id) rows.  The "user" is the global jobanalyzer user name, and "local-id" is
the name the user is known by on this cluster or the special flag "-" indicating a superuser on this
cluster.

The daemon consults the user mapping and ensures that every query (for which it makes sense) carries
a "-user" record filter that selects records according to the mapped local-id.

(There's an assumption in all of that that local-ids are never reused on the various clusters.)

This works as follows.  On cluster C, consult the mapping `{(C,U) -> L}`; if there is no such (C,U)
then error out.  L is the local-id for U on cluster C.  Then the command is augmented:

* If there is not a -user K argument given in the command then -user L is added.

* If there is a -user K argument then either L==K or L==- (note both can be true), otherwise we
  error out.

The effect of this is (should be) that any U with L==- can see anything, and any U with L other than
"-" can see only the records for L.

#### Implementation

In keeping with the rest of the database structure the file
`$JOBANALYZER/users.conf.d/$CLUSTER-users.json` is an array of JSON objects with keys "user" and
"local-id", implementing the `{(C,U) -> L}` mapping described above.  The file is read into memory
once at startup, like all the other config files.

We need to ensure that every query for which it makes sense carries "-user".  Which queries are
these?  In principle it is every query based on Sonar sample data or Slurm sacct data, since these
carry user information.  That excludes only `cluster`, `config`, and `node`.


### Upload (`add`, etc) authentication and authorization

#### Design

For upload (the `add` command and its aliases) authentication first checks that the provided
username and password exist in the upload authorization data base.  A request comes in to the HTTP
server with username/password U/P (encoded as an HTTP header asking for the simple HTTP
authentication).  Identity checking is performed by the daemon before the request is acted on.  If
the check fails the request is rejected.

Then an authorization check is made that the provided username matches the cluster argument given to
`add` - different clusters have different upload passwords, which are secrets they own.  This check
makes it impossible for somebody with only the password for one cluster to spoof data for another
cluster.

#### Implementation

The file `$JOBANALYZER/secrets/exfil-auth.txt` is a list of `clustername:password` pairs, passwords
are stored in plaintext.  New clusters are added at the end.  The file is read once on daemon
startup.  The directory, and indeed the file, should be umask 0600.

(Also see comments for the implementation of user authentication, above.)

Any other kind of identity management would work OK here; for authorization to work we just need the
cluster name.


## REST protocol

The `sonalyze` server can be accessed via a simple REST protocol.

For the following to make sense you should be familiar with data model and command line syntax, see
MANUAL.md.

### Background

The protocol was originally developed to allow a command-line invocation of `sonalyze` on the
"sending" side (some endpoint computer) to be transmitted to the `sonalyze` daemon on the
Jobanalyzer server on the "receiving" side and there to be translated into an actual command line
for a server-side invocation of `sonalyze`.  A small amount of translation has to take place:
`-cluster` is required on the sending side to imply the values for `-data-dir` and `-config-file` on
the receiving side.

Later it became clear that this was just a general REST protocol that could be handled by standard
HTTP mechanisms, and so the protocol became open for use by eg the dashboard and for command-line
scripts that wanted to bypass `sonalyze` for some reason.

On the server side, the `sonalyze` daemon no longer recursively spins up a new `sonalyze` process,
but the internal logic is still that a command line is built from the query parameters for the
invocation of a request handler.

The protocol is defined jointly by `sonalyze/command/reify.go` (which constructs requests) and
`sonalyze/daemon/perform.go` (which parses them and turns them into command lines that are then
processed in the normal manner).

### Definition

#### Syntax

The request URL is always `<verb>?<query>`.  The Jobanalyzer HTTP server is generally set up so that
these requests are top-level: `/<verb>?<query>`.

The `<verb>` is one of the verbs accepted by sonalyze on the command line: `add`, `cluster`,
`config`, `jobs`, `load`, `node`, `uptime`, `profile`, `sacct`, `sample` (aka `parse`), `metadata`,
`top`.  In addition two special verbs are accepted for backward compatibility, `sonar-freecsv` and
`sysinfo`; these are aliases for `add -sample` and `add -sysinfo` respectively.

For `add`, `sonar-freecsv`, and `sysinfo` the HTTP operation must be `POST` and the payload to be
inserted into the database is the body of the the request.

For the other verbs the HTTP operation must be `GET`.

Query parameters are always URL-encoded and separated by `&` in the normal way.

Query parameters that carry values are specified as `name=value`, with the value presented in the
syntax required by the sonalyze verb in question, eg `host=gpu-[1,4-8],c[1,2]-[8,9]` or
`user=frobnitz`.

Value-less query parameters (flags) are a special case.  For historical reasons described in the
code, these carried "true" values that were always encoded as `xxxxxtruexxxxx` (a string assumed
never to occur in any other context - it's not a user name, host name, or other value), e.g.,
`some-gpu=xxxxxtruexxxxx`.  That encoding remains valid and will remain valid, but is no longer
necessary.  Currently, the value must be a boolean value, `true` or `false` (`some-gpu=true`).
Passing a parameter with a `false` value is redundant, and it would be better to omit the parameter.
Also, while many "boolean" values are accepted by the current flags parser, please stick to `true`
or `false` if you use a value at all.

#### Parameters and their values

By and large, all parameters accepted by `sonalyze` are accepted as query parameters, with the same
name and syntax for both the parameter names (without the leading `-`) and parameter values.  Try
`sonalyze help` or `sonalyze <verb> -h` for more information, read MANUAL.md in this directory, or
examine the code.

Some parameters are scrubbed by `sonalyze` when it constructs the remote request, and various
consistency checks are applied.  For example, `-remote` usually requires `-cluster` (and
`-auth-file` can be used with these) and are exclusive with `-data-dir` and `-- logfile...`.  `-v`
is not forwarded (a remote query executed with `-v` will print the final URL).

When constructing a query by hand, there are no client-side restrictions, but the server will
quietly ignore the query parameters `cpuprofile`, `data-dir`, `data-path`, `remote`, `auth-file`,
`config-file`, `v`, `verbose`, and `raw`.

The `cluster` parameter is required except for with the `cluster` verb.

The server will infer `config-file` and `data-dir` from `cluster`.

#### Limitations

Query URLs are limited in length by parts of the infrastructure (and possibly by underlying web
standards).  Very long lists of e.g. job IDs used for selection criteria may result in errors being
reported.  The workaround for this is currently to either run multiple queries and merge the
results, or to query less selectively and filter the data on the client side.

(The "long list of selection criteria" is unfortunately a common scenario because the criteria may
be extracted from a broad query of Slurm jobs data which are then filtered locally, forming a long
list of user or job IDs to be used in querying Sonar data.)


## Data field vocabulary

(This is evolving.)

Field naming is pretty arbitrary and it is not going to be cleaned up right now.  For the most part
we can fix things over time through the use of aliases.

"Old" names such as "rcpu", "rmem" should probably not be used more than absolutely necessary,
ideally all new names are fairly self-explanatory and not very abbreviated.

Contextuality is important to make things hang together.  The precise meaning of the field must be
derivable from name + context + type + documentation, ideally from name + context + documentation
since the user may not have access to the type.  Name + documentation must be visible from -fmt
help, and context is given by the verb.  (Hence plain "Name" in the cluster info is not as bad as it
looks because it is plain from context and documentation that we're talking about the cluster name;
"Clustername" might have been better, but not massively much better.)

Spelling standards that we should follow when we have a chance to (re)name a field:

* Cpu, Cpus not CPU, CPUS, CPUs
* Gpu, Gpus not GPU, GPUS, GPUs
* GB not GiB, the unit is 2^30
* MB not MiB, the unit is 2^20
* KB not KiB, the unit is 2^10
* JobId not JobID
* Units on fields that can have multiple natural units, eg, ResidentMemGB not ResidentMem

(And yet there may be other considerations.  The sacct table names such as UsedCPU and MaxRSS are
currently the way they are because those are the names adopted by sacct.  But on the whole it'd
probably be better to follow our own naming standards and explain the mapping in the documentation.)
