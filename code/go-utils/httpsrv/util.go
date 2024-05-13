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

func AssertMethod(w http.ResponseWriter, r *http.Request, method string, verbose bool) bool {
	if r.Method != method {
		w.WriteHeader(403)
		fmt.Fprintf(w, "Bad method")
		if verbose {
			status.Warningf("Bad method: %s", r.Method)
		}
		return false
	}
	return true
}

// Read the payload from a request and return it.  If the reading fails, signal a 400 and log it.

func ReadPayload(w http.ResponseWriter, r *http.Request, verbose bool) ([]byte, bool) {
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
			if verbose {
				status.Warning("Bad content - can't read the body")
			}
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
//
// The realm name can be empty, in which case no further information is requested, the request will
// just fail.  This is sensible when the client is supposed to know the realm for which to
// authenticate.

func Authenticate(
	w http.ResponseWriter,
	r *http.Request,
	authenticator *auth.Authenticator,
	realm string,
	verbose bool,
) (bool, string) {
	user, pass, ok := r.BasicAuth()
	passed := (!ok && authenticator == nil) ||
		(ok && authenticator != nil && authenticator.Authenticate(user, pass))
	if !passed {
		if authenticator != nil && realm != "" {
			w.Header().Add("WWW-Authenticate", "Basic realm=\""+realm+"\", charset=\"utf-8\"")
		}
		w.WriteHeader(401)
		fmt.Fprintf(w, "Unauthorized")
		if verbose {
			status.Warning("Authorization failed")
		}
		return false, ""
	}
	return true, user
}

func AssertContentType(
	w http.ResponseWriter,
	r *http.Request,
	requestedContentType string,
	verbose bool,
) bool {
	ct, ok := r.Header["Content-Type"]
	if !ok || ct[0] != requestedContentType {
		receivedContentType := "(no type)"
		if ok {
			receivedContentType = ct[0]
		}
		w.WriteHeader(400)
		fmt.Fprintf(w, "Bad content-type")
		if verbose {
			status.Warningf(
				"Bad content-type got %s wanted %s",
				receivedContentType,
				requestedContentType,
			)
		}
		return false
	}
	return true
}
