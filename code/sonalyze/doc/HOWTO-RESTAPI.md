# The REST APIs

Sonalyze running in daemon mode presents two different REST-style APIs depending on options.

## Classical REST API

The first is the "classical" API described by the doc comment at the start of daemon/daemon.go and
below as "REST API v0".  This API is currently always active in daemon mode.  The listen port is
8087 by default but can be overridden with the `-port` argument.

Via this API, sonalyze processes top-level requests that take the form of sonalyze commands as
described for the non-daemon mode: `/jobs?user=x&from=y&to=z` corresponds directly to the command
`sonalyze jobs -user x -from y -to z`.

At this time, there is no way of extracting API information from this API or generating client code
from such a spec; read below or the source code.  (Almost certainly this API will be reimagined as a
proper REST API implementation before long, mitigating these problems.)

Authentication is via HTTP basic authentication, ie, username/password headers.  The API checks that
the credentials allow access to the data by looking them up in an internal user database - see
[TECHNICAL.md](TECHNICAL.md).  There are separate authentication realms for insertion and lookup.

For convenience, Sonalyze, with the -remote option, translates a "local" command to an API call in
the former style (with authentication), but there's nothing special about this: under the hood it is
currently just a curl invocation of the translated request.  See MANUAL.md.

## Slurm-monitor REST API

The second API is a partial [slurm-monitor](https://github.com/2maz/slurm-monitor) style API.  This
is a proper REST API built on modern infrastructure.  It is off by default but is enabled with the
`-rest-api` argument to the daemon, which takes an interface value, frequently something like
`127.0.0.1:8888`.  All requests start with `/api/v2` - `/api/v2/cluster/my.cluster.name/nodes/info`.

Documentation is available via `https://127.0.0.1:8888/openapi.yaml` (or .json) when the server is
running on that interface.

Authentication for this API is via OAUTH and is in principle set up so that only a super-user can
query data for other users than themselves.

## REST API v0

For the following to make sense you need to be familiar with data model and command line syntax, see
MANUAL.md.  The API mimics the sonalyze command line, and (except for some proscribed parameters)
every argument is effectively passed through to a recursive invocation of sonalyze on the server.  A
request `/jobs?user=x&from=y&to=z` becomes the command `sonalyze jobs -user x -from y -to z`.

The request URL is always `<verb>?<query>`.  The Jobanalyzer HTTP server is generally set up so that
these requests must be top-level: `/<verb>?<query>`.

The `<verb>` is one of the verbs accepted by sonalyze on the command line, run `sonalyze help` for
the full list.

For the `add` verb the HTTP operation must be `POST` and the payload to be inserted into the
database is the body of the the request.

For the other verbs the HTTP operation must be `GET`.

Query parameters are always URL-encoded and separated by `&` in the normal way.

Query parameters that carry values are specified as `name=value`, with the value presented in the
syntax required by the sonalyze verb in question, eg `host=gpu-[1,4-8],c[1,2]-[8,9]` or
`user=frobnitz`.

Value-less query parameters (flags) are a special case.  The value must be a boolean value, `true`
or `false` (`some-gpu=true`).  Passing a parameter with a `false` value is redundant, and it would
be better to omit the parameter.  Also, while many "boolean" values are accepted by the current
flags parser, please stick to `true` or `false` if you use a value at all.

By and large, all parameters accepted by `sonalyze` are accepted as query parameters, with the same
name and syntax for both the parameter names (without the leading `-`) and parameter values.  Try
`sonalyze help` or `sonalyze <verb> -h` for more information, read MANUAL.md in this directory, or
examine the code.

Some parameters are scrubbed by `sonalyze` when it constructs the remote request, and various
consistency checks are applied.  Errors are signalled for bad behavior.

When constructing a query by hand, however, there are no client-side restrictions, but the server
will quietly ignore the query parameters `cpuprofile`, `data-dir`, `data-path`, `remote`,
`auth-file`, `config-file`, `v`, `verbose`, and `raw`.

The `cluster` parameter is required except for with the `cluster` verb.

The server will infer `config-file` and `data-dir` from the `cluster` parameter, as appropriate.

Query URLs are limited in length by parts of the infrastructure (and possibly by underlying web
standards).  Very long lists of e.g. job IDs used for selection criteria may result in errors being
reported.  The workaround for this is currently to either run multiple queries and merge the
results, or to query less selectively and filter the data on the client side.
