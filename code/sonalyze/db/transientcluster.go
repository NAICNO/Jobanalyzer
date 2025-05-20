// A TransientSampleCluster maintains a static list of file names from which Sonar `ps` sample
// records are read.  The files are read-only and not cacheable.  Mostly this functionality is used
// for testing.
//
// A TransientSacctCluster does the same, but for Slurm `sacct` data.

package db

import (
	"path"
	"sync"
	"time"

	"go-utils/config"
	. "sonalyze/common"
)

// This is a transient cluster mixin that has only one type of files.

type TransientCluster struct {
	// MT: Immutable after initialization
	cfg   *config.ClusterConfig
	files []*LogFile

	sync.Mutex
	closed bool
}

func processTransientFiles(fileNames []string, ty FileAttr) []*LogFile {
	if len(fileNames) == 0 {
		panic("Empty list of files")
	}
	files := make([]*LogFile, 0, len(fileNames))
	for _, fn := range fileNames {
		files = append(files,
			newLogFile(
				Fullname{
					cluster:  "",
					dirname:  path.Dir(fn),
					basename: path.Base(fn),
				},
				ty,
			),
		)
	}
	return files
}

func (tc *TransientCluster) Config() *config.ClusterConfig {
	tc.Lock()
	defer tc.Unlock()
	if tc.closed {
		return nil
	}

	return tc.cfg
}

func (tc *TransientCluster) Close() error {
	tc.Lock()
	defer tc.Unlock()
	if tc.closed {
		return ClusterClosedErr
	}

	tc.closed = true
	return nil
}

func (tc *TransientCluster) Filenames() ([]string, error) {
	tc.Lock()
	defer tc.Unlock()
	if tc.closed {
		return nil, ClusterClosedErr
	}

	return filenames(tc.files), nil
}

type TransientSampleCluster struct /* implements SampleCluster */ {
	// MT: Immutable after initialization
	samplesMethods  ReadSyncMethods
	loadDataMethods ReadSyncMethods
	gpuDataMethods  ReadSyncMethods

	TransientCluster
}

func newTransientSampleCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientSampleCluster {
	return &TransientSampleCluster{
		samplesMethods:  newSampleFileMethods(cfg, sampleFileKindSample),
		loadDataMethods: newSampleFileMethods(cfg, sampleFileKindLoadDatum),
		gpuDataMethods:  newSampleFileMethods(cfg, sampleFileKindGpuDatum),
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
) (sampleBlobs [][]*Sample, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readSampleSlice(tsc.files, verbose, tsc.samplesMethods)
}

func (tsc *TransientSampleCluster) ReadLoadData(
	_, _ time.Time,
	_ *Hosts,
	verbose bool,
) (dataBlobs [][]*LoadDatum, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readLoadDatumSlice(tsc.files, verbose, tsc.loadDataMethods)
}

func (tsc *TransientSampleCluster) ReadGpuData(
	_, _ time.Time,
	_ *Hosts,
	verbose bool,
) (dataBlobs [][]*GpuDatum, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readGpuDatumSlice(tsc.files, verbose, tsc.gpuDataMethods)
}

type TransientSacctCluster struct /* implements SacctCluster */ {
	// MT: Immutable after initialization
	methods ReadSyncMethods

	TransientCluster
}

func newTransientSacctCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientSacctCluster {
	return &TransientSacctCluster{
		methods: newSacctFileMethods(cfg),
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
) (recordBlobs [][]*SacctInfo, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readSacctSlice(tsc.files, verbose, tsc.methods)
}

type TransientSysinfoCluster struct /* implements SysinfoCluster */ {
	// MT: Immutable after initialization
	methods ReadSyncMethods

	TransientCluster
}

var _ SysinfoCluster = (*TransientSysinfoCluster)(nil)

func newTransientSysinfoCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientSysinfoCluster {
	return &TransientSysinfoCluster{
		methods: newSysinfoFileMethods(cfg),
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

func (tsc *TransientSysinfoCluster) ReadSysinfoData(
	fromDate, toDate time.Time,
	_ *Hosts,
	verbose bool,
) (recordBlobs [][]*config.NodeConfigRecord, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readNodeConfigRecordSlice(tsc.files, verbose, tsc.methods)
}

type TransientCluzterCluster struct /* implements CluzterCluster */ {
	// MT: Immutable after initialization
	methods ReadSyncMethods

	TransientCluster
}

func newTransientCluzterCluster(
	fileNames []string,
	ty FileAttr,
	cfg *config.ClusterConfig,
) *TransientCluzterCluster {
	return &TransientCluzterCluster{
		methods: newCluzterFileMethods(cfg),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames, ty),
		},
	}
}

func (tsc *TransientCluzterCluster) CluzterFilenames(_, _ time.Time) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientCluzterCluster) ReadCluzterData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*CluzterInfo, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readCluzterSlice(tsc.files, verbose, tsc.methods)
}
