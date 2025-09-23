package filedb

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
)

type SampleDataNeeded int

const (
	DataNeedSamples SampleDataNeeded = iota
	DataNeedNodeSamples
	DataNeedCpuSamples
	DataNeedGpuSamples
)

type sampleFileReadSyncMethods struct {
	// The payload is a *SampleData always.
	dataNeeded SampleDataNeeded

	// The config is passed to the rectifiers, if they are not nil
	cfg *config.ClusterConfig
}

var _ = ReadSyncMethods((*sampleFileReadSyncMethods)(nil))

type SampleFileKind int

const (
	SampleFileKindSample SampleFileKind = iota
	SampleFileKindNodeSample
	SampleFileKindCpuSamples
	SampleFileKindGpuSamples
)

func NewSampleFileMethods(cfg *config.ClusterConfig, kind SampleFileKind) *sampleFileReadSyncMethods {
	switch kind {
	case SampleFileKindSample:
		return &sampleFileReadSyncMethods{DataNeedSamples, cfg}
	case SampleFileKindNodeSample:
		return &sampleFileReadSyncMethods{DataNeedNodeSamples, cfg}
	case SampleFileKindCpuSamples:
		return &sampleFileReadSyncMethods{DataNeedCpuSamples, cfg}
	case SampleFileKindGpuSamples:
		return &sampleFileReadSyncMethods{DataNeedGpuSamples, cfg}
	default:
		panic("Unexpected")
	}
}

type sampleData struct {
	samples     []*repr.Sample
	nodeSamples []*repr.NodeSample
	cpuSamples  []*repr.CpuSamples
	gpuSamples  []*repr.GpuSamples
}

type samplePayloadType = *sampleData

func (_ *sampleFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *sampleFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	switch sfr.dataNeeded {
	case DataNeedSamples:
		return payload.(samplePayloadType).samples
	case DataNeedNodeSamples:
		return payload.(samplePayloadType).nodeSamples
	case DataNeedCpuSamples:
		return payload.(samplePayloadType).cpuSamples
	case DataNeedGpuSamples:
		return payload.(samplePayloadType).gpuSamples
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
	var nodeSamples []*repr.NodeSample
	var cpuSamples []*repr.CpuSamples
	var gpuSamples []*repr.GpuSamples
	if (attr & FileSampleV0JSON) != 0 {
		samples, nodeSamples, cpuSamples, gpuSamples, softErrors, err =
			parse.ParseSamplesV0JSON(inputFile, uf, verbose)
	} else {
		samples, cpuSamples, gpuSamples, softErrors, err =
			parse.ParseSampleCSV(inputFile, uf, verbose)
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
	if CpuSamplesRectifier != nil {
		cpuSamples = CpuSamplesRectifier(cpuSamples, sfr.cfg)
	}
	if GpuSamplesRectifier != nil {
		gpuSamples = GpuSamplesRectifier(gpuSamples, sfr.cfg)
	}
	// TODO: Ideally, nodeSample rectifier?  We're not using the rectifiers for anything except
	// samples at the moment.
	payload = &sampleData{samples, nodeSamples, cpuSamples, gpuSamples}
	return
}

func (_ *sampleFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(samplePayloadType)
	size := unsafe.Sizeof(data)
	// Pointers to Samples
	size += uintptr(len(data.samples)) * repr.PointerSize
	// Every Sample has the same size
	size += uintptr(len(data.samples)) * repr.SizeofSample
	// Pointers to CpuSamples
	size += uintptr(len(data.cpuSamples)) * repr.PointerSize
	// Every CpuSamples object is unique
	for _, d := range data.cpuSamples {
		size += d.Size()
	}
	// Pointers to GpuSamples
	size += uintptr(len(data.gpuSamples)) * repr.PointerSize
	// Every GpuSamples object is unique
	for _, d := range data.gpuSamples {
		size += d.Size()
	}
	return size
}

func ReadSampleSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return ReadRecordsFromFiles[repr.Sample](files, verbose, reader)
}

func ReadNodeSampleSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (sampleBlobs [][]*repr.NodeSample, dropped int, err error) {
	return ReadRecordsFromFiles[repr.NodeSample](files, verbose, reader)
}

func ReadCpuSamplesSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (loadDataBlobs [][]*repr.CpuSamples, dropped int, err error) {
	return ReadRecordsFromFiles[repr.CpuSamples](files, verbose, reader)
}

func ReadGpuSamplesSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (gpuDataBlobs [][]*repr.GpuSamples, dropped int, err error) {
	return ReadRecordsFromFiles[repr.GpuSamples](files, verbose, reader)
}
