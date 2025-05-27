package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
)

type DataNeeded int

const (
	DataNeedSamples DataNeeded = iota
	DataNeedLoadData
	DataNeedGpuData
)

type sampleFileReadSyncMethods struct {
	// The payload is a *SampleData always.
	dataNeeded DataNeeded

	// The config is passed to the rectifiers, if they are not nil
	cfg *config.ClusterConfig
}

var _ = ReadSyncMethods((*sampleFileReadSyncMethods)(nil))

type sampleFileKind int

const (
	sampleFileKindSample sampleFileKind = iota
	sampleFileKindLoadDatum
	sampleFileKindGpuDatum
)

func newSampleFileMethods(cfg *config.ClusterConfig, kind sampleFileKind) *sampleFileReadSyncMethods {
	switch kind {
	case sampleFileKindSample:
		return &sampleFileReadSyncMethods{DataNeedSamples, cfg}
	case sampleFileKindLoadDatum:
		return &sampleFileReadSyncMethods{DataNeedLoadData, cfg}
	case sampleFileKindGpuDatum:
		return &sampleFileReadSyncMethods{DataNeedGpuData, cfg}
	default:
		panic("Unexpected")
	}
}

type sampleData struct {
	samples  []*repr.Sample
	loadData []*repr.LoadDatum
	gpuData  []*repr.GpuDatum
}

type samplePayloadType = *sampleData

func (_ *sampleFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *sampleFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	switch sfr.dataNeeded {
	case DataNeedSamples:
		return payload.(samplePayloadType).samples
	case DataNeedLoadData:
		return payload.(samplePayloadType).loadData
	case DataNeedGpuData:
		return payload.(samplePayloadType).gpuData
	default:
		panic("Unexpected")
	}
}

func (sfr *sampleFileReadSyncMethods) ReadDataLockedAndRectify(
	attr FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var samples []*repr.Sample
	var loadData []*repr.LoadDatum
	var gpuData []*repr.GpuDatum
	if (attr & FileSampleV0JSON) != 0 {
		samples, loadData, gpuData, softErrors, err = parse.ParseSamplesV0JSON(inputFile, uf, verbose)
	} else {
		samples, loadData, gpuData, softErrors, err = parse.ParseSampleCSV(inputFile, uf, verbose)
	}
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
	if GpuDatumRectifier != nil {
		gpuData = GpuDatumRectifier(gpuData, sfr.cfg)
	}
	payload = &sampleData{samples, loadData, gpuData}
	return
}

func (_ *sampleFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(samplePayloadType)
	size := unsafe.Sizeof(data)
	// Pointers to Samples
	size += uintptr(len(data.samples)) * repr.PointerSize
	// Every Sample has the same size
	size += uintptr(len(data.samples)) * repr.SizeofSample
	// Pointers to loadData
	size += uintptr(len(data.loadData)) * repr.PointerSize
	// Every LoadDatum is unique
	for _, d := range data.loadData {
		size += d.Size()
	}
	// Pointers to GpuDatums
	size += uintptr(len(data.gpuData)) * repr.PointerSize
	// Every GpuDatum is unique
	for _, d := range data.gpuData {
		size += d.Size()
	}
	return size
}

func readSampleSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return readRecordsFromFiles[repr.Sample](files, verbose, reader)
}

func readLoadDatumSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (loadDataBlobs [][]*repr.LoadDatum, dropped int, err error) {
	return readRecordsFromFiles[repr.LoadDatum](files, verbose, reader)
}

func readGpuDatumSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (gpuDataBlobs [][]*repr.GpuDatum, dropped int, err error) {
	return readRecordsFromFiles[repr.GpuDatum](files, verbose, reader)
}
