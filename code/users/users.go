// This is experimental code to manage a table of users in an SQL database for sonalyze.
//
// Usage: see the `help` function below or run `users help`.
//
// Tentative schema for the relevant tables:
//
// There is a USER table
//   username  varchar index
//   name      varchar free-text index?
//   password  varchar
//
// There is a CAPABILITY table
//   username  varchar index
//   cluster   varchar index
//   local-id  varchar
//
// Authentication: The username/password authenticates a user with a unique username.
//
// Authorization: On a given cluster, the user is authorized to see the data for the named local-id
// on the cluster if the local-id is other than "-", or for all users if the local-id is "-".
//
// The ability to see the data for eg a group of users is not considered a requirement at this time.
//
// There's an assumption that local-ids on a cluster are never reused.  If they could be reused,
// then we could add eg a date range here, but the likelihood of the admins getting this right is
// slim.  Instead, when a new CAPABILITY is added for a cluster and user name, we need to check that
// no other user on that cluster has that local-id.

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	// postgres
	"github.com/jackc/pgx/v5"
)

// postgres - for experimentation only
// This is appropriate for `trust` authentication - an experiment
const url = "postgres://larstha@localhost:5432/sonalyze?sslmode=disable"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		help(1)
	}
	command := args[0]
	args = args[1:]
	if command == "help" || command == "-h" || command == "--help" {
		help(0)
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	switch {
	case command == "add" && len(args) == 2:
		cmdAdd(conn, args[0], args[1])
	case command == "passwd" && len(args) == 1:
		cmdPasswd(conn, args[0])
	case command == "delete" && len(args) == 1:
		cmdDelete(conn, args[0])
	case command == "lookup" && len(args) == 1:
		cmdLookup(conn, args[0])
	case command == "search" && len(args) == 1:
		cmdSearch(conn, args[0])
	case command == "authorize" && len(args) == 3:
		cmdAuthorize(conn, args[0], args[1], args[2])
	case command == "revoke" && len(args) == 2:
		cmdRevoke(conn, args[0], args[1])
	case command == "revoke" && len(args) == 1:
		cmdRevokeAll(conn, args[0])
	default:
		help(1)
	}
}

func help(exit int) {
	fmt.Fprintf(os.Stderr,
		`Usage:
  users add username real-name
    Add user to the database, username must not exist.  Will query for the user's password.

  users passwd username
    Change user's password, username must exist.  Will query for the user's password.

  users delete username
    Remove the user from the database along with all capabilities, username must exist.
    Will prompt for confirmation.

  users lookup username
    Lookup user in the database and print their information.

  users search real-name-substring
    Search database for match of real name and print matching real names and user names.
    Match is a regular expression, go wild.

  users authorize username cluster local-id
    Add or change authorization for username on cluster to be local-id.  The user must
    exist.  Will prompt for confirmation.

  users revoke username cluster
    Revoke any authorization for username on cluster.  The user must exist but need not
    have any authorization on the cluster.  Will prompt for confirmation.

  users revoke username
    Revoke any authorization for username on all clusters.  The user must exist.  Will
    prompt for confirmation.

All commands will exit(1) with a sensible error if preconditions are not met (usually,
user exists or does not exist, as the case may be).
`)
	os.Exit(exit)
}

func cmdAdd(conn *pgx.Conn, username, realname string) {
	passwd := promptNewPasswd(username)
	// For proper error handling the username must be unique and this must be enforced by the DB.
	_, err := conn.Exec(
		context.Background(),
		"insert into user(username, realname, passwd) values ($1, $2, $3)",
		username, realname, passwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add user failed: %v\n", err)
		os.Exit(1)
	}
}

func cmdPasswd(conn *pgx.Conn, username string) {
	passwd := promptNewPasswd(username)
	// Transaction
	//   Find the user
	//   update the record
	_ = passwd
}

func cmdDelete(conn *pgx.Conn, username string) {
	if !confirm(
		fmt.Sprintf("Do you want to delete the user %s and all their capabilities",
			username),
	) {
		return
	}
	// Transaction
	//   Find the username record
	//   Delete all capabilities for the user
	//   Delete the user
}

func cmdLookup(conn *pgx.Conn, username string) {
	// Probably a tx so that capabilities for user don't change under our feet?

	// First lookup the real name
	var realname string
	err := conn.QueryRow(
		context.Background(),
		"select realname from user where username=$1", username,
	).Scan(&realname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find user %s: %v\n", username, err)
		os.Exit(1)
	}

	// Then lookup the capabilities for the users
	// ...

	fmt.Printf("%s: %s\n", username, realname)
}

func cmdSearch(conn *pgx.Conn, username string) {
}

func cmdAuthorize(conn *pgx.Conn, username, clustername, localid string) {
}

func cmdRevoke(conn *pgx.Conn, username, clustername string) {
}

func cmdRevokeAll(conn *pgx.Conn, username string) {
}

var stdinReader = bufio.NewScanner(os.Stdin)

func promptNewPasswd(username string) string {
	for {
		fmt.Printf("Enter new password for user `%s`\n", username)
		if !stdinReader.Scan() {
			os.Exit(1)
		}
		pass1 := stdinReader.Text()
		fmt.Println("Re-enter password:")
		if !stdinReader.Scan() {
			os.Exit(1)
		}
		pass2 := stdinReader.Text()
		if pass1 == pass2 {
			return pass1
		}
		fmt.Println("Passwords are not equal")
	}
}

func confirm(msg string) bool {
	for {
		fmt.Println(msg + "(yes/NO)")
		if !stdinReader.Scan() {
			os.Exit(1)
		}
		s := strings.TrimSpace(stdinReader.Text())
		if s == "" {
			return false
		}
		ans := strings.ToUpper(s)
		if ans == "YES" {
			return true
		}
		if ans == "NO" {
			return false
		}
		fmt.Println(`Please answer "yes" or "no"`)
	}
}
