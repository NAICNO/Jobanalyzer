# Query language / query reflection

This is a sketch.


With issue 615 we cleaned up how printing was handled so that printing is always implemented by
reflecting on a table definition (sometimes with virtual rows).  The printing code is then generic,
there is no ad-hoc formatting, and when there is a special need it is handled in a non-ad-hoc way
through formatting directives / type annotations that are general.

We also have shared data source options and record filtering options, to the benefit of all, even if
a source of non-uniformity is that the defaults differ from command to command.  (Some have `-u -`
by default, others `-u $USER`.)

For our next trick, observe that a bunch of the tables have fairly ad-hoc query and row filtering
operations defined on them.  Some columns can be used for filtering, others not.  Names are sort of
arbitrary, and with printing introducing new (and better) column names, the filtering switches
really need to be upgraded too.

What if we could reflect on the table again and automatically define the row filtering options in
some uniform way?

The UX for this would be either that we define -min-X and -max-X fields switches for all X, which
fits with the current scheme and is easy to use for simple cases.  So this would be the first step:
In addition to `jobs` supporting `-min-cpu-avg` it would support `-min-CpuAvgPct` automatically (the
canonical name for that row).

The alternative or complementary UX is an expression language, so we could consider:
```
sonalyze jobs -q 'CpuAvgPct > 10 and Duration > 2h and SomeGpu'
```

(where `SomeGpu` is a virtual boolean column I guess).  Columns would need to be typed, but they
would need to be typed anyway.  This would be Fun(tm) but may simply be too much.

Speaking of which, boolean columns need an operator, so it might be `-if-SomeGpu` and
`-ifnot-SomeGpu` -- or maybe, the column is the switch, so `-some-gpu` really is just testing that
virtual column.  I don't know if that extends cleanly to everything we already have.
