# Database data representations

These are data representations of data coming from the database (irrespective of the type of
database).  They are mostly flat (ie simple structs with primitive values) but some contain lists of
elements, and some of the list elements are themselves structured, so in practice we have 1D, 2D and
3D data here - not quite tabular.

The data that are represented here may be held in a cache and should be considered *strictly*
read-only.  After creation, the data may be accessed concurrently without locking from many threads
and must not be written by any of them.

As a general rule, if a data type is particularly large and frequent (currently Sample data of all
kinds and Slurm job data fit this), then it should be pointer-free so as to play nice with the Go
garbage collector.  Most of these data will have many string fields however.  To square that circle,
the strings are represented as Ustr, which are indices into an internal string table, see
../../common/ustr.go.

(The CpuSamples and GpuSamples structures currently contain native strings as well as arrays and
other pointer types, see comments there.  This may be laborsome to fix.)

If a data type is not particularly large and frequent, then it's normally best to keep string fields
as strings.
