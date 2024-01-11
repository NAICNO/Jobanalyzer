# See comments in run_tests.sh in this directory.

if ((tr_hard_errors + tr_soft_errors != 0)); then
    echo "$tr_hard_errors HARD ERRORS"
    echo "$tr_soft_errors SOFT ERRORS"
    if ((tr_hard_errors > 0)); then
	exit 1
    else
	exit 2
    fi
fi
