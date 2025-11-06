package sample

import (
	"time"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

type SampleDataProvider struct {
	theLog db.ProcessSampleDataProvider
}

func OpenSampleDataProvider(meta special.ClusterMeta) (*SampleDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, special.SampleData)
	if err != nil {
		return nil, err
	}
	return &SampleDataProvider{theLog}, nil
}

func (sdp *SampleDataProvider) Query(
	fromDate, toDate time.Time,
	hostGlobber *Hosts,
	recordFilter *SampleFilter,
	wantBounds bool,
	verbose bool,
) (
	streams InputStreamSet,
	bounds Timebounds,
	read, dropped int,
	err error,
) {
	return ReadSampleStreamsAndMaybeBounds(
		sdp.theLog,
		fromDate,
		toDate,
		hostGlobber,
		recordFilter,
		wantBounds,
		verbose,
	)
}

func (sdp *SampleDataProvider) QueryRaw(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return sdp.theLog.ReadProcessSamples(fromDate, toDate, hosts, verbose)
}

func (sdp *SampleDataProvider) Filenames(
	fromDate, toDate time.Time,
	hostGlobber *Hosts,
) ([]string, error) {
	if sampleDir, ok := sdp.theLog.(db.SampleFilenameProvider); ok {
		return sampleDir.SampleFilenames(fromDate, toDate, hostGlobber)
	}
	panic("Bad cluster type")
}
