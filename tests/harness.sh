if [[ $t_output != $t_expected ]]; then
    echo "TEST " $t_name
    echo "EXPECTED " $t_expected
    echo "GOT " $t_output
    exit 1
else
    echo "$t_name: OK"
fi
