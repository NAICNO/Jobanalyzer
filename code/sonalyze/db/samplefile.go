package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

type sampleFileReadSyncMethods struct {
	// The payload is a *SampleData always.  `wantSamples` is true if the client wants the `samples`
	// array, false if it wants the `loadData` array.
	wantSamples bool

	// The config is passed to the rectifiers, if they are not nil
	cfg *config.ClusterConfig
}

type sampleFileKind int

const (
	sampleFileKindSample sampleFileKind = iota
	sampleFileKindLoadDatum
)

func newSampleFileMethods(cfg *config.ClusterConfig, kind sampleFileKind) *sampleFileReadSyncMethods {
	switch kind {
	case sampleFileKindSample:
		return &sampleFileReadSyncMethods{true, cfg}
	case sampleFileKindLoadDatum:
		return &sampleFileReadSyncMethods{false, cfg}
	default:
		panic("Unexpected")
	}
}

type sampleData struct {
	samples  []*Sample
	loadData []*LoadDatum
}

func (_ *sampleFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *sampleFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	if sfr.wantSamples {
		return payload.(*sampleData).samples
	}
	return payload.(*sampleData).loadData
}

func (sfr *sampleFileReadSyncMethods) ReadDataLockedAndRectify(
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var samples []*Sample
	var loadData []*LoadDatum
	samples, loadData, softErrors, err = ParseSonarLog(inputFile, uf, verbose)
	if err != nil {
		return
	}
	// There is a tricky invariant here that we're not going to check: If data are
	// cached, then we require not just that there is a ClusterConfig for the
	// cluster, but config info for each node. This is so that we can rectify the
	// input data based on system info.  If there is no config for the cluster or
	// the code then these data may remain wonky.
	//
	// The reason we don't check the invariant is that the effects of not having a
	// config are fairly benign, and also that so much else depends on having a
	// config that we'll get a more thorough check in other ways.
	if SampleRectifier != nil {
		samples = SampleRectifier(samples, sfr.cfg)
	}
	if LoadDatumRectifier != nil {
		loadData = LoadDatumRectifier(loadData, sfr.cfg)
	}
	payload = &sampleData{samples, loadData}
	return
}

var (
	// MT: Constant after initialization; immutable
	perSampleSize int64
)

func init() {
	var s Sample
	perSampleSize = int64(unsafe.Sizeof(s) + unsafe.Sizeof(&s))
}

func (_ *sampleFileReadSyncMethods) CachedSizeOfPayload(payload any) int64 {
	data := payload.(*sampleData)
	return perSampleSize*int64(len(data.samples)) + 8*int64(len(data.loadData))
}

func readSampleSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (sampleBlobs [][]*Sample, dropped int, err error) {
	return readRecordsFromFiles[Sample](files, verbose, reader)
}

func readLoadDatumSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (loadDataBlobs [][]*LoadDatum, dropped int, err error) {
	return readRecordsFromFiles[LoadDatum](files, verbose, reader)
}
