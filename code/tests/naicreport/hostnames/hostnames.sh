output=$($NAICREPORT hostnames -sonalyze $SONALYZE -to 2024-10-21 -- sysinfo*json)
CHECK hostnames_basic '["ml1.hpc.uio.no","ml2.hpc.uio.no","ml3.hpc.uio.no","ml4.hpc.uio.no","ml6.hpc.uio.no","ml7.hpc.uio.no","ml9.hpc.uio.no"]' "$output"
