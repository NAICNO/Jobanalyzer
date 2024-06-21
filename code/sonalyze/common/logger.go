package common

import (
	"go-utils/status"
)

// MT: Constant after initialization; thread-safe
var Log status.Logger = status.Default()

func init() {
	if DEBUG {
		Log.SetLevel(status.LogLevelInfo)
	}
}
