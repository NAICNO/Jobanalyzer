# Table generator

## How to use

In this directory, run:
```
go build
```

Then in a directory that has a file that declares a table (eg `sonalyze/cmd/nodes/nodes.go`) run:
```
go generate
```
to regenerate table code (`node-table.go` in that case).  Normally, `make generate` at some higher
level will do this for you.

It's helpful to run `make fmt` after generation since the generator emits code that is not perfectly
formatted.

The easiest way to understand the transformation is to look at the existing output files:
`../sonalyze/cmd/*/*-table.go`.

## Input form

See the documentation in ../go-utils/table.

## Output form

The output is Go code with the following definitions:

* Formatters and other field attributes will be in a map called <table-name>Formatters.
* Help will be in a multi-line string called <table-name>Help.
* Defaults will be a string called <table-name>DefaultFields.
* Aliases will be in a map called <table-name>Aliases.
* Any `MaybeFormatHelp` function will be as described under `HELP` above, and it will
  reference the other generated values by name, which will need to be manually supplied
  if not generated from the declaration.

The output form will likely change over time, but is currently compatible with what we've been
using previously.

No output is emitted for missing, optional sections.

## Appendices

### TODO

For a CommandType we could generate a parameter-less ValidateFormatArgs method that would be called
by the manually written Validate() function and could abstract away the names of the generated data.
Not sure if it's worth it.

For generalized query logic against the tables, we could generate query converters and predicates
(this is bug #714).

We handle a single interdependency, namely, fields that require a config file have the NeedsConfig
attribute set if the field name contains the word `Relative` or if the config attribute is set to
true.  But there are others: In jobs/print.go there is a dependency for Slurm data: Slurm fields
require an earlier step to join with the sacct table.  This is handled ad-hoc by having a table in
the earlier step that looks for slurm field names.  When new slurm names are added to the printout
they must also be added to the table.  This is flagged with a comment and is OK, but an attribute
bit might be better.  And so instead of the config / NeedsConfig attribute there may be a more
general flags attribute on fields: flag:"slurm", flag:"config|slurm".

### DONTDO

While it's tempting to put all the output definitions - formatters, predicates, summaries, help
text, whatever - together in some sort of struct, we should not do that.  Allowing them to be
independently created and created only if present in the spec allows the spec to omit definitions
and those definitions to be hand-written if need be.

