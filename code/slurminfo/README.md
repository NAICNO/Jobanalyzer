A utility program that processes the output of an sinfo command into JSON for a cluster config file.

This is closely related to the `sysinfo` utility, see `../sysinfo`.

Basically, run `slurminfo` on the cluster and you'll get some JSON output describing the information
about the cluster that the program can extract using `sinfo`.

That information from `sinfo` is deficient in at least three ways:

- it does not contain information about non-slurm nodes that we may still care about, eg login nodes and
  interactive nodes.
- it does not reveal information about GPUs.
- it contains only the node name prefix (eg `c1-10`) but some HPC systems, such as Fox, will
  have a more complex node name, such as `c1-10.fox`, but not consistently so, and login and interactive
  nodes may also have complete (FQDN) node names.

All of those problems can be remedied by providing `slurminfo` with a "background" file which
provides the missing information.  For example,
`../../production/jobanalyzer-server/misc/fox.educloud.no/fox.educloud.no-background.json` is a stub
configuration file that contains the missing information in compact form: full descriptions for
interactive and login nodes, gpu information for the GPU nodes, and host name suffixes when
applicable for other nodes.
