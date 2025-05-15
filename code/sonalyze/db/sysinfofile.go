package db

import (
	"io"

	"go-utils/config"
	. "sonalyze/common"
)

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
	return payload
}

func (sfr *sysinfoFileReadSyncMethods) ReadDataLockedAndRectify(
	attr FileAttr,
	inputFile io.Reader,
	uf *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	if (attr & FileSysinfoV0JSON) != 0 {
		payload, err = ParseSysinfoV0JSON(inputFile, verbose)
	} else {
		payload, err = ParseSysinfoOldJSON(inputFile, verbose)
	}
	return
}

func (_ *sysinfoFileReadSyncMethods) CachedSizeOfPayload(payload any) int64 {
	return 0
}

func readNodeConfigRecordSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*config.NodeConfigRecord, int, error) {
	return readRecordsFromFiles[config.NodeConfigRecord](files, verbose, reader)
}
