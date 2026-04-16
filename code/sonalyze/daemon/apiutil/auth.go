// Authorization stuff

package apiutil

import (
	"encoding/base64"
	"strings"
)

type AuthHeader struct {
	Auth string `header:"Authorization"`
}

func DecodeAuth(auth string) (string, string) {
	if strings.HasPrefix(auth, "Basic ") {
		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(auth[6:]))
		if err == nil {
			user, pass, found := strings.Cut(string(data), ":")
			if found {
				return strings.TrimSpace(user), strings.TrimSpace(pass)
			}
		}
	}
	return "-NO-USER-", "-NO-PASS-"
}
