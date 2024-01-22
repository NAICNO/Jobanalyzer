package httpsrv

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"go-utils/auth"
	"go-utils/status"
)

// This is old code, currently unused but rescued for future use.  It serves a file from a request.

func NewFileHandler(
	verbose bool,
	fileRoot, authRealm string,
	authenticator *auth.Authenticator,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if verbose {
			status.Infof("Request from %s: %v", r.RemoteAddr, r.Header)
			status.Infof("%v", r.URL)
		}

		if !AssertMethod(w, r, "GET") ||
			!Authenticate(w, r, authenticator, authRealm) {
			return
		}
		_, havePayload := ReadPayload(w, r)
		if !havePayload {
			return
		}

		p := path.Clean(r.URL.Path)
		if strings.Index(p, "..") != -1 {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Bad URL")
			return
		}
		filepath := path.Join(fileRoot, p)
		f, err := os.Open(filepath)
		if err != nil {
			w.WriteHeader(404)
			fmt.Fprintf(w, "No such file .../%s", p)
			return
		}
		defer f.Close()
		all, err := io.ReadAll(f)
		if err != nil {
			w.WriteHeader(404)
			fmt.Fprintf(w, "Unable to read file .../%s", p)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write(all)
		// The write may have gone badly but there you have it.  Let the client mop it up.
	}
}

