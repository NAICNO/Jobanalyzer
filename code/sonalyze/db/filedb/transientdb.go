// A TransientXCluster maintains a static list of file names holding data of type X.  The files are
// read-only and not cacheable.  Mostly this functionality is used for testing.

package filedb

import (
	"path"
	"time"

	. "sonalyze/common"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

// TransientCluster is a transient cluster mixin that has only one type of files.
type TransientCluster struct {
	// MT: Immutable after initialization
	meta  types.Context
	files []*LogFile
}

func processTransientFiles(fileNames []string, ty FileAttr) []*LogFile {
	if len(fileNames) == 0 {
		panic("Empty list of files")
	}
	files := make([]*LogFile, 0, len(fileNames))
	for _, fn := range fileNames {
		files = append(files,
			NewLogFile(
				Fullname{
					Cluster:  "",
					Dirname:  path.Dir(fn),
					Basename: path.Base(fn),
				},
				ty,
			),
		)
	}
	return files
}

func (tc *TransientCluster) Filenames() ([]string, error) {
	return Filenames(tc.files), nil
}

type TransientSampleCluster struct /* implements SampleCluster */ {
	// MT: Immutable after initialization
	samplesMethods     ReadSyncMethods
	nodeSamplesMethods ReadSyncMethods
	diskSamplesMethods ReadSyncMethods
	loadDataMethods    ReadSyncMethods
	gpuDataMethods     ReadSyncMethods

	TransientCluster
}

func NewTransientSampleCluster(
	meta types.Context,
	ty FileAttr,
	fileNames []string,
) *TransientSampleCluster {
	return &TransientSampleCluster{
		samplesMethods:     NewSampleFileMethods(SampleFileKindSample),
		nodeSamplesMethods: NewSampleFileMethods(SampleFileKindNodeSample),
		diskSamplesMethods: NewSampleFileMethods(SampleFileKindDiskSample),
		loadDataMethods:    NewSampleFileMethods(SampleFileKindCpuSamples),
		gpuDataMethods:     NewSampleFileMethods(SampleFileKindGpuSamples),
		TransientCluster: TransientCluster{
			meta:  meta,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientSampleCluster) SampleFilenames(
	_ types.DataProviderFilter,
) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSampleCluster) ReadProcessSamples(
	_ types.DataProviderFilter,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return readProcessSampleSlice(tsc.files, tsc.samplesMethods)
}

func (tsc *TransientSampleCluster) ReadNodeSamples(
	_ types.DataProviderFilter,
) (sampleBlobs [][]*repr.NodeSample, dropped int, err error) {
	return readNodeSampleSlice(tsc.files, tsc.nodeSamplesMethods)
}

func (tsc *TransientSampleCluster) ReadDiskSamples(
	_ types.DataProviderFilter,
) (sampleBlobs [][]*repr.DiskSample, dropped int, err error) {
	return readDiskSampleSlice(tsc.files, tsc.diskSamplesMethods)
}

func (tsc *TransientSampleCluster) ReadCpuSamples(
	_ types.DataProviderFilter,
) (dataBlobs [][]*repr.CpuSamples, dropped int, err error) {
	return readCpuSamplesSlice(tsc.files, tsc.loadDataMethods)
}

func (tsc *TransientSampleCluster) ReadGpuSamples(
	_ types.DataProviderFilter,
) (dataBlobs [][]*repr.GpuSamples, dropped int, err error) {
	return readGpuSamplesSlice(tsc.files, tsc.gpuDataMethods)
}

type TransientSacctCluster struct /* implements SacctCluster */ {
	// MT: Immutable after initialization
	methods ReadSyncMethods

	TransientCluster
}

func NewTransientSacctCluster(
	meta types.Context,
	ty FileAttr,
	fileNames []string,
) *TransientSacctCluster {
	return &TransientSacctCluster{
		methods: NewSacctFileMethods(),
		TransientCluster: TransientCluster{
			meta:  meta,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientSacctCluster) SacctFilenames(_, _ time.Time) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSacctCluster) ReadSacctData(
	_ types.DataProviderFilter,
) (recordBlobs [][]*repr.SacctInfo, dropped int, err error) {
	return ReadSacctSlice(tsc.files, tsc.methods)
}

type TransientSysinfoCluster struct /* implements SysinfoCluster */ {
	// MT: Immutable after initialization
	nodeDataMethods ReadSyncMethods
	cardDataMethods ReadSyncMethods

	TransientCluster
}

func NewTransientSysinfoCluster(
	meta types.Context,
	ty FileAttr,
	fileNames []string,
) *TransientSysinfoCluster {
	return &TransientSysinfoCluster{
		nodeDataMethods: NewSysinfoFileMethods(SysinfoFileKindNodeData),
		cardDataMethods: NewSysinfoFileMethods(SysinfoFileKindCardData),
		TransientCluster: TransientCluster{
			meta:  meta,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientSysinfoCluster) SysinfoFilenames(
	_,
	_ time.Time,
	_ Hosts,
) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSysinfoCluster) ReadSysinfoNodeData(
	_ types.DataProviderFilter,
) (recordBlobs [][]*repr.SysinfoNodeData, dropped int, err error) {
	return ReadSysinfoNodeDataSlice(tsc.files, tsc.nodeDataMethods)
}

func (tsc *TransientSysinfoCluster) ReadSysinfoCardData(
	_ types.DataProviderFilter,
) (recordBlobs [][]*repr.SysinfoCardData, dropped int, err error) {
	return ReadSysinfoCardDataSlice(tsc.files, tsc.cardDataMethods)
}

type TransientCluzterCluster struct /* implements CluzterCluster */ {
	// MT: Immutable after initialization
	attributeMethods ReadSyncMethods
	partitionMethods ReadSyncMethods
	nodeMethods      ReadSyncMethods

	TransientCluster
}

func NewTransientCluzterCluster(
	meta types.Context,
	ty FileAttr,
	fileNames []string,
) *TransientCluzterCluster {
	return &TransientCluzterCluster{
		attributeMethods: NewCluzterFileMethods(CluzterFileKindAttributeData),
		partitionMethods: NewCluzterFileMethods(CluzterFileKindPartitionData),
		nodeMethods:      NewCluzterFileMethods(CluzterFileKindNodeData),
		TransientCluster: TransientCluster{
			meta:  meta,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientCluzterCluster) CluzterFilenames(_, _ time.Time) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientCluzterCluster) ReadCluzterAttributeData(
	_ types.DataProviderFilter,
) (recordBlobs [][]*repr.CluzterAttributes, dropped int, err error) {
	return ReadCluzterAttributeDataSlice(tsc.files, tsc.attributeMethods)
}

func (tsc *TransientCluzterCluster) ReadCluzterPartitionData(
	_ types.DataProviderFilter,
) (recordBlobs [][]*repr.CluzterPartitions, dropped int, err error) {
	return ReadCluzterPartitionDataSlice(tsc.files, tsc.partitionMethods)
}

func (tsc *TransientCluzterCluster) ReadCluzterNodeData(
	_ types.DataProviderFilter,
) (recordBlobs [][]*repr.CluzterNodes, dropped int, err error) {
	return ReadCluzterNodeDataSlice(tsc.files, tsc.nodeMethods)
}
