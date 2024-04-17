#!/bin/bash
#
# Test that `sonalyze help` and `sonalyze -h` produce help text on stdout and nothing on stderr.

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}

set -e

if [[ $($GO_SONALYZE help 2>&1 > /dev/null | grep Usage | wc -l) -ne 1 ]]; then
    echo "Bogus help - no usage"
    exit 1
fi

if [[ $($GO_SONALYZE help 2> /dev/null | wc -l) -ne 0 ]]; then
    echo "Bogus help - stdout"
    exit 1
fi

if [[ $($GO_SONALYZE -h 2>&1 > /dev/null | grep Usage | wc -l) -ne 1 ]]; then
    echo "Bogus -h - no usage"
    exit 1
fi

if [[ $($GO_SONALYZE -h 2> /dev/null | wc -l) -ne 0 ]]; then
    echo "Bogus -h - stdout"
    exit 1
fi
