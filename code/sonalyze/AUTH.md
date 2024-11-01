# Authentication and authorization

The following pertains to sonalyze running as a daemon.


## Query authentication

### Design

Jobanalyzer user identities U are global as opposed to cluster-local, this is a necessity for the
web dashboard that provides views to several clusters' data on the same domain.

Each user has one password P associated with it. Passwords are stored in the server's secrets - any
kind of identity management can be used can in principle be used.

A request comes in to the HTTP server with username/password U/P (encoded as an HTTP header asking
for the simple HTTP authentication).  Identity checking is performed by the daemon before the
request is acted on.  If the check fails the request is rejected.

### Implementation

The file `$JOBANALYZER/secrets/sonalyzed-auth.txt` is a list of `username:password` pairs, passwords
are stored in plaintext.  New users are added at the end.  The file is read once on daemon startup.
The directory, and indeed the file, should be umask 0600.

(The design dates back to the earliest days but works well enough for now.  Plaintext passwords are
OK-ish since the passwords only provide query authorization for sonalyze, and not access to the
system in general, but something hashed would be better.)

Any other kind of identity management would work OK here; for authorization to work (below) we just
need the user name.


## Query authorization

### Design

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

However, there's a complication.  Suppose a user is a superuser on a cluster, hence L=-.  Suppose
there is no user argument.  Suppose the command is "jobs".  In this case, the rules would add
`-user=-`, but this would be wrong because it changes the default for the command from "me" to
"everyone".  Worse, for "jobs", is that other arguments may change this setting: `--zombie` changes
the default from "me" to "everyone"; `-job` and `-exclude-user` ditto.  And this logic is scattered
a bit here and there.  So the first rule above is amended:

* If there is not a -user K argument given in the command then -user L is added unless the default
  computed user for the command is "everyone".

There can be multiple explicit user arguments.  This is handled by the second rule above.

### Implementation

In keeping with the rest of the database structure the file
`$JOBANALYZER/users.conf.d/$CLUSTER-users.json` is an array of JSON objects with keys "user" and
"local-id", implementing the `{(C,U) -> L}` mapping described above.  The file is read into memory
once at startup, like all the other config files.

We need to ensure that every query for which it makes sense carries "-user".  Which queries are
these?  In principle it is every query based on Sonar sample data or Slurm sacct data, since these
carry user information.  That excludes only `cluster`, `config`, and `node`.

(On second thought, not sure that one file per cluster is a great idea.  But this is one of those
things that *really should be a proper database* anyway.  User permissions are not append-only
like the rest of our data, we want transactions etc.)

## Upload (`add`, etc) authentication and authorization

### Design

For upload (the `add` command and its aliases) authentication first checks that the provided
username and password exist in the upload authorization data base.  A request comes in to the HTTP
server with username/password U/P (encoded as an HTTP header asking for the simple HTTP
authentication).  Identity checking is performed by the daemon before the request is acted on.  If
the check fails the request is rejected.

Then an authorization check is made that the provided username matches the cluster argument given to
`add` - different clusters have different upload passwords, which are secrets they own.  This check
makes it impossible for somebody with only the password for one cluster to spoof data for another
cluster.

### Implementation

The file `$JOBANALYZER/secrets/exfil-auth.txt` is a list of `clustername:password` pairs, passwords
are stored in plaintext.  New clusters are added at the end.  The file is read once on daemon
startup.  The directory, and indeed the file, should be umask 0600.

(Also see comments for the implementation of user authentication, above.)

Any other kind of identity management would work OK here; for authorization to work we just need the
cluster name.
