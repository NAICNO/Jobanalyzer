// Adapter for reading and caching sacct-info files.

package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

type sacctFileReadSyncMethods struct {
}

func newSacctFileMethods(_ *config.ClusterConfig) *sacctFileReadSyncMethods {
	return &sacctFileReadSyncMethods{}
}

func (_ *sacctFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *sacctFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	return payload
}

func (sfr *sacctFileReadSyncMethods) ReadDataLockedAndRectify(
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	payload, softErrors, err = ParseSacctLog(inputFile, uf, verbose)
	return
}

var (
	// MT: Constant after initialization; immutable
	perSacctSize int64
)

func init() {
	var s SacctInfo
	perSacctSize = int64(unsafe.Sizeof(s) + unsafe.Sizeof(&s))
}

func (_ *sacctFileReadSyncMethods) CachedSizeOfPayload(payload any) int64 {
	data := payload.([]*SacctInfo)
	return perSampleSize * int64(len(data))
}

func readSacctSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]*SacctInfo, int, error) {
	return readRecordsFromFiles[SacctInfo](files, verbose, reader)
}
