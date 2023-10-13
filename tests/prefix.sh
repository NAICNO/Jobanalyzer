# See comments in run_tests.sh in this directory for all information.

# These variables are read by the suffix.sh program.  They are private to the test runner but
# visible also to the test cases, hence the tr_ prefix.
tr_hard_errors=0
tr_soft_errors=0

# CHECK test_name expected_value observed_value [ known_bug_number ]

CHECK () {
    local name="$TEST_NAME: $1"
    local expected="$2"
    local output="$3"
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

# CHECK_ERR test_name observed_exit_code observed_output error_pattern [ known_bug_number ]

CHECK_ERR () {
    local name="$TEST_NAME: $1"
    local observed_code="$2"
    local observed_output="$3"
    local error_pattern="$4"
    local knownbug="$5"
    local failed=0
    local matches=0
    if [[ $observed_output =~ $error_pattern ]]; then
        matches=1
    fi
    if (( observed_code == 0 || matches == 0 )); then
        failed=1
    fi
    if (( failed == 1 )); then
	if [[ $knownbug != "" ]]; then
	    echo "FAILED " $name ": KNOWN BUG $knownbug"
	    echo " EXPECTED ERROR EXIT WITH MATCHING MESSAGE"
            if (( observed_code == 0 )); then
	        echo " GOT      SUCCESS EXIT"
            else
                echo " GOT      PATTERN MATCH FAILURE"
            fi
	    tr_soft_errors=$((tr_soft_errors + 1))
	else
	    echo "FAILED " $name
	    echo " EXPECTED ERROR EXIT WITH MATCHING MESSAGE"
            if (( observed_code == 0 )); then
	        echo " GOT      SUCCESS EXIT"
            else
                echo " GOT      PATTERN MATCH FAILURE"
            fi
	    echo " OUTPUT   $observed_output"
            echo " PATTERN  $error_pattern"
	    tr_hard_errors=$((tr_hard_errors + 1))
	fi
    elif [[ $knownbug != "" ]]; then
	echo "UNEXPECTED SUCCESS " $name ": KNOWN BUG $knownbug"
	echo " EXPECTED SUCCESS EXIT"
        if (( observed_code == 0 )); then
	    echo " GOT      PATTERN MATCH SUCCESS"
        else
	    echo " GOT      ERROR EXIT"
        fi
	tr_hard_errors=$((tr_hard_errors + 1))
    else
	echo "$name: OK"
    fi
}
