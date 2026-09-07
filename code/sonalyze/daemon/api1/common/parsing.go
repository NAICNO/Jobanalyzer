package common

import (
	"strconv"
	"strings"
	"unsafe" // for Sizeof
)

func StringList(s string) []string {
	var xs []string
	for _, x := range strings.Split(s, ",") {
		x = strings.TrimSpace(x)
		if x != "" {
			xs = append(xs, x)
		}
	}
	return xs
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func UintList[T Unsigned](s string) ([]T, error) {
	var xs []T
	var z T
	bits := int(unsafe.Sizeof(z) * 8)
	for _, x := range StringList(s) {
		n, err := strconv.ParseUint(x, 10, bits)
		if err != nil {
			return nil, err
		}
		xs = append(xs, T(n))
	}
	return xs, nil
}
