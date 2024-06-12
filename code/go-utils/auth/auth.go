// Authorization checking abstraction.
//
// A password file has a sequence of lines, each with a username:password syntax (blanks are
// significant, but empty lines are ignored).  This can be read with ReadPasswords() to produce an
// Authenticator object that can be used to authenticate credentials.
//
// An authorization file has a single line with the same syntax.  This can be read with ParseAuth()
// to produce a username/password pair that can be passed to the authenticator, or the file name can
// be passed as an argument to curl -u.
//
// The authenticator can be reinitialized after creation (reading from the same file, which is
// presumed to have changed).  Reinitialization is thread-safe, and if it fails to read the file the
// authenticator is unchanged.

package auth

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go-utils/filesys"
)

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

// MT: Locked
type Authenticator struct {
	lock       sync.RWMutex
	filepath   string
	identities map[string]string
}

func ReadPasswords(filename string) (*Authenticator, error) {
	mapping, err := readPasswords(filename)
	if err != nil {
		return nil, err
	}
	return &Authenticator{
		filepath:   filename,
		identities: mapping,
	}, nil
}

func readPasswords(filename string) (map[string]string, error) {
	lines, err := filesys.FileLines(filename)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
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
	return m, nil
}

func (a *Authenticator) Authenticate(user, pass string) bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	probe, found := a.identities[user]
	return found && probe == pass
}

func (a *Authenticator) Reread() error {
	m, err := readPasswords(a.filepath)
	if err != nil {
		return err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.identities = m
	return nil
}
