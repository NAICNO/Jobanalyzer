package auth

import (
	"fmt"
	"os"
	"strings"
)

// The file must provide a user name and password on the form username/password (old form) or
// username:password (new form, compatible with curl), to be used in an HTTP basic authentication
// header.

func ParseAuth(filename string) (string, string, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return "", "", err
	}
	// The format of the file is a single pair a/b (old style) or a:b (new style), with leading and
	// trailing whitespace ignored.  In the old style, ":" was never legal in the username.
	xs := strings.Split(strings.TrimSpace(string(bs)), "/")
	if len(xs) != 2 {
		xs = strings.Split(strings.TrimSpace(string(bs)), ":")
	}
	if len(xs) != 2 || strings.Index(xs[0], ":") != -1 {
		return "", "", fmt.Errorf("Authentication file has the wrong format or illegal values")
	}
	return xs[0], xs[1], nil
}

// Read a file with lines of username and password pairs and return a function that will check a
// username/password pair.  Only the new form, "username:password", is accepted.  Lines can be blank
// (mostly a concession to an empty last line, sort of silly).

func ParsePasswdFile(filename string) (func(user, pass string) bool, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	lines := strings.Split(string(bs), "\n")
	for i, l := range lines {
		s := strings.TrimSpace(l)
		if s == "" {
			continue
		}
		xs := strings.Split(s, ":")
		if len(xs) != 2 {
			return nil, fmt.Errorf("Password file has the wrong format (line %d)", i+1)
		}
		if _, found := m[xs[0]]; found {
			return nil, fmt.Errorf("Password file has duplicated user name (line %d)", i+1)
		}
		m[xs[0]] = xs[1]
	}
	return func(user, pass string) bool {
		probe, found := m[user]
		return found && probe == pass
	}, nil
}
