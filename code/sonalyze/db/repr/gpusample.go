// Data representations of both cached and expanded per-GPU sample values.
//
// The data representation produced here is a 3-dimensional volume:
//
//   time host per-card-data ...
//   ...
//
// where per-card-data is itself a record of various card-attr values; the contents of this record
// vary with the source of data (older data or newer data), a bit vector holds metadata about the
// varying parts.  As hosts have different numbers of cards, the row length of the outer
// 2-dimensional table is variable, from 2 (no cards) to some unbounded n (n-2 cards).
//
// There is a straightforward expansion from the 3-dimensional table to a 2-dimensional table with
// uniform row length:
//
//   time host card-uuid card-attr ...
//   ...
//
// except that in older data we do not have the card-uuid in the data, it would need to be looked up
// in sysinfo data using the card's on-node index and the timestamp as the combined lookup index.

package repr

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	. "sonalyze/common"
)

// GpuSamples represents GPU sample data for all the cards on a host at a point in time.
// This is the cached representation, hence there is a level of indirection here for the encoding as
// that can vary based on the underlying data representation (older and newer on-disk data encode
// these differently).  A simple time series table of per-GPU samples can be obtained by decoding
// the encoded data and prepending the host and time to each resulting row.
type GpuSamples struct {
	Timestamp int64
	Hostname  Ustr
	Encoded   EncodedGpuSamples
}

func (g *GpuSamples) Size() uintptr {
	size := unsafe.Sizeof(*g)
	a, b := decodeEncodedGpuSamples(g.Encoded)
	if a != nil {
		size += uintptr(len(a))
	} else {
		// FIXME: Not right since the UUID and ComputeMode are string, not Ustr
		size += uintptr(len(b)) * sizeofPerGpuSample
	}
	return size
}

// The GpuAttr bits are set if certain fields are present
type GpuAttr uint

const (
	GpuHasUuid GpuAttr = 1 << iota
	GpuHasComputeMode
	GpuHasUtil // CEUtil, MemoryUtil
	GpuHasFailing
)

// GPU sample data for a single card (on a host at a given point in time, here implicit).
type PerGpuSample struct {
	Attr GpuAttr
	*newfmt.SampleGpu
}

var (
	// MT: Constant after initialization; immutable
	sizeofPerGpuSample uintptr
)

func init() {
	var x PerGpuSample
	sizeofPerGpuSample = unsafe.Sizeof(x)
}

// EncodedGpuSamples is a union of the different representations of GPU sample data.
type EncodedGpuSamples struct {
	x any
}

func EncodedGpuSamplesFromBytes(xs []byte) EncodedGpuSamples {
	return EncodedGpuSamples{x: xs}
}

func EncodedGpuSamplesFromValues(xs []PerGpuSample) EncodedGpuSamples {
	return EncodedGpuSamples{x: xs}
}

func decodeEncodedGpuSamples(e EncodedGpuSamples) ([]byte, []PerGpuSample) {
	switch xs := e.x.(type) {
	case []byte:
		return xs, nil
	case []PerGpuSample:
		return nil, xs
	default:
		panic("Unexpected")
	}
}

// Decode GPU sample data.
func DecodeEncodedGpuSamples(edata EncodedGpuSamples) (result []PerGpuSample, err error) {
	data, decodedVals := decodeEncodedGpuSamples(edata)
	if decodedVals != nil {
		result = decodedVals
	} else {
		result = decodeCSVEncodedGpuSamples(data)
	}
	return
}

// Decode old-style CSV-encoded GPU sample data.  The input is a comma-separated string of arrays,
// each array represented as a substring tag=x|y|...|z, where the array fields contain no ",".  The
// tag identifies the field:
//
// fan%=27|28|28,perf=P8|P8|P8,musekib=1024|1024|1024,tempc=26|27|28,poww=5|2|20,powlimw=250|250|250,cez=300|300|300,memz=405|405|405
//
// All the arrays must be the same length, ie, contain the same number of "|".
func decodeCSVEncodedGpuSamples(data []byte) []PerGpuSample {
	var result []PerGpuSample
	fields := strings.Split(string(data), ",")
	for _, f := range fields {
		tag, adata, _ := strings.Cut(f, "=")
		data := strings.Split(adata, "|")
		if result == nil {
			result = make([]PerGpuSample, len(data))
		}
		for i := 0; i < len(data); i++ {
			switch tag {
			case "fan%":
				result[i].Fan, _ = strconv.ParseUint(data[i], 10, 64)
			case "perf":
				fmt.Sscanf(data[i], "P%d", &result[i].PerformanceState)
			case "musekib":
				result[i].Memory, _ = strconv.ParseUint(data[i], 10, 64)
			case "tempc":
				result[i].Temperature, _ = strconv.ParseInt(data[i], 10, 64)
			case "poww":
				result[i].Power, _ = strconv.ParseUint(data[i], 10, 64)
			case "powlimw":
				result[i].PowerLimit, _ = strconv.ParseUint(data[i], 10, 64)
			case "cez":
				result[i].CEClock, _ = strconv.ParseUint(data[i], 10, 64)
			case "memz":
				result[i].MemoryClock, _ = strconv.ParseUint(data[i], 10, 64)
			}
		}
	}
	return result
}
