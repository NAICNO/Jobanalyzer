// Interface to a database based on file lists.  See doc.go in this directory and in filedb/
// for more information.

package db

import (
	"errors"

	"go-utils/config"
	"sonalyze/db/filedb"
)

// The types can be OR'ed together if they are provided by the same file type.  This is sort of a
// secret handshake for now.
type FileListDataType int

const (
	FileListSampleData FileListDataType = 1 << iota
	FileListCpuSampleData
	FileListGpuSampleData
	FileListNodeData
	FileListCardData
	FileListSlurmJobData
	FileListSlurmNodeData
	FileListSlurmPartitionData
	FileListSlurmClusterData
)

type FileListDataProvider struct {
	*filedb.TransientSampleCluster
	*filedb.TransientSysinfoCluster
	*filedb.TransientSacctCluster
	*filedb.TransientCluzterCluster
	cfg      *config.ClusterConfig
	dataType FileListDataType
}

func (tdp *FileListDataProvider) Config() *config.ClusterConfig {
	return tdp.cfg
}

func (tdb *FileListDataProvider) DataType() FileListDataType {
	return tdb.dataType
}

func OpenFileListDB(
	dataType FileListDataType,
	files []string,
	cfg *config.ClusterConfig,
) (DataProvider, error) {
	var transientSampleCluster *filedb.TransientSampleCluster
	var transientSysinfoCluster *filedb.TransientSysinfoCluster
	var transientSacctCluster *filedb.TransientSacctCluster
	var transientCluzterCluster *filedb.TransientCluzterCluster
	var err error
	switch {
	case dataType&(FileListSampleData|FileListCpuSampleData|FileListGpuSampleData) != 0:
		if dataType&^(FileListSampleData|FileListCpuSampleData|FileListGpuSampleData) != 0 {
			panic("Incompatible type flags")
		}
		transientSampleCluster, err = OpenFileListSampleDB(files, cfg)
	case dataType&(FileListNodeData|FileListCardData) != 0:
		if dataType&^(FileListNodeData|FileListCardData) != 0 {
			panic("Incompatible type flags")
		}
		transientSysinfoCluster, err = OpenFileListSysinfoDB(files, cfg)
	case dataType&FileListSlurmJobData != 0:
		if dataType&^FileListSlurmJobData != 0 {
			panic("Incompatible type flags")
		}
		transientSacctCluster, err = OpenFileListSacctDB(files, cfg)
	case dataType&(FileListSlurmNodeData|FileListSlurmPartitionData|FileListSlurmClusterData) != 0:
		if dataType&^(FileListSlurmNodeData|FileListSlurmPartitionData|FileListSlurmClusterData) != 0 {
			panic("Incompatible type flags")
		}
		transientCluzterCluster, err = OpenFileListCluzterDB(files, cfg)
	default:
		panic("NYI")
	}
	if err != nil {
		return nil, err
	}
	return &FileListDataProvider{
		transientSampleCluster,
		transientSysinfoCluster,
		transientSacctCluster,
		transientCluzterCluster,
		cfg,
		dataType,
	}, nil
}

// The following are internal but are public for testing

func OpenFileListSampleDB(
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

func OpenFileListSacctDB(
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

func OpenFileListSysinfoDB(
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

func OpenFileListCluzterDB(
	files []string,
	cfg *config.ClusterConfig,
) (*filedb.TransientCluzterCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return filedb.NewTransientCluzterCluster(files, filedb.FileCluzterV0JSON, cfg), nil
}
