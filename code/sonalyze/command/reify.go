// The reifier is used to build a command line for remote execution from parsed and checked
// arguments.
//
// Uniformly, repeatable strings that could be comma-separated on input are exploded as separate
// arguments here, to keep it simple.

package command

import (
	"fmt"
	"net/url"
)

type Reifier struct {
	options string
}

func NewReifier() Reifier {
	return Reifier{""}
}

// This must equal `magicBoolean` in the `sonalyzed` sources.
const MagicBoolean = "xxxxxtruexxxxx"

func (r *Reifier) addString(name, val string) {
	if r.options != "" {
		r.options += "&"
	}
	r.options += url.QueryEscape(name)
	r.options += "="
	r.options += url.QueryEscape(val)
}

func (r *Reifier) Bool(n string, v bool) {
	if v {
		r.addString(n, MagicBoolean)
	}
}

func (r *Reifier) Uint(n string, v uint) {
	if v != 0 {
		r.addString(n, fmt.Sprint(v))
	}
}

func (r *Reifier) Float64(n string, v float64) {
	if v != 0 {
		r.addString(n, fmt.Sprint(v))
	}
}

func (r *Reifier) UintUnchecked(n string, v uint) {
	r.addString(n, fmt.Sprint(v))
}

func (r *Reifier) String(n, v string) {
	if v != "" {
		r.addString(n, v)
	}
}

func (r *Reifier) RepeatableString(n string, vs []string) {
	for _, v := range vs {
		r.String(n, v)
	}
}

func (r *Reifier) RepeatableUint32(n string, vs []uint32) {
	for _, v := range vs {
		r.Uint(n, uint(v))
	}
}

func (r *Reifier) EncodedArguments() string {
	return r.options
}
