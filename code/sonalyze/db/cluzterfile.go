// Adapter for reading and caching cluzter-info files.

package db

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
)

type cluzterFileReadSyncMethods struct {
}

func newCluzterFileMethods(_ *config.ClusterConfig) *cluzterFileReadSyncMethods {
	return &cluzterFileReadSyncMethods{}
}

func (_ *cluzterFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *cluzterFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	return payload
}

func (sfr *cluzterFileReadSyncMethods) ReadDataLockedAndRectify(
	_ FileAttr,
	inputFile io.Reader,
	_ *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	payload, softErrors, err = ParseCluzterV0JSON(inputFile, verbose)
	return
}

var (
	// MT: Constant after initialization; immutable
	perCluzterSize int64
)

func init() {
	var s CluzterInfo
	perCluzterSize = int64(unsafe.Sizeof(s) + unsafe.Sizeof(&s))
}

func (_ *cluzterFileReadSyncMethods) CachedSizeOfPayload(payload any) int64 {
	data := payload.([]*CluzterInfo)
	return perSampleSize * int64(len(data))
}

func readCluzterSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*CluzterInfo, int, error) {
	return readRecordsFromFiles[CluzterInfo](files, verbose, reader)
}
