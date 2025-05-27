// Adapter for reading and caching cluzter-info files.

package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
)

type cluzterPayloadType = []*repr.CluzterInfo

type cluzterFileReadSyncMethods struct {
}

func newCluzterFileMethods(_ *config.ClusterConfig) *cluzterFileReadSyncMethods {
	return &cluzterFileReadSyncMethods{}
}

func (_ *cluzterFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *cluzterFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	var _ = payload.(cluzterPayloadType)
	return payload
}

func (sfr *cluzterFileReadSyncMethods) ReadDataLockedAndRectify(
	_ FileAttr,
	inputFile io.Reader,
	_ *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var p cluzterPayloadType
	p, softErrors, err = parse.ParseCluzterV0JSON(inputFile, verbose)
	payload = p
	return
}

func (_ *cluzterFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(cluzterPayloadType)
	size := unsafe.Sizeof(data)
	// Pointers to CluzterInfo
	size += uintptr(len(data)) * repr.PointerSize
	// Every CluzterInfo is different
	for _, d := range data {
		size += repr.CluzterInfoSize(d)
	}
	return size
}

func readCluzterSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]cluzterPayloadType, int, error) {
	return readRecordsFromFiles[repr.CluzterInfo](files, verbose, reader)
}
