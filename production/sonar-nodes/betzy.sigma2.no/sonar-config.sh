# This is included by the other scripts, it is not meant to be run by itself.
#
# $sonar_dir is the directory containing the sonar binaries and the secrets/ subdir.

if [[ ! -v sonar_dir ]]; then
    echo "sonar_dir must be defined"
    exit 1
fi

# The canonical name of the cluster we're on.
cluster=betzy.sigma2.no

# The server receiving data.
upload_address=https://naic-monitor.uio.no

# This is a netrc-format file, see the curl documentation.  The identity must be known to the
# receiving server.
curl_auth_file=$sonar_dir/secrets/upload-auth.netrc

# The upload window is set to 280 seconds so that the upload is almost certain to be done before
# sonar runs the next time, assuming a 5-minute interval for sonar runs.  Correctness does not
# depend on that - data can arrive at the server in any order, so long as they are tagged with the
# correct timestamp - but it's nice to not have more processes running concurrently than necessary.
upload_window=${upload_window:-280}

# On most systems this is just "curl" but sometimes there needs to be a custom build.  The version
# must be new enough to handle modern web protocols (TLS v1.3 probably) and --retry-connrefused.
curl_binary=/cluster/shared/sonar/bin/curl
