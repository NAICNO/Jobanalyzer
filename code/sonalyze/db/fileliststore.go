// Interface to a database based on file lists.  See doc.go in this directory and in filedb/
// for more information.

package db

import (
	"errors"

	"go-utils/config"
	"sonalyze/db/filedb"
)

func OpenTransientSampleCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*filedb.TransientSampleCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := filedb.SniffTypeFromFilenames(files, filedb.FileSampleCSV, filedb.FileSampleV0JSON)
	if err != nil {
		return nil, err
	}
	return filedb.NewTransientSampleCluster(files, ty, cfg), nil
}

func OpenTransientSacctCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*filedb.TransientSacctCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := filedb.SniffTypeFromFilenames(files, filedb.FileSlurmCSV, filedb.FileSlurmV0JSON)
	if err != nil {
		return nil, err
	}
	return filedb.NewTransientSacctCluster(files, ty, cfg), nil
}

func OpenTransientSysinfoCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*filedb.TransientSysinfoCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := filedb.SniffTypeFromFilenames(files, filedb.FileSysinfoOldJSON, filedb.FileSysinfoV0JSON)
	if err != nil {
		return nil, err
	}
	return filedb.NewTransientSysinfoCluster(files, ty, cfg), nil
}

func OpenTransientCluzterCluster(
	files []string,
	cfg *config.ClusterConfig,
) (*filedb.TransientCluzterCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return filedb.NewTransientCluzterCluster(files, filedb.FileCluzterV0JSON, cfg), nil
}
