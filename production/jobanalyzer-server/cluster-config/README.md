# Cluster configuration

There is one subdirectory for each cluster that sends data.  The name of the subdirectory is the
canonical cluster name, eg `mlx.hpc.uio.no`, `fox.educloud.no`, `saga.sigma2.no`.

In each cluster subdirectory there must be a file `cluster-config.json` that specifies what the
cluster looks like.  This file is documented below.

In this directory there may optionally be a file called `cluster-aliases.json`.  The format of this
file is an array of objects; each object has keys `alias` and `value`, where `alias` is the alias
we're defining and `value` is the canonical name of the cluster, eg,

```
[{"alias":"ml", "value":"mlx.hpc.uio.no"},
 {"alias":"fox", "value":"fox.educloud.no"}]
```

**NOTE:** This is not a good structure for configuration information; it is a consequence of some
old decisions.  It will likely change.

## The format of cluster-config.json

FIXME.  For now, this is specified by `code/go-utils/config/config.go` in the Jobanalyzer source code.


## Clippings

//
// Cluster names and aliases:
//
//  Cluster names are the aliases of login nodes (fox.educloud.no for the UiO Fox supercomputer) or
//  synthesized names for a group of machines in the same family (mlx.hpc.uio.no for the UiO ML
//  nodes cluster).
//
//  The cluster alias file is a JSON array containing objects with "alias" and "value" fields:
//
//    [{"alias":"ml","value":"mlx.hpc.uio.no"}, ...]
//
//  so that the short name "ml" can be used to name the cluster "mlx.hpc.uio.no" in requests.
