output=$($NAICREPORT hostnames -sonalyze $SONALYZE -- sysinfo*json)
CHECK hostnames_basic '["ml1.hpc.uio.no","ml2.hpc.uio.no","ml3.hpc.uio.no","ml4.hpc.uio.no","ml6.hpc.uio.no","ml7.hpc.uio.no","ml9.hpc.uio.no"]' "$output"
