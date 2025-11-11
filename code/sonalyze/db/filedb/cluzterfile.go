// Adapter for reading and caching cluzter-info files.

package filedb

import (
	"io"
	"unsafe"

	. "sonalyze/common"
	"sonalyze/db/parse"
	"sonalyze/db/repr"
	"sonalyze/db/special"
)

type CluzterDataNeeded int

const (
	DataNeedAttributesData CluzterDataNeeded = iota
	DataNeedPartitionsData
	DataNeedNodesData
)

type cluzterData struct {
	attributeData []*repr.CluzterAttributes
	partitionData []*repr.CluzterPartitions
	nodeData      []*repr.CluzterNodes
}

type cluzterPayloadType = *cluzterData

type cluzterFileReadSyncMethods struct {
	dataNeeded CluzterDataNeeded
}

var _ = ReadSyncMethods((*cluzterFileReadSyncMethods)(nil))

type CluzterFileKind int

const (
	CluzterFileKindAttributeData CluzterFileKind = iota
	CluzterFileKindPartitionData
	CluzterFileKindNodeData
)

func NewCluzterFileMethods(_ special.ClusterMeta, kind CluzterFileKind) *cluzterFileReadSyncMethods {
	switch kind {
	case CluzterFileKindAttributeData:
		return &cluzterFileReadSyncMethods{DataNeedAttributesData}
	case CluzterFileKindPartitionData:
		return &cluzterFileReadSyncMethods{DataNeedPartitionsData}
	case CluzterFileKindNodeData:
		return &cluzterFileReadSyncMethods{DataNeedNodesData}
	default:
		panic("Unexpected")
	}
}

func (_ *cluzterFileReadSyncMethods) IsCacheable() bool {
	return true
}

func (sfr *cluzterFileReadSyncMethods) SelectDataFromPayload(payload any) (data any) {
	switch sfr.dataNeeded {
	case DataNeedAttributesData:
		return payload.(cluzterPayloadType).attributeData
	case DataNeedPartitionsData:
		return payload.(cluzterPayloadType).partitionData
	case DataNeedNodesData:
		return payload.(cluzterPayloadType).nodeData
	default:
		panic("Unexpected")
	}
}

func (sfr *cluzterFileReadSyncMethods) ReadDataLocked(
	_ FileAttr,
	inputFile io.Reader,
	_ *UstrCache,
	verbose bool,
) (payload any, softErrors int, err error) {
	var attributes []*repr.CluzterAttributes
	var partitions []*repr.CluzterPartitions
	var nodes []*repr.CluzterNodes
	attributes, partitions, nodes, softErrors, err = parse.ParseCluzterV0JSON(inputFile, verbose)
	if err != nil {
		return
	}
	payload = &cluzterData{attributes, partitions, nodes}
	return
}

func (_ *cluzterFileReadSyncMethods) CachedSizeOfPayload(payload any) uintptr {
	data := payload.(cluzterPayloadType)
	size := unsafe.Sizeof(data)
	for _, p := range data.attributeData {
		size += p.Size()
	}
	for _, p := range data.partitionData {
		size += p.Size()
	}
	for _, p := range data.nodeData {
		size += p.Size()
	}
	return size
}

func ReadCluzterAttributeDataSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*repr.CluzterAttributes, int, error) {
	return readRecordsFromFiles[repr.CluzterAttributes](files, verbose, reader)
}

func ReadCluzterPartitionDataSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*repr.CluzterPartitions, int, error) {
	return readRecordsFromFiles[repr.CluzterPartitions](files, verbose, reader)
}

func ReadCluzterNodeDataSlice(
	files []*LogFile,
	verbose bool,
	reader ReadSyncMethods,
) ([][]*repr.CluzterNodes, int, error) {
	return readRecordsFromFiles[repr.CluzterNodes](files, verbose, reader)
}
