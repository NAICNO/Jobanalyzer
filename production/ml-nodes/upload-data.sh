#!/usr/bin/env bash

# Upload generated reports to a web server

# We need globbing, stay away from -f
set -eu -o pipefail

sonar_dir=$HOME/sonar
data_path=$sonar_dir/data
load_report_path=$data_path/load-reports

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $load_report_path/*.json

#scp -q -i ~/.ssh/axis_of_eval_naic $load_report_path/*.json naic@axis-of-eval.org:/var/www/naic/output
scp -q -i ~/.ssh/ubuntu-vm.pem $load_report_path/*.json ubuntu@158.39.48.160:/var/www/html/output
