// HTTP authentication logic.

package httpsrv

import (
	"fmt"
	"net/http"

	"go-utils/auth"
	"go-utils/status"
)

// Given a (possibly nil) authenticator (user-to-password mapping) and a request, apply HTTP basic
// authentication for the given realm.  If the authentication fails then signal a 401 response and
// log the error.
//
// The realm name probably should not contain a `"` character.

func Authenticate(w http.ResponseWriter, r *http.Request, authenticator *auth.Authenticator, realm string) bool {
	user, pass, ok := r.BasicAuth()
	passed := !ok && authenticator == nil || ok && authenticator != nil && authenticator.Authenticate(user, pass)
	if !passed {
		if authenticator != nil {
			w.Header().Add("WWW-Authenticate", "Basic realm=\"" + realm + "\", charset=\"utf-8\"")
		}
		w.WriteHeader(401)
		fmt.Fprintf(w, "Unauthorized")
		status.Warning("Authorization failed")
		return false
	}
	return true
}

