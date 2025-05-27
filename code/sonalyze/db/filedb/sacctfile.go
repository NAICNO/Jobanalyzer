// Adapter for reading and caching sacct-info files.

package filedb

import (
	"io"
	"unsafe"

	"go-utils/config"
	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
)

type sacctPayloadType = []*repr.SacctInfo

type sacctFileReadSyncMethods struct {
}

var _ = ReadSyncMethods((*sacctFileReadSyncMethods)(nil))

func NewSacctFileMethods(_ *config.ClusterConfig) *sacctFileReadSyncMethods {
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
	attr FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var p sacctPayloadType
	if (attr & FileSlurmV0JSON) != 0 {
		p, softErrors, err = parse.ParseSlurmV0JSON(inputFile, uf, verbose)
	} else {
		p, softErrors, err = parse.ParseSlurmCSV(inputFile, uf, verbose)
	}
	payload = p
	return
}

func (_ *sacctFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(sacctPayloadType) // []*repr.SacctInfo
	size := unsafe.Sizeof(data)
	// Pointers to SacctInfo
	size += uintptr(len(data)) * repr.PointerSize
	// Every SacctInfo is the same
	size += uintptr(len(data)) * repr.SizeofSacctInfo
	return size
}

func ReadSacctSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([]sacctPayloadType, int, error) {
	return ReadRecordsFromFiles[repr.SacctInfo](files, verbose, reader)
}
