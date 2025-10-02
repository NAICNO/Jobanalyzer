package filedb

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
)

type SysinfoDataNeeded int

const (
	DataNeedNodeData SysinfoDataNeeded = iota
	DataNeedCardData
)

type sysinfoData struct {
	nodeData []*repr.SysinfoNodeData
	cardData []*repr.SysinfoCardData
}

type sysinfoPayloadType = *sysinfoData

type sysinfoFileReadSyncMethods struct {
	dataNeeded SysinfoDataNeeded
}

var _ = ReadSyncMethods((*sysinfoFileReadSyncMethods)(nil))

type SysinfoFileKind int

const (
	SysinfoFileKindNodeData SysinfoFileKind = iota
	SysinfoFileKindCardData
)

func NewSysinfoFileMethods(_ *config.ClusterConfig, kind SysinfoFileKind) *sysinfoFileReadSyncMethods {
	switch kind {
	case SysinfoFileKindNodeData:
		return &sysinfoFileReadSyncMethods{DataNeedNodeData}
	case SysinfoFileKindCardData:
		return &sysinfoFileReadSyncMethods{DataNeedCardData}
	default:
		panic("Unexpected")
	}
}

func (_ *sysinfoFileReadSyncMethods) IsCacheable() bool {
	return false
}

func (sfr *sysinfoFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	switch sfr.dataNeeded {
	case DataNeedNodeData:
		return payload.(sysinfoPayloadType).nodeData
	case DataNeedCardData:
		return payload.(sysinfoPayloadType).cardData
	default:
		panic("Unexpected")
	}
}

func (sfr *sysinfoFileReadSyncMethods) ReadDataLockedAndRectify(
	attr FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var nodeData []*repr.SysinfoNodeData
	var cardData []*repr.SysinfoCardData
	if (attr & FileSysinfoV0JSON) != 0 {
		nodeData, cardData, _, err = parse.ParseSysinfoV0JSON(inputFile, verbose)
	} else {
		nodeData, cardData, _, err = parse.ParseSysinfoOldJSON(inputFile, verbose)
	}
	if err != nil {
		return
	}
	payload = &sysinfoData{nodeData, cardData}
	return
}

func (_ *sysinfoFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(sysinfoPayloadType)
	size := unsafe.Sizeof(data)
	size += uintptr(len(data.nodeData)) * repr.PointerSize
	for _, d := range data.nodeData {
		size += d.Size()
	}
	size += uintptr(len(data.cardData)) * repr.PointerSize
	for _, d := range data.cardData {
		size += d.Size()
	}
	return size
}

func ReadSysinfoNodeDataSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*repr.SysinfoNodeData, int, error) {
	return readRecordsFromFiles[repr.SysinfoNodeData](files, verbose, reader)
}

func ReadSysinfoCardDataSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*repr.SysinfoCardData, int, error) {
	return readRecordsFromFiles[repr.SysinfoCardData](files, verbose, reader)
}
