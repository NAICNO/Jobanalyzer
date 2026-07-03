package table

import (
	"strings"
)

// We know we're dealing with ASCII so this is good enough
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}
