package filedb

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
)

type sysinfoPayloadType = []*repr.SysinfoData

type sysinfoFileReadSyncMethods struct {
}

var _ = ReadSyncMethods((*sysinfoFileReadSyncMethods)(nil))

func NewSysinfoFileMethods(_ *config.ClusterConfig) *sysinfoFileReadSyncMethods {
	return &sysinfoFileReadSyncMethods{}
}

func (_ *sysinfoFileReadSyncMethods) IsCacheable() bool {
	return false
}

func (sfr *sysinfoFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	var _ = payload.(sysinfoPayloadType)
	return payload
}

func (sfr *sysinfoFileReadSyncMethods) ReadDataLockedAndRectify(
	attr FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var p sysinfoPayloadType
	if (attr & FileSysinfoV0JSON) != 0 {
		p, err = parse.ParseSysinfoV0JSON(inputFile, verbose)
	} else {
		p, err = parse.ParseSysinfoOldJSON(inputFile, verbose)
	}
	payload = p
	return
}

var (
	// MT: Constant after initialization; immutable
	perSysinfoSize int64
)

func init() {
	var s repr.SysinfoData
	perSysinfoSize = int64(unsafe.Sizeof(s))
}

func (_ *sysinfoFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(sysinfoPayloadType) // []*config.NodeConfigRecord
	size := unsafe.Sizeof(data)
	size += uintptr(len(data)) * repr.PointerSize
	for _, d := range data {
		size += repr.SysinfoDataSize(d)
	}
	return size
}

func ReadSysinfoSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]sysinfoPayloadType, int, error) {
	return ReadRecordsFromFiles[repr.SysinfoData](files, verbose, reader)
}
