if [[ $($SONALYZE version) =~ untagged_sonar_data ]]; then
    # Parse the old old untagged format (eight fields)

    output=$($SONALYZE jobs --user michaelm --fmt=csv,end,job,mem-peak -- old-old-data.csv)
    CHECK old_old_untagged_data "2023-10-10 22:40,9625735,1" "$output"

fi
