#!/usr/bin/bash

# Upload generated reports to a web server.

# We need globbing, stay away from -f
set -eu -o pipefail

cluster=mlx.hpc.uio.no
sonar_dir=${sonar_dir:-$HOME/sonar}
script_dir=$sonar_dir/scripts/$cluster
report_dir=$sonar_dir/reports/$cluster

upload_files="$report_dir/*.json"
source $script_dir/upload-subr.sh


