// The ArgReifier is used to build a command line for remote execution from parsed and checked
// arguments.
//
// Uniformly, repeatable strings that could be comma-separated on input are exploded as separate
// arguments here, to keep it simple.
//
// See the "REST protocol" section of ../TECHNICAL.md for a definition of the protocol.

package cmd

import (
	"fmt"
	"net/url"
)

type ArgReifier struct {
	options string
}

func NewArgReifier() ArgReifier {
	return ArgReifier{""}
}

// MagicBoolean is the old "true" value that was used for flag options in query strings.  We'll
// continue to allow this indefinitely but new code should use just "true" or (for pedants) "false".
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

func (r *ArgReifier) addString(name, val string) {
	if r.options != "" {
		r.options += "&"
	}
	r.options += url.QueryEscape(name)
	r.options += "="
	r.options += url.QueryEscape(val)
}

func (r *ArgReifier) Bool(n string, v bool) {
	if v {
		r.addString(n, "true")
	}
}

func (r *ArgReifier) Uint(n string, v uint) {
	if v != 0 {
		r.addString(n, fmt.Sprint(v))
	}
}

func (r *ArgReifier) Float64(n string, v float64) {
	if v != 0 {
		r.addString(n, fmt.Sprint(v))
	}
}

func (r *ArgReifier) UintUnchecked(n string, v uint) {
	r.addString(n, fmt.Sprint(v))
}

func (r *ArgReifier) String(n, v string) {
	if v != "" {
		r.addString(n, v)
	}
}

func (r *ArgReifier) RepeatableString(n string, vs []string) {
	for _, v := range vs {
		r.String(n, v)
	}
}

func (r *ArgReifier) RepeatableUint32(n string, vs []uint32) {
	for _, v := range vs {
		r.Uint(n, uint(v))
	}
}

func (r *ArgReifier) EncodedArguments() string {
	return r.options
}
