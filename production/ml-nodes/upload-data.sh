#!/usr/bin/env bash

# Upload generated reports to a web server.
#
# NOTE!  If upload logic needs to change here, also consider webload-5m-and-upload.sh, which
# performs its own upload.

# We need globbing, stay away from -f
set -eu -o pipefail

sonar_dir=$HOME/sonar
data_path=$sonar_dir/data
load_report_path=$data_path/load-reports

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $load_report_path/*.json

source $sonar_dir/upload-config.sh

#scp -C -q -i $IDENTITY_FILE_NAME $load_report_path/*.json $WWWUSER_AND_HOST:$WWWUSER_UPLOAD_PATH

# StrictHostKeyChecking has to be disabled here because this is not an interactive script,
# and the VM has not been configured to respond in such a way that the value in known_hosts
# will bypass the interactive prompt.
scp -C -q -o StrictHostKeyChecking=no -i $IDENTITY_FILE_NAME \
    $load_report_path/*.json \
    $WWWUSER_AND_HOST:$WWWUSER_UPLOAD_PATH


