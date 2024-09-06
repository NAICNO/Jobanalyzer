// The reifier is used to build a command line for remote execution from parsed and checked
// arguments.
//
// Uniformly, repeatable strings that could be comma-separated on input are exploded as separate
// arguments here, to keep it simple.
//
// See ../REST.md for a definition of the protocol.

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

// This is the old "true" value that was used for flag options in query strings.  We'll continue to
// allow this indefinitely but new code should use just "true" or (God forbid) "false".
//
// This value existed because Rust's `clap` library insisted that boolean flags should not carry
// parameter values (--some-gpu=true is illegal), and so the command line builder in
// ../daemon/perform.go could not simply append received values to the command line.  But filtering
// out the value "true" uniformly or assuming that empty string meant a true value would be risky.
// And filtering by type would require maintaining name -> type mappings.
//
// With Sonalyze rewritten in Go, this hack is no longer necessary but backward compatibility
// requires us to continue to parse this value.

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
		// TODO: This will change from MagicBoolean to true when the new server has
		// been deployed.
		r.addString(n, MagicBoolean)
		// Future code:
		//r.addString(n, "true")
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
