# Response generator

This generates Go response boilerplate code (for use by REST API responses) from table definitions
in the code.

## How to use

In this directory, run:
```
go build
```

Then in a directory that has a file that declares a response (eg `sonalyze/daemon/api1/clusters/clusters.go`) run:
```
go generate
```

to regenerate response code (`respond.go`).  Normally, `make generate` at some higher level will do
this for you.

It's helpful to run `make fmt` after generation since the generator emits code that is not perfectly
formatted.

The easiest way to understand the transformation is to look at the existing output files:
`../sonalyze/daemon/api1/*/respond.go`.

## Input form

See the documentation in ../go-utils/table.

## Output form

To be documented; look at the existing outputs for a rough idea.

