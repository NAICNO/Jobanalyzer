// Interface to a database based on file lists.  See doc.go in this directory and in filedb/
// for more information.

package db

import (
	"errors"

	"sonalyze/db/filedb"
	"sonalyze/db/special"
)

type FileListDataProvider struct {
	*filedb.TransientSampleCluster
	*filedb.TransientSysinfoCluster
	*filedb.TransientSacctCluster
	*filedb.TransientCluzterCluster
	meta     special.ClusterMeta
	dataType special.DataType
}

func (tdb *FileListDataProvider) DataType() special.DataType {
	return tdb.dataType
}

func OpenFileListDB(
	meta special.ClusterMeta,
	dataType special.DataType,
) (DataProvider, error) {
	var transientSampleCluster *filedb.TransientSampleCluster
	var transientSysinfoCluster *filedb.TransientSysinfoCluster
	var transientSacctCluster *filedb.TransientSacctCluster
	var transientCluzterCluster *filedb.TransientCluzterCluster
	var err error
	switch {
	case dataType&special.ProcessSampleData != 0:
		if dataType&^special.ProcessSampleData != 0 {
			panic("Incompatible type flags")
		}
		transientSampleCluster, err = OpenFileListSampleDB(meta, meta.LogFiles(special.ProcessSampleData))
	case dataType&special.SysinfoData != 0:
		if dataType&^special.SysinfoData != 0 {
			panic("Incompatible type flags")
		}
		transientSysinfoCluster, err = OpenFileListSysinfoDB(meta, meta.LogFiles(special.SysinfoData))
	case dataType&special.SlurmJobData != 0:
		if dataType&^special.SlurmJobData != 0 {
			panic("Incompatible type flags")
		}
		transientSacctCluster, err = OpenFileListSacctDB(meta, meta.LogFiles(special.SlurmJobData))
	case dataType&special.SlurmSystemData != 0:
		if dataType&^special.SlurmSystemData != 0 {
			panic("Incompatible type flags")
		}
		transientCluzterCluster, err = OpenFileListCluzterDB(meta, meta.LogFiles(special.SlurmSystemData))
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
	meta special.ClusterMeta,
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
	meta special.ClusterMeta,
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
	meta special.ClusterMeta,
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
	meta special.ClusterMeta,
	files []string,
) (*filedb.TransientCluzterCluster, error) {
	if len(files) == 0 {
		return nil, errors.New("Empty list of files")
	}
	return filedb.NewTransientCluzterCluster(meta, filedb.FileCluzterV0JSON, files), nil
}
