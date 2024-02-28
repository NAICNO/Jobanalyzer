#!/bin/bash
#
# For this, you need the Firefox JS shell installed, see instructions in run_tests.js.

which js > /dev/null 2>&1
if [[ $? == 0 ]]; then
    output=$(js --version)
    if [[ $? == 0 && $output =~ JavaScript- ]]; then
        js run_tests.js
    else
        echo "JS shell not conformant, tests skipped"
    fi
else
    echo "No JS shell, tests skipped"
fi
