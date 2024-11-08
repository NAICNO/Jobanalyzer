#!/usr/bin/bash
#
# monmon - monitor whether the naic-monitor server is up
#
# This will probably fall into disuse after we get systemd integration with automatic restarting,
# and/or we get a bigger server.

MAILTO=${MAILTO:-larstha@uio.no}

SONALYZE=~/sonar/sonalyze
REMOTE=https://naic-monitor.uio.no
AUTH=~/.ssh/monmon-auth.netrc

output=$("$SONALYZE" cluster -remote "$REMOTE" -auth-file "$AUTH")
exitcode=$?
if [[ $exitcode != 0 ]]; then
    if [[ ! -f monmon-cookie ]]; then
        touch monmon-cookie
        mail -s "Sonalyzed is down" "$MAILTO" <<EOF

Sonalyzed appears to be down and may need to be restarted!
Don't forget to delete monmon-cookie.

Exit code ${exitcode}
Command output:

${output}

EOF
    fi
fi
