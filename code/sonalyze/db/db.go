package db

import (
	"fmt"
	"sonalyze/db/special"
)

// Utility to open a database from a set of parameters.  We must have dataDir xor logFiles.
func OpenReadOnlyDB(
	configFile string,
	dataDir string,
	dataType FileListDataType,
	logFiles []string,
) (
	DataProvider,
	error,
) {
	cfg, err := special.MaybeGetConfig(configFile)
	if err != nil {
		return nil, err
	}
	var theLog DataProvider
	if len(logFiles) > 0 {
		// This does not work, probably a default value gets in the way?
		// if dataDir != "" {
		// 	return nil, fmt.Errorf("Can't have both dataDir and logFiles")
		// }
		theLog, err = OpenFileListDB(dataType, logFiles, cfg)
	} else {
		if dataDir == "" {
			return nil, fmt.Errorf("Must have either dataDir or logFiles")
		}
		theLog, err = OpenPersistentDirectoryDB(dataDir, cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to open log store: %v", err)
	}
	return theLog, nil
}
