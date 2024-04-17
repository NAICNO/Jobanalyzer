# The --user - is needed since --command applies in addition to the user filter.
output=$($SONALYZE jobs --user - --command python --min-samples=1 -f 2023-10-04 --fmt=csv,job,cmd -- select_command.csv)
CHECK one_command_1 "1269178,python" "$output"

# Ditto
output=$($SONALYZE jobs --user - --command python3 --min-samples=1 -f 2023-10-04 --fmt=csv,job,cmd -- select_command.csv)
CHECK one_command_2 "1269179,python3" "$output"

# `--exclude-command` implies that we include all the others
output=$($SONALYZE jobs --user - --exclude-command python2 --min-samples=1 -f 2023-10-04 --fmt=csv,job,cmd -- select_command.csv | sort)
CHECK two_commands "1269178,python
1269179,python3" "$output"

# User and matched command - should find one
output=$($SONALYZE jobs --user nissen --command python --min-samples=1 -f 2023-10-04 --fmt=csv,job,cmd -- select_command.csv)
CHECK one_user_and_command "1269178,python" "$output"

# Users and matched commands - should find two
output=$($SONALYZE jobs --user nissen -u zappa --command python --command python3 --min-samples=1 -f 2023-10-04 --fmt=csv,job,user,cmd -- select_command.csv)
CHECK two_users_and_commands "1269178,nissen,python
1269179,zappa,python3" "$output"

# User and mismatched command - should find none
output=$($SONALYZE jobs --user satan --command python --min-samples=1 -f 2023-10-04 --fmt=csv,job,cmd -- select_command.csv)
CHECK no_commands "" "$output"
