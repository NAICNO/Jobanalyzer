#!/bin/bash

set -e
if [[ -f test-$(hostname).sh ]]; then
    ./test-$(hostname).sh
fi
