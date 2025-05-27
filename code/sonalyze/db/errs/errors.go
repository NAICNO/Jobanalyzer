package errs

import (
	"errors"
)

var (
	// MT: Constant after initialization; immutable
	BadTimestampErr  = errors.New("Bad timestamp")
	ClusterClosedErr = errors.New("ClusterStore is closed")
	ReadOnlyDirErr   = errors.New("Cluster is read-only list of files")
)
