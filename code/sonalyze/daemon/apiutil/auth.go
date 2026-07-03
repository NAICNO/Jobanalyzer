// Authorization stuff

package apiutil

import (
	"encoding/base64"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"go-utils/auth"
)

type AuthHeader struct {
	Auth string `header:"Authorization"`
}

func DecodeAuth(authHdr string) (string, string) {
	if strings.HasPrefix(authHdr, "Basic ") {
		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(authHdr[6:]))
		if err == nil {
			user, pass, found := strings.Cut(string(data), ":")
			if found {
				return strings.TrimSpace(user), strings.TrimSpace(pass)
			}
		}
	}
	return "-NO-USER-", "-NO-PASS-"
}

func CheckAuth(command string, authenticator *auth.Authenticator, authHdr string) huma.StatusError {
	if authenticator != nil {
		user, pass := DecodeAuth(authHdr)
		if !authenticator.Authenticate(user, pass) {
			return huma.Error401Unauthorized(command + ": Unknown user/pass combination")
		}
	}
	return nil
}
