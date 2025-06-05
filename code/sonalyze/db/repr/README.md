These are data representations of data coming from the database (irrespective of the type of
database).  They are mostly flat (ie simple structs with primitive values) but some contain lists of
elements, and some of the list elements are themselves structured, so in practice we have 1D, 2D and
3D data here - not quite tabular.

The data that are represented here may be held in a cache and should be considered *strictly*
read-only.  After creation, the data may be accessed concurrently without locking from many threads
and must not be written by any of them.
