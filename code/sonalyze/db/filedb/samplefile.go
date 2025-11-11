package filedb

import (
	"io"
	"unsafe"

	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
	"sonalyze/db/special"
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

	// TODO: Obsolete?  This was used with rectification
	meta special.ClusterMeta
}

var _ = ReadSyncMethods((*sampleFileReadSyncMethods)(nil))

type SampleFileKind int

const (
	SampleFileKindSample SampleFileKind = iota
	SampleFileKindNodeSample
	SampleFileKindCpuSamples
	SampleFileKindGpuSamples
)

func NewSampleFileMethods(meta special.ClusterMeta, kind SampleFileKind) *sampleFileReadSyncMethods {
	switch kind {
	case SampleFileKindSample:
		return &sampleFileReadSyncMethods{DataNeedSamples, meta}
	case SampleFileKindNodeSample:
		return &sampleFileReadSyncMethods{DataNeedNodeSamples, meta}
	case SampleFileKindCpuSamples:
		return &sampleFileReadSyncMethods{DataNeedCpuSamples, meta}
	case SampleFileKindGpuSamples:
		return &sampleFileReadSyncMethods{DataNeedGpuSamples, meta}
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

func (sfr *sampleFileReadSyncMethods) ReadDataLocked(
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

func readProcessSampleSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (sampleBlobs [][]*repr.Sample, dropped int, err error) {
	return readRecordsFromFiles[repr.Sample](files, verbose, reader)
}

func readNodeSampleSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (sampleBlobs [][]*repr.NodeSample, dropped int, err error) {
	return readRecordsFromFiles[repr.NodeSample](files, verbose, reader)
}

func readCpuSamplesSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (loadDataBlobs [][]*repr.CpuSamples, dropped int, err error) {
	return readRecordsFromFiles[repr.CpuSamples](files, verbose, reader)
}

func readGpuSamplesSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) (gpuDataBlobs [][]*repr.GpuSamples, dropped int, err error) {
	return readRecordsFromFiles[repr.GpuSamples](files, verbose, reader)
}
