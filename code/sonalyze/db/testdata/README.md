# What is this?

This is a full test database for sonalyze.  It should eventually have files of all types, for
multiple node names, and it should also have eg bughunt.csv and other red herrings that need to be
handled.

The files should ideally hold real data, though some of the cluster/node names may be changed and
the user/account names should be redacted or changed.  Some file should be empty, some files not.

There are redundancies here relative to what's elsewhere in the tree, but so be it.  The test cases
here are *primarily* for the use by the db layer, using file storage.

As per normal, `cluster-config` and `data` together make up a persistent, date-indexed and
host-indexed database.

Files for a transient database are TBD.
