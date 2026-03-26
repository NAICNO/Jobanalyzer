# `sonalyze` technical notes (internal use mostly)

This is probably quite stale and a number of points are explained better in other docs.




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
