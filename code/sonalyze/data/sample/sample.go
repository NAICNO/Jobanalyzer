package sample

import (
	"time"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

type SampleDataProvider struct {
	theLog db.ProcessSampleDataProvider
}

func OpenSampleDataProvider(meta types.Context) (*SampleDataProvider, error) {
	theLog, err := db.OpenReadOnlyDB(meta, types.SampleData)
	if err != nil {
		return nil, err
	}
	return &SampleDataProvider{theLog}, nil
}

func (sdp *SampleDataProvider) Query(
	fromDate, toDate time.Time,
	hostGlobber Multihost,
	recordFilter *SampleFilter,
	wantBounds bool,
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
	)
}

func (sdp *SampleDataProvider) QueryRaw(
	fromDate, toDate time.Time,
	hosts Multihost,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return sdp.theLog.ReadProcessSamples(
		types.DataProviderFilter{
			FromDate: fromDate,
			ToDate:   toDate,
			Node:     hosts,
		})
}

func (sdp *SampleDataProvider) Filenames(
	filter types.DataProviderFilter,
) ([]string, error) {
	if sampleDir, ok := sdp.theLog.(db.SampleFilenameProvider); ok {
		return sampleDir.SampleFilenames(filter)
	}
	panic("Bad cluster type")
}
