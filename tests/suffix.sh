# See comments in run_tests.sh in this directory.

if ((HARD_ERRORS + SOFT_ERRORS != 0)); then
    echo $HARD_ERRORS " HARD ERRORS"
    echo $SOFT_ERRORS " SOFT ERRORS"
    exit 1
fi
