# See comments in run_tests.sh in this directory for all information.

# These variables are read by the suffix.sh program.  They are private to the test runner but
# visible also to the test cases, hence the tr_ prefix.
tr_hard_errors=0
tr_soft_errors=0

# CHECK test_name expected_value observed_value [ known_bug_number ]

CHECK () {
    local name="$TEST_NAME: $1"
    local expected="$2"
    local got="$3"
    local knownbug="$4"
    if [[ "$output" != "$expected" ]]; then
	if [[ $knownbug != "" ]]; then
	    echo "FAILED " $name ": KNOWN BUG $knownbug"
	    echo " EXPECTED $expected"
	    echo " GOT      $output"
	    tr_soft_errors=$((tr_soft_errors + 1))
	else
	    echo "FAILED " $name
	    echo "EXPECTED $expected"
	    echo "GOT      $output"
	    tr_hard_errors=$((tr_hard_errors + 1))
	fi
    elif [[ $knownbug != "" ]]; then
	echo "UNEXPECTED SUCCESS " $name ": KNOWN BUG $knownbug"
	echo " EXPECTED $expected"
	echo " GOT      $output"
	tr_hard_errors=$((tr_hard_errors + 1))
    else
	echo "$name: OK"
    fi
}
