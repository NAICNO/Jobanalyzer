// A TransientXCluster maintains a static list of file names holding data of type X.  The files are
// read-only and not cacheable.  Mostly this functionality is used for testing.

package filedb

import (
	"path"
	"time"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/repr"
)

// TransientCluster is a transient cluster mixin that has only one type of files.
type TransientCluster struct {
	// MT: Immutable after initialization
	cfg   *config.ClusterConfig
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

func (tc *TransientCluster) Config() *config.ClusterConfig {
	return tc.cfg
}

func (tc *TransientCluster) Filenames() ([]string, error) {
	return Filenames(tc.files), nil
}

type TransientSampleCluster struct /* implements SampleCluster */ {
	// MT: Immutable after initialization
	samplesMethods     ReadSyncMethods
	nodeSamplesMethods ReadSyncMethods
	loadDataMethods    ReadSyncMethods
	gpuDataMethods     ReadSyncMethods

	TransientCluster
}

func NewTransientSampleCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientSampleCluster {
	return &TransientSampleCluster{
		samplesMethods:     NewSampleFileMethods(cfg, SampleFileKindSample),
		nodeSamplesMethods: NewSampleFileMethods(cfg, SampleFileKindNodeSample),
		loadDataMethods:    NewSampleFileMethods(cfg, SampleFileKindCpuSamples),
		gpuDataMethods:     NewSampleFileMethods(cfg, SampleFileKindGpuSamples),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientSampleCluster) SampleFilenames(
	_, _ time.Time,
	_ *Hosts,
) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSampleCluster) ReadSamples(
	_, _ time.Time,
	_ *Hosts,
	verbose bool,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return ReadSampleSlice(tsc.files, verbose, tsc.samplesMethods)
}

func (tsc *TransientSampleCluster) ReadNodeSamples(
	_, _ time.Time,
	_ *Hosts,
	verbose bool,
) (sampleBlobs [][]*repr.NodeSample, dropped int, err error) {
	return ReadNodeSampleSlice(tsc.files, verbose, tsc.nodeSamplesMethods)
}

func (tsc *TransientSampleCluster) ReadCpuSamples(
	_, _ time.Time,
	_ *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.CpuSamples, dropped int, err error) {
	return ReadCpuSamplesSlice(tsc.files, verbose, tsc.loadDataMethods)
}

func (tsc *TransientSampleCluster) ReadGpuSamples(
	_, _ time.Time,
	_ *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.GpuSamples, dropped int, err error) {
	return ReadGpuSamplesSlice(tsc.files, verbose, tsc.gpuDataMethods)
}

type TransientSacctCluster struct /* implements SacctCluster */ {
	// MT: Immutable after initialization
	methods ReadSyncMethods

	TransientCluster
}

func NewTransientSacctCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientSacctCluster {
	return &TransientSacctCluster{
		methods: NewSacctFileMethods(cfg),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientSacctCluster) SacctFilenames(_, _ time.Time) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSacctCluster) ReadSacctData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.SacctInfo, dropped int, err error) {
	return ReadSacctSlice(tsc.files, verbose, tsc.methods)
}

type TransientSysinfoCluster struct /* implements SysinfoCluster */ {
	// MT: Immutable after initialization
	nodeDataMethods ReadSyncMethods
	cardDataMethods ReadSyncMethods

	TransientCluster
}

func NewTransientSysinfoCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientSysinfoCluster {
	return &TransientSysinfoCluster{
		nodeDataMethods: NewSysinfoFileMethods(cfg, SysinfoFileKindNodeData),
		cardDataMethods: NewSysinfoFileMethods(cfg, SysinfoFileKindCardData),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientSysinfoCluster) SysinfoFilenames(
	_,
	_ time.Time,
	_ *Hosts,
) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSysinfoCluster) ReadSysinfoNodeData(
	fromDate, toDate time.Time,
	_ *Hosts,
	verbose bool,
) (recordBlobs [][]*repr.SysinfoNodeData, dropped int, err error) {
	return ReadSysinfoNodeDataSlice(tsc.files, verbose, tsc.nodeDataMethods)
}

func (tsc *TransientSysinfoCluster) ReadSysinfoCardData(
	fromDate, toDate time.Time,
	_ *Hosts,
	verbose bool,
) (recordBlobs [][]*repr.SysinfoCardData, dropped int, err error) {
	return ReadSysinfoCardDataSlice(tsc.files, verbose, tsc.cardDataMethods)
}

type TransientCluzterCluster struct /* implements CluzterCluster */ {
	// MT: Immutable after initialization
	attributeMethods ReadSyncMethods
	partitionMethods ReadSyncMethods
	nodeMethods      ReadSyncMethods

	TransientCluster
}

func NewTransientCluzterCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientCluzterCluster {
	return &TransientCluzterCluster{
		attributeMethods: NewCluzterFileMethods(cfg, CluzterFileKindAttributeData),
		partitionMethods: NewCluzterFileMethods(cfg, CluzterFileKindPartitionData),
		nodeMethods:      NewCluzterFileMethods(cfg, CluzterFileKindNodeData),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientCluzterCluster) CluzterFilenames(_, _ time.Time) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientCluzterCluster) ReadCluzterAttributeData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterAttributes, dropped int, err error) {
	return ReadCluzterAttributeDataSlice(tsc.files, verbose, tsc.attributeMethods)
}

func (tsc *TransientCluzterCluster) ReadCluzterPartitionData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterPartitions, dropped int, err error) {
	return ReadCluzterPartitionDataSlice(tsc.files, verbose, tsc.partitionMethods)
}

func (tsc *TransientCluzterCluster) ReadCluzterNodeData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterNodes, dropped int, err error) {
	return ReadCluzterNodeDataSlice(tsc.files, verbose, tsc.nodeMethods)
}
