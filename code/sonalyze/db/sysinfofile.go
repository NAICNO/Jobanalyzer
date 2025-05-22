package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

type sysinfoPayloadType = []*config.NodeConfigRecord

type sysinfoFileReadSyncMethods struct {
}

var _ = ReadSyncMethods((*sysinfoFileReadSyncMethods)(nil))

func newSysinfoFileMethods(_ *config.ClusterConfig) *sysinfoFileReadSyncMethods {
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
		p, err = ParseSysinfoV0JSON(inputFile, verbose)
	} else {
		p, err = ParseSysinfoOldJSON(inputFile, verbose)
	}
	payload = p
	return
}

var (
	// MT: Constant after initialization; immutable
	perSysinfoSize int64
)

func init() {
	var s config.NodeConfigRecord
	perSysinfoSize = int64(unsafe.Sizeof(s))
}

func (_ *sysinfoFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(sysinfoPayloadType) // []*config.NodeConfigRecord
	size := unsafe.Sizeof(data)
	size += uintptr(len(data)) * pointerSize
	for _, d := range data {
		size += sysinfoInfoSize(d)
	}
	return size
}

func sysinfoInfoSize(d *config.NodeConfigRecord) uintptr {
	size := unsafe.Sizeof(*d)
	size += uintptr(len(d.Timestamp))
	size += uintptr(len(d.Hostname))
	size += uintptr(len(d.Description))
	// Ignore Metadata
	return size
}

func readNodeConfigRecordSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]sysinfoPayloadType, int, error) {
	return readRecordsFromFiles[config.NodeConfigRecord](files, verbose, reader)
}
