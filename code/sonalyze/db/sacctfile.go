// Adapter for reading and caching sacct-info files.

package db

import (
	"bytes"
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
	attr FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var p sacctPayloadType
	if (attr & FileSlurmV0JSON) != 0 {
		p, softErrors, err = ParseSlurmV0JSON(inputFile, uf, verbose)
	} else {
		p, softErrors, err = ParseSlurmCSV(inputFile, uf, verbose)
	}
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

// The AllocTRES field will usually be present and nonzero and will have a lot of values, so
// allocating the raw field is a bad idea, and allocating much during parsing is also bad.  Use a
// temp buffer to hold the field during construction, and extract only the gres/gpu fields on the
// form model=n with * being inserted for "any" model.  I'm a little uncertain about whether it's
// worth it to hold onto the "any" field if a specific GPU model has been requested, but the
// documentation suggests that I should, and it's better to not editorialize too much here.

var (
	comma   = []byte{','}
	gresGpu = []byte("gres/gpu")
)

func ParseAllocTRES(val []byte, ustrs UstrAllocator, temp []byte) (Ustr, []byte) {
	t := temp[:0]
	for len(val) > 0 {
		before, after, _ := bytes.Cut(val, comma)
		if bytes.HasPrefix(before, gresGpu) {
			if len(t) > 0 {
				t = append(t, ',')
			}
			if before[8] == '=' {
				t = append(t, '*')
				t = append(t, before[8:]...)
			} else {
				// gres/gpu:model=n, skip the :
				t = append(t, before[9:]...)
			}
		}
		val = after
	}
	return ustrs.AllocBytes(t), t
}
