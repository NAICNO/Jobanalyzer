// Adapter for reading and caching cluzter-info files.

package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

type cluzterPayloadType = []*CluzterInfo

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
	p, softErrors, err = ParseCluzterV0JSON(inputFile, verbose)
	payload = p
	return
}

func (_ *cluzterFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(cluzterPayloadType)
	size := unsafe.Sizeof(data)
	// Pointers to CluzterInfo
	size += uintptr(len(data)) * pointerSize
	// Every CluzterInfo is different
	for _, d := range data {
		size += cluzterInfoSize(d)
	}
	return size
}

func readCluzterSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]cluzterPayloadType, int, error) {
	return readRecordsFromFiles[CluzterInfo](files, verbose, reader)
}
