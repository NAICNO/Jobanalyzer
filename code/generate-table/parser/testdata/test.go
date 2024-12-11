// Print sysinfo for individual nodes.  The sysinfo is a simple config.NodeConfigRecord, without any
// surrounding context.  All fields can be printed, we just print raw data except booleans are "yes"
// or "no".
//
// If there are logfiles present in the input then we use those as a transient cluster of sysinfo.
//
// TODO: On big systems it would clearly be interesting to filter by various criteria, eg memory
// size, number of cores or cards.


//go:generate ../../../generate-table/generate-table -o node-table.go nodes.go

package main

/*TABLE node

package nodes

import (
	"go-utils/config"
	. "sonalyze/table"
)

%%

FIELDS *config.NodeConfigRecord

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


ELBAT*/

