package db

import (
	"fmt"
	"sonalyze/db/special"
)

// Utility to open a database from a set of parameters.  We must have dataDir xor logFiles.
func OpenReadOnlyDB(
	meta special.ClusterMeta,
	dataDir string,
	dataType FileListDataType,
	logFiles []string,
) (
	DataProvider,
	error,
) {
	var theLog DataProvider
	var err error
	if len(logFiles) > 0 {
		// This does not work, probably a default value gets in the way?
		// if dataDir != "" {
		// 	return nil, fmt.Errorf("Can't have both dataDir and logFiles")
		// }
		theLog, err = OpenFileListDB(meta, dataType, logFiles)
	} else {
		if dataDir == "" {
			return nil, fmt.Errorf("Must have either dataDir or logFiles")
		}
		theLog, err = OpenPersistentDirectoryDB(meta, dataDir)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to open log store: %v", err)
	}
	meta.SetDataProvider(theLog)
	return theLog, nil
}
