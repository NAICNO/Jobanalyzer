# The REST APIs

Sonalyze running in daemon mode can present a REST-style API to local and remote clients.  This API has
multiple versions.

To enable the REST APIs, pass the `-rest-api` switch to the daemon.  It takes an interface value,
frequently something like `127.0.0.1:8888`.  Additionally, individual APIs must be enabled with
`-v0`, `-v1`, and `-v2` (several can be enabled at the same time).

All requests to the API start with `/api/vK` where K is a version number, currently 0, 1, or 2.  For
example, `/api/v2/cluster/my.cluster.name/nodes/info`.

Documentation is available via `https://127.0.0.1:8888/openapi.yaml` (or `openapi.json`) when the
server is running on that interface, this will describe the v0, v1, and v2 versions.

Authentication for v0 and v1 is via HTTP basic authentication, ie, username/password headers.  The
API checks that the credentials allow access to the data by looking them up in an internal user
database - see [TECHNICAL.md](TECHNICAL.md).  There are separate authentication realms for insertion
and lookup.

## REST API v0

The v0 API is close to the "classical" Sonalyze REST API and follows the command line syntax
closely.  A GET request to `/api/v0/jobs?...` will be a jobs query, for example, and the query
arguments are the same as for the jobs command at the command line.  Parameter names are always the
long parameter names for sonalyze (`user` not `u`, `from` not `f`).

The returned output is the raw output from sonalyze, whether for success or error, encoded as a JSON
string (which must be parsed by the consumer).  A successful run yields 2xx and an error yields 4xx
or 5xx.

Via this API, sonalyze processes top-level requests that take the form of sonalyze commands as
described for the non-daemon mode: `/api/v0/jobs?cluster=c&user=x&from=y&to=z` corresponds directly
to the command `sonalyze jobs -cluster c -user x -from y -to z`.

For convenience, Sonalyze, with the -remote option, translates a "local" command to a v0 API call
(with authentication), but there's nothing special about this: under the hood it is currently just a
curl invocation of the translated request.  See MANUAL.md.

## REST API v1

The v1 API is intended to follow the v0 API, with the difference that where the v0 API always
returns a JSON string for all output types, the v1 API will return plain JSON data.  Additionally,
the v0 API (following the classical Sonalyze REST API), when it returns JSON encoded data (with
`-fmt=json`), encodes all field values as strings.  The v1 API will use natural encodings.

At the moment, the v1 API presents only a data insertion API that is new (the old v0 data insertion
API being obsoleted since those data formats are no longer supported).  A POST to
`/api/v1/insert/<type>` will present data of the given `<type>` (sample, sysinfo, job, cluster) for
insertion in the data store.  The data must be presented as JSON and have the form defined by the
Sonar data format spec.

## REST API v2 - the slurm-monitor REST API

The v2 API is a *partial* and *probably buggy*
[slurm-monitor](https://github.com/2maz/slurm-monitor) style API.

Authentication for this API is via OAUTH and is in principle set up so that only a super-user can
query data for other users than themselves.  Since this authentication scheme is poorly integrated
with Sonalyze at this time, the switch `-v2` must be passed to the daemon to enable this API.
