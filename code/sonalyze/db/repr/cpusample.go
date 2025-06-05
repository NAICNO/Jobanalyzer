// Data representations of both cached and expanded per-CPU sample values.  The data representation
// produced here is a 2-dimensional volume:
//
//   time host per-core-cpu-time ...
//   ...
//
// where per-core-cpu-time is the CPU time on that core at that time, since boot.  The rows for a
// given host are all the same length but as hosts have different numbers of cores the cores for
// different hosts can have different length.  (An alternative view is therefore that this is a
// 3-column 3-dimensional volume where the third column is some variable-length array.)

package repr

import (
	"errors"
	"unsafe"

	. "sonalyze/common"
)

// CpuSamples represents CPU sample data for all the cores on a host at a point in time.
// This is the cached representation, hence there is a level of indirection here for the encoding as
// that can vary based on the underlying data representation (older and newer on-disk data encode
// these differently).  A simple time series table of per-CPU samples can be obtained by decoding
// the encoded data and prepending the host and time to each resulting row.
type CpuSamples struct {
	Timestamp int64
	Hostname  Ustr
	Encoded   EncodedCpuSamples
}

func (l *CpuSamples) Size() uintptr {
	size := unsafe.Sizeof(*l)
	a, b := decodeEncodedCpuSamples(l.Encoded)
	if a != nil {
		size += uintptr(len(a))
	} else {
		size += uintptr(len(b)) * 8
	}
	return size
}

// EncodedCpuSamples is a union of the different representations of CPU sample data.
type EncodedCpuSamples struct {
	x any
}

func EncodedCpuSamplesFromBytes(xs []byte) EncodedCpuSamples {
	return EncodedCpuSamples{x: xs}
}

func EncodedCpuSamplesFromValues(xs []uint64) EncodedCpuSamples {
	return EncodedCpuSamples{x: xs}
}

func DecodeEncodedCpuSamples(edata EncodedCpuSamples) ([]uint64, error) {
	data, decodedVals := decodeEncodedCpuSamples(edata)
	if decodedVals != nil {
		return decodedVals, nil
	}
	return decodeBase45CpuSamples(data)
}

func decodeEncodedCpuSamples(e EncodedCpuSamples) ([]byte, []uint64) {
	switch xs := e.x.(type) {
	case []byte:
		return xs, nil
	case []uint64:
		return nil, xs
	default:
		panic("Unexpected")
	}
}

// Decode base-45 delta-encoded data, see Sonar documentation.

const (
	base       = 45
	none       = uint8(255)
	initial    = "(){}[]<>+-abcdefghijklmnopqrstuvwxyz!@#$%^&*_"
	subsequent = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ~|';:.?/`"
)

var (
	initialVal    = make([]byte, 256)
	subsequentVal = make([]byte, 256)
	decodeError   = errors.New("Could not decode load datum")
	noDataError   = errors.New("Empty data array")
)

func init() {
	for i := 0; i < 255; i++ {
		initialVal[i] = none
		subsequentVal[i] = none
	}
	for i := byte(0); i < base; i++ {
		initialVal[initial[i]] = i
		subsequentVal[subsequent[i]] = i
	}
}

func decodeBase45CpuSamples(data []byte) ([]uint64, error) {
	var (
		// shift==0 means no value
		val, shift uint64
		vals       = make([]uint64, 0, len(data)*3)
	)
	for _, c := range data {
		if initialVal[c] != none {
			if shift != 0 {
				vals = append(vals, val)
			}
			val = uint64(initialVal[c])
			shift = base
			continue
		}
		if subsequentVal[c] == none {
			return nil, decodeError
		}
		val += uint64(subsequentVal[c]) * shift
		shift *= base
	}
	if shift != 0 {
		vals = append(vals, val)
	}
	if len(vals) == 0 {
		return nil, noDataError
	}
	minVal := vals[0]
	for i := 1; i < len(vals); i++ {
		vals[i] += minVal
	}
	return vals[1:], nil
}
