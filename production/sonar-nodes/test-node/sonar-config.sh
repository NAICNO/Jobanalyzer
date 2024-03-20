# This is included by the other scripts, it is not meant to be run by itself.
#
# $sonar_dir is the directory containing the sonar binaries and the secrets/ subdir.

if [[ ! -v sonar_dir ]]; then
    echo "sonar_dir must be defined"
    exit 1
fi

# The canonical name of the cluster we're on.  naic-monitor.uio.no is a "test cluster" that we can
# use freely.
cluster=naic-monitor.uio.no

# The server receiving data.
upload_address=https://naic-monitor.uio.no

# This is a netrc-format file, see the curl documentation.  The identity must be known to the
# receiving server
curl_auth_file=$sonar_dir/secrets/upload-auth.netrc

# The test system upload window is set to 10 seconds for expediency.  Most other systems use "280"
# here.
upload_window=10

lockdir=$sonar_dir
