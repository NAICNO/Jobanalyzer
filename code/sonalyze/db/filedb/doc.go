// filedb implements a simple database on top of either a list of files (this is always read-only and carries only
// one kind of data) or on a directory tree (this can be appendable and can carry all data types).
//
// When the files are in a directory tree the tree has subdirectory paths of the form YYYY/MM/DD for
// data.  At the leaf of each path are read-only data files for the given date, as follows.
//
// FILE NAME SCHEMES.
//
// Older data files follow these naming patterns:
//
//   - <hostname>.csv contain Sonar `ps` (ie sample) log data for the host
//   - sysinfo-<hostname>.json contain Sonar `sysinfo` system data for the host
//   - slurm-sacct.csv contains Sonar `sacct` data from Slurm for the given cluster
//
// Newer data files follow the naming pattern <version>+<type>-<originator>.json:
//
//   - 0+sample-<hostname>.json contains Sonar sample data for the host
//   - 0+sysinfo-<hostname>.json contains Sonar sysinfo data for the host
//   - 0+job-slurm.json contains Sonar slurm job data for the cluster (more general than sacct)
//   - 0+cluzter-slurm.json contains Sonar cluzter status data for the cluster
//
// In the latter scheme, `0` indicates version 0 of the new file format, for more see
// github.com/NordicHPC/sonar/util/formats/newfmt/types.go.  Extensions to the format are always
// backward compatible and require no new version number, however should there ever be reason to
// move to version 1, we can increment the file name scheme version number.  The intent is that the
// new file names have enough information to parse the contents correctly and index them coarsely
// within a given calendar day.
//
// For correctness, we assume host names cannot contain '+' (per spec they cannot).
//
// (In very old directories there may also be files `bughunt.csv` and `cpuhog.csv` that are state
// files used by some reports, these should be considered off-limits.  And note that in the old
// data, hosts cannot be named "slurm-sacct", or there will be a conflict between sacct job data and
// normal sample data.)
package filedb
