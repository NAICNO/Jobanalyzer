# See comments in run_tests.sh in this directory.
#
# The test runner will define the free variable TEST_NAME.  The name that is passed as a parameter
# to this function is always interpreted in the context of that name.

# Is read by the suffix.sh program.

export HARD_ERRORS=0
export SOFT_ERRORS=0

# CHECK test_name expected_value observed_value [ known_bug_number ]

CHECK () {
    local name="$TEST_NAME: $1"
    local expected=$2
    local got=$3
    local knownbug=$4
    if [[ $output != $expected ]]; then
	if [[ $knownbug != "" ]]; then
	    echo "FAILED " $name ": KNOWN BUG $knownbug"
	    echo " EXPECTED " $expected
	    echo " GOT      " $output
	    SOFT_ERRORS=$((SOFT_ERRORS + 1))
	else
	    echo "FAILED " $name
	    echo "EXPECTED " $expected
	    echo "GOT " $output
	    HARD_ERRORS=$((HARD_ERRORS + 1))
	fi
    elif [[ $knownbug != "" ]]; then
	echo "UNEXPECTED SUCCESS " $name ": KNOWN BUG $knownbug"
	echo " EXPECTED " $expected
	echo " GOT      " $output
	HARD_ERRORS=$((HARD_ERRORS + 1))
    else
	echo "$name: OK"
    fi
}
