package repr

import (
	"reflect"
	"testing"
)

func TestDecodeBase45Delta(t *testing.T) {
	// This is the test from the Sonar sources, it's pretty basic.  The string should represent the
	// array [*1, *0, *29, *43, 1, *11] with * denoting an INITIAL char.
	xs, err := DecodeEncodedCpuSamples(EncodedCpuSamplesFromBytes([]byte(")(t*1b")))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(xs, []uint64{1, 30, 89, 12}) {
		t.Fatal("Failed decode")
	}
}
