// Utilities for HTTP service that seem to be needed in many servers.

package httpsrv

import (
	"fmt"
	"io"
	"net/http"

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
