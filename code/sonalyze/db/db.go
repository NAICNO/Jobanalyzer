package db

import (
	"fmt"
	"sonalyze/db/special"
)

// Utility to open a database from a set of parameters.  We must have dataDir xor logFiles.
func OpenReadOnlyDB(
	meta special.ClusterMeta,
	dataType FileListDataType,
) (
	DataProvider,
	error,
) {
	var theLog DataProvider
	var err error
	if len(meta.LogFiles()) > 0 {
		theLog, err = OpenFileListDB(meta, dataType)
	} else {
		theLog, err = OpenPersistentDirectoryDB(meta)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to open log store: %v", err)
	}
	return theLog, nil
}
