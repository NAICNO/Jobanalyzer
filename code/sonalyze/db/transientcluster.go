// A TransientSampleCluster maintains a static list of file names from which Sonar `ps` sample
// records are read.  The files are read-only and not cacheable.  Mostly this functionality is used
// for testing.

package db

import (
	"path"
	"sync"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
)

type TransientSampleCluster struct /* implements SampleCluster */ {
	sync.Mutex
	closed bool
	cfg    *config.ClusterConfig
	files  []*LogFile
}

func newTransientSampleCluster(
	fileNames []string,
	cfg *config.ClusterConfig,
) *TransientSampleCluster {
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
				fileSonarSamples,
			),
		)
	}
	return &TransientSampleCluster{
		cfg:   cfg,
		files: files,
	}
}

func (fc *TransientSampleCluster) Config() *config.ClusterConfig {
	fc.Lock()
	defer fc.Unlock()
	if fc.closed {
		return nil
	}

	return fc.cfg
}

func (fc *TransientSampleCluster) Close() error {
	fc.Lock()
	defer fc.Unlock()
	if fc.closed {
		return ClusterClosedErr
	}

	fc.closed = true
	return nil
}

func (fc *TransientSampleCluster) SampleFilenames(
	_, _ time.Time,
	_ *hostglob.HostGlobber,
) ([]string, error) {
	fc.Lock()
	defer fc.Unlock()
	if fc.closed {
		return nil, ClusterClosedErr
	}

	return filenames(fc.files), nil
}

func (fc *TransientSampleCluster) ReadSamples(
	_, _ time.Time,
	_ *hostglob.HostGlobber,
	verbose bool,
) (samples []*Sample, dropped int, err error) {
	fc.Lock()
	defer fc.Unlock()
	if fc.closed {
		return nil, 0, ClusterClosedErr
	}

	return readSamples(fc.files, verbose, fc.cfg)
}
