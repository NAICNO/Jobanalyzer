// Adapter for reading and caching sacct-info files.

package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

type sacctPayloadType = []*SacctInfo

type sacctFileReadSyncMethods struct {
}

var _ = ReadSyncMethods((*sacctFileReadSyncMethods)(nil))

func newSacctFileMethods(_ *config.ClusterConfig) *sacctFileReadSyncMethods {
	return &sacctFileReadSyncMethods{}
}

func (_ *sacctFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *sacctFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	var _ = payload.(sacctPayloadType)
	return payload
}

func (sfr *sacctFileReadSyncMethods) ReadDataLockedAndRectify(
	_ FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var p sacctPayloadType
	p, softErrors, err = ParseSlurmCSV(inputFile, uf, verbose)
	payload = p
	return
}

func (_ *sacctFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(sacctPayloadType) // []*SacctInfo
	size := unsafe.Sizeof(data)
	// Pointers to SacctInfo
	size += uintptr(len(data)) * pointerSize
	// Every SacctInfo is the same
	size += uintptr(len(data)) * sizeofSacctInfo
	return size
}

func readSacctSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]sacctPayloadType, int, error) {
	return readRecordsFromFiles[SacctInfo](files, verbose, reader)
}
