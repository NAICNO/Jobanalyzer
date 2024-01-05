// Utilities for HTTP service that seem to be needed in many servers.

package httpsrv

import (
	"fmt"
	"io"
	"net/http"

	"go-utils/auth"
	"go-utils/status"
)

// Assert that the method in the request is `method`.  If not, signal a 403 response and log it.

func AssertMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		w.WriteHeader(403)
		fmt.Fprintf(w, "Bad method")
		status.Warningf("Bad method: %s", r.Method)
		return false
	}
	return true
}

// Read the payload from a request and return it.  If the reading fails, signal a 400 and log it.

func ReadPayload(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	payload := make([]byte, r.ContentLength)
	haveRead := 0
	for haveRead < int(r.ContentLength) {
		n, err := r.Body.Read(payload[haveRead:])
		haveRead += n
		if err != nil {
			if err == io.EOF && haveRead == int(r.ContentLength) {
				break
			}
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad content")
			status.Warning("Bad content - can't read the body")
			return nil, false
		}
	}
	return payload, true
}

// Given a (possibly nil) authenticator (user-to-password mapping) and a request, apply HTTP basic
// authentication for the given realm.  If the authentication fails then signal a 401 response and
// log the error.
//
// The realm name probably should not contain a `"` character.

func Authenticate(
	w http.ResponseWriter,
	r *http.Request,
	authenticator *auth.Authenticator,
	realm string,
) bool {
	user, pass, ok := r.BasicAuth()
	passed := (!ok && authenticator == nil) ||
		(ok && authenticator != nil && authenticator.Authenticate(user, pass))
	if !passed {
		if authenticator != nil {
			w.Header().Add("WWW-Authenticate", "Basic realm=\""+realm+"\", charset=\"utf-8\"")
		}
		w.WriteHeader(401)
		fmt.Fprintf(w, "Unauthorized")
		status.Warning("Authorization failed")
		return false
	}
	return true
}
