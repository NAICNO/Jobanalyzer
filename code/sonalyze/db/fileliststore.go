// Interface to a database based on file lists.  See doc.go in this directory and in filedb/
// for more information.

package db

import (
	"errors"

	"sonalyze/db/filedb"
	"sonalyze/db/types"
)

type FileListDataProvider struct {
	*filedb.TransientSampleCluster
	*filedb.TransientSysinfoCluster
	*filedb.TransientSacctCluster
	*filedb.TransientCluzterCluster
	meta     types.Context
	dataType types.DataType
}

func (tdb *FileListDataProvider) DataType() types.DataType {
	return tdb.dataType
}

func OpenFileListDB(
	meta types.Context,
	dataType types.DataType,
) (DataProvider, error) {
	var transientSampleCluster *filedb.TransientSampleCluster
	var transientSysinfoCluster *filedb.TransientSysinfoCluster
	var transientSacctCluster *filedb.TransientSacctCluster
	var transientCluzterCluster *filedb.TransientCluzterCluster
	var err error
	switch {
	case dataType&types.ProcessSampleData != 0:
		if dataType&^types.ProcessSampleData != 0 {
			panic("Incompatible type flags")
		}
		transientSampleCluster, err = OpenFileListSampleDB(meta, meta.LogFiles(types.ProcessSampleData))
	case dataType&types.SysinfoData != 0:
		if dataType&^types.SysinfoData != 0 {
			panic("Incompatible type flags")
		}
		transientSysinfoCluster, err = OpenFileListSysinfoDB(meta, meta.LogFiles(types.SysinfoData))
	case dataType&types.SlurmJobData != 0:
		if dataType&^types.SlurmJobData != 0 {
			panic("Incompatible type flags")
		}
		transientSacctCluster, err = OpenFileListSacctDB(meta, meta.LogFiles(types.SlurmJobData))
	case dataType&types.SlurmSystemData != 0:
		if dataType&^types.SlurmSystemData != 0 {
			panic("Incompatible type flags")
		}
		transientCluzterCluster, err = OpenFileListCluzterDB(meta, meta.LogFiles(types.SlurmSystemData))
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
		meta,
		dataType,
	}, nil
}

// The following are internal but are public for testing

func OpenFileListSampleDB(
	meta types.Context,
	files []string,
) (*filedb.TransientSampleCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := filedb.SniffTypeFromFilenames(files, filedb.FileSampleCSV, filedb.FileSampleV0JSON)
	if err != nil {
		return nil, err
	}
	return filedb.NewTransientSampleCluster(meta, ty, files), nil
}

func OpenFileListSacctDB(
	meta types.Context,
	files []string,
) (*filedb.TransientSacctCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := filedb.SniffTypeFromFilenames(files, filedb.FileSlurmCSV, filedb.FileSlurmV0JSON)
	if err != nil {
		return nil, err
	}
	return filedb.NewTransientSacctCluster(meta, ty, files), nil
}

func OpenFileListSysinfoDB(
	meta types.Context,
	files []string,
) (*filedb.TransientSysinfoCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	ty, err := filedb.SniffTypeFromFilenames(files, filedb.FileSysinfoOldJSON, filedb.FileSysinfoV0JSON)
	if err != nil {
		return nil, err
	}
	return filedb.NewTransientSysinfoCluster(meta, ty, files), nil
}

func OpenFileListCluzterDB(
	meta types.Context,
	files []string,
) (*filedb.TransientCluzterCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return filedb.NewTransientCluzterCluster(meta, filedb.FileCluzterV0JSON, files), nil
}
