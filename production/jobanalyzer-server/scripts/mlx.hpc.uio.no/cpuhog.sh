#!/usr/bin/bash
#
# Run sonalyze for the `cpuhog` use case and capture its output in a file appropriate for the
# current time and system.

set -euf -o pipefail

cluster=mlx.hpc.uio.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

output_dir=${state_dir}/$(date +'%Y/%m/%d')
mkdir -p ${output_dir}

# Report jobs that have used "a lot" of CPU and have run for at least 10 minutes but have not touched the
# GPU.  Reports go to stdout.
#
# What's "a lot" of CPU?  We define this for now as a peak of at least 10 cores.  This is imperfect
# but at least not completely wrong.
#
# The report is run on the data for the last 24h.  This should therefore be run about once every 12h
# at least, but ==> IMPORTANTLY, it MUST be run often enough that job IDs are not reused between
# consecutive runs.  It is not expensive, and can be run fairly often.

$naicreport_dir/sonalyze jobs \
		    $data_source \
		    -u - \
		    --no-gpu --min-rcpu-peak=10 --min-runtime=10m \
		    --fmt=csvnamed,tag:cpuhog,now,std,cpu-peak,gpu-peak,rcpu,rmem,start,end,cmd \
		    "$@" \
		    >> ${output_dir}/cpuhog.csv

