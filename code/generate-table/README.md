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

### Lexical syntax

In the input file there should be one or more comment blocks with the following syntax:

```
/*TABLE <table-name>
...
%%
...
ELBAT*/
```

The the prefix - the text before `%%` - is copied verbatim to the output.  It must contain a package
directive and any imports referenced by type names in the table specification, and may not contain
anything else.  Note these imports are added by the table generator:

```
	"cmp"
	"fmt"
	"io"
	. "sonalyze/common"
	. "sonalyze/table"
```

The syntax is otherwise line-oriented.

Logical lines can be split across multiple input lines: if the line ends with `\` then the `\`, the
newline, and every whitespace character at the beginning of the following line are removed, and the
remaining part of the following line is attached to the current line.

Blank lines are allowed everywhere but are treated specially in the `HELP` section.  Comment lines
are treated as blank lines: comments can have some leading space followed by `#` and arbitrary text
up to the end of the line.

Note line joining is performed before stripping blank or comment lines.

### Grammar

The text after `%%` consists of a number of sections which must appear IN THE FOLLOWING ORDER, unless
optional and omitted.  A section starts with a header consisting of a tag (`FIELDS` etc) and maybe
some arguments.  The tag starts in column 1 and is all upper case as indicated.

#### FIELDS section

`FIELDS <type-name>`

The `type-name` is the type name of the record or a pointer to such a name, that is, the row type.
The header is followed by field definitions, one per line.  Each field definition is of the form:

```
  <field-name> <type-name> <attr>...
```
where the the type-name combines the representation type and formatting information.  The representation
type must match the underlying type of the data field exactly; no conversion is inserted.

Formatting information is pretty ad-hoc and can be anything we want it to be; it is specific to the
implementation of the data structures in sonalyze, and some are weird to capture existing output
conventions.

Type definitions that combine representation and non-standard formatting are defined in
`sonalyze/tables/data.go`.  A typical case is `DateTimeValue`, which is defined to be an `int64`
(representing seconds since epoch) that is to be formatted as `yyyy-mm-dd hh:mm` and which admits
modifiers to print it as iso time or as a second count.

Attributes describe how fields are accessed, provide help and aliases, and sometimes dictate the
generation of auxiliary data:

* `desc`  - description for -fmt help
* `alias` - comma-separated aliases
* `field` - the actual field name, if different
* `indirect` - the named field is a pointer, which may be nil; the field is to be
   fetched from the pointed-to structure
* `config` - boolean `"true"` or `"false"`: field requires a config file to work,
   false by default except for fields whose names contain the substring `Relative`

Each attribute is a `name : value` pair where the value must be a string literal, it can
contain escape characters.

#### GENERATE section

`GENERATE <record-type>`

Optional section.  Generate a Go struct definition from the field with the type name `<record-type>`.
This is useful in cases where the record is local to the package and a simple reflection of the
already-provided field list.  Fields that have a `field:` attribute are not emitted as part of the type.

#### SUMMARY section

`SUMMARY CommandType`

Optional section.  Leading and trailing blank lines are skipped, otherwise the payload is a brief
help text that is printed for `-h`.

#### HELP section

`HELP` | `HELP CommandType`

Optional section.  Leading and trailing blank lines are skipped, otherwise the payload is a brief
help text that is printed for `-fmt=help`.

If the `CommandType` is present then a standard `MaybeFormatHelp` method for `*CommandType` is also
generated.

#### ALIASES section

`ALIASES`

Optional section.  Each line is on the form `<name> <field-list>` where the `<name>` is the
alias being defined and the `<field-list>` is a comma-separated list of field names and
other aliases.

#### DEFAULTS section

`DEFAULTS <field-names>`

Optional section.  List of field names and/or aliases that comprise the default fields.

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

While it's tempting to put all the output definitions together in some sort of struct, allowing them
to be independently defined and defined only if present in the declaration allows the declaration to
omit definitions and those definitions to be hand-written if need be.

### Formal-ish grammar

The goyacc formalism in `parser/parser.y` is a little arcane so here's a summary, with terminals as
quoted strings, `^` meaning "start of line", `$` meaning "end of line", `*` meaning repeated zero or
more times, `?` meaning repeated zero or one time, and parens for grouping.  In particular, `^`
followed by a literal means "literally at the beginning of line", in other cases both `^` and `$`
allow arbitrary white space after and before, respectively.

```
Table     ::= ^ "/*TABLE" Ident $
              Prefix
              ^ "%%" $
              Fields
              Generate?
              Summary?
              Help?
              Aliases?
              Defaults?
              ^ "ELBAT*/"

Prefix    ::= TextLines

Fields    ::= ^ "FIELDS" TypeName $ Field*
Field     ::= ^ Ident TypeName Attribute* $
Attribute ::= Ident ":" String

Generate  ::= ^ "GENERATE" Ident $

Summary   ::= ^ "SUMMARY" Ident? $ TextLines

Help      ::= ^ "HELP" Ident? $ TextLines

Aliases   ::= ^ "ALIASES" $ Alias*
Alias     ::= ^ Ident Ident ("," Ident)* $

Defaults  ::= ^ "DEFAULTS" Ident ("," Ident)* $

TypeName  ::= "*" TypeName | "[" "]" TypeName | TypeName "." Ident | Ident

TextLines ::= <uninterpreted full lines of text terminated by FIELDS, GENERATE etc at beginning of line>

Ident     ::= /[a-zA-Z][-/%_a-zA-Z0-9]*/ but not FIELDS, GENERATE etc at beginning of line
String    ::= /"([^\\"]|\\.)*"/
```

### The goyacc tool

If `parser/parser.y` needs to be changed for bug fixes or amendments, then `parser/parser.go` needs to
be regenerated.  Running `go generate` in the parser subdirectory is enough.  For this to work,
`goyacc` must be installed and in the path.  To install it in `~/go/bin`:
```
go install golang.org/x/tools/cmd/goyacc@master
```
