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
	"go-utils/hostglob"
)

// This is a transient cluster mixin that has only one type of files.

type TransientCluster struct {
	// MT: Immutable after initialization
	cfg   *config.ClusterConfig
	files []*LogFile

	sync.Mutex
	closed bool
}

func processTransientFiles(fileNames []string) []*LogFile {
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
				0,
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

	TransientCluster
}

func newTransientSampleCluster(
	fileNames []string,
	cfg *config.ClusterConfig,
) *TransientSampleCluster {
	return &TransientSampleCluster{
		samplesMethods:  newSampleFileMethods(cfg, sampleFileKindSample),
		loadDataMethods: newSampleFileMethods(cfg, sampleFileKindLoadDatum),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames),
		},
	}
}

func (tsc *TransientSampleCluster) SampleFilenames(
	_, _ time.Time,
	_ *hostglob.HostGlobber,
) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSampleCluster) ReadSamples(
	_, _ time.Time,
	_ *hostglob.HostGlobber,
	verbose bool,
) (samples []*Sample, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readSampleSlice(tsc.files, verbose, tsc.samplesMethods)
}

func (tsc *TransientSampleCluster) ReadLoadData(
	_, _ time.Time,
	_ *hostglob.HostGlobber,
	verbose bool,
) (data []*LoadDatum, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readLoadDatumSlice(tsc.files, verbose, tsc.loadDataMethods)
}

type TransientSacctCluster struct /* implements SacctCluster */ {
	// MT: Immutable after initialization
	methods ReadSyncMethods

	TransientCluster
}

func newTransientSacctCluster(
	fileNames []string,
	cfg *config.ClusterConfig,
) *TransientSacctCluster {
	return &TransientSacctCluster{
		methods: newSacctFileMethods(cfg),
		TransientCluster: TransientCluster{
			cfg:   cfg,
			files: processTransientFiles(fileNames),
		},
	}
}

func (tsc *TransientSacctCluster) SacctFilenames(_, _ time.Time) ([]string, error) {
	return tsc.Filenames()
}

func (tsc *TransientSacctCluster) ReadSacctData(
	fromDate, toDate time.Time,
	verbose bool,
) (records []*SacctInfo, dropped int, err error) {
	tsc.Lock()
	defer tsc.Unlock()
	if tsc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readSacctSlice(tsc.files, verbose, tsc.methods)
}
