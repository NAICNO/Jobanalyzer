package parse

import (
	"bytes"
	. "sonalyze/common"
)

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
