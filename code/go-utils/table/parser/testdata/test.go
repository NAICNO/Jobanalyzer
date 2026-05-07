// Test input derived from the `node` verb.

package main

/*TABLE node

package nodes

import (
	. "sonalyze/table"
    "sonalyze/db/repr"
)

%%

FIELDS *repr.NodeSummary

 # Note the CrossNodeJobs field is a config-level attribute, it does not appear in the raw sysinfo
 # data, and so it is not included here.

 Timestamp   string desc:"Full ISO timestamp of when the reading was taken" alias:"timestamp"
 Hostname    string desc:"Name that host is known by on the cluster" alias:"host"
 Description string desc:"End-user description, not parseable" alias:"desc"
 CpuCores    int    desc:"Total number of cores x threads" alias:"cores"
 MemGB       int    desc:"GB of installed main RAM" alias:"mem"
 GpuCards    int    desc:"Number of installed cards" alias:"gpus"
 GpuMemGB    int    desc:"Total GPU memory across all cards" alias:"gpumem"
 GpuMemPct   bool   desc:"True if GPUs report accurate memory usage in percent" alias:"gpumempct"

GENERATE NodeSomething

SUMMARY NodeCommand

Display self-reported information about nodes in a cluster.

For overall cluster data, use "cluster".  Also see "config" for
closely related data.

The node configuration is time-dependent and is reported by the node
periodically, it will usually only change if the node is upgraded or
components are inserted/removed.

HELP NodeCommand

  Extract information about individual nodes on the cluster from sysinfo and present
  them in primitive form.  Output records are sorted by node name.  The default
  format is 'fixed'.

ALIASES

  default  host,cores,mem,gpus,gpumem,desc
  Default  Hostname,CpuCores,MemGB,GpuCards,GpuMemGB,Description
  all      timestamp,host,desc,cores,mem,gpus,gpumem,gpumempct
  All      Timestamp,Hostname,Description,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct

DEFAULTS default


ELBAT*/

