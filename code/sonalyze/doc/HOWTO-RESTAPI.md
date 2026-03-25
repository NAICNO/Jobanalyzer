# The REST APIs

Sonalyze running in daemon mode presents two different REST-style APIs depending on options.

## Classical REST API

The first is the "classical" API described by the doc comment at the start of daemon/daemon.go and
in TECHNICAL.md in this directory.  This is currently always active in daemon mode.  The listen port
is 8087 by default but can be overridden with the `-port` argument.

Via this API, sonalyze processes top-level requests that take the form of sonalyze commands as
described for the non-daemon mode: `/jobs?user=x&from=y&to=z` corresponds directly to the command
`sonalyze jobs -user x -from y -to z`.

At this time, there is no way of extracting API information from this API; read TECHNICAL.md or the
source code.  (It would be good for this API to migrate toward a more modern REST API style, away
from the CGI style that it currently uses.)

Authentication is via HTTP basic authentication, ie, username/password headers.  The API checks that
the credentials allow access to the data by looking them up in an internal user database.

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
