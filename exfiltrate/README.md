# Exfiltrate

`exfiltrate` - data forwarder for on-node agents.

## How to run

See comment block at the beginning of exfiltrate.go for all information.

## Future developments

In the future, I expect we'll want to do at least some other source formats:

"sysinfo/json"
   The sysinfo program figures out the machine configuration and prints a json package with the
   data

"diskstatus/json"
  The diskstatus program prints the fullness of disks

"slurm-seff/json"
  Slurm data from "seff"

In the future, at least some type of binary format is expected for the output.

In the future, other protocols (mqtt?) are expected for the target
