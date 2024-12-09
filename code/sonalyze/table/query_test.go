package table

import (
	"testing"
)

func TestParse(t *testing.T) {
	n, err := ParseQuery(`a=37`)
	if err != nil {
		t.Fatal(err)
	}
	bin := n.(*binaryOp)
	if bin.field != "a" || bin.value != "37" {
		t.Fatal(bin)
	}
}
