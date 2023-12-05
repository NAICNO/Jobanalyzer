package auth

import (
	"errors"
	"os"
	"strings"
)

// The file named must provide a user name and password on the form username/password, to be used in
// an HTTP basic authentication header.

func ParseAuth(filename string) (string, string, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return "", "", err
	}
	// The format of the file is a single pair a/b, with leading and trailing whitespace ignored
	xs := strings.Split(strings.TrimSpace(string(bs)), "/")
	if len(xs) != 2 || strings.Index(xs[0], ":") != -1 {
		return "", "", errors.New("Authentication file has the wrong format or illegal values")
	}
	return xs[0], xs[1], nil
}


