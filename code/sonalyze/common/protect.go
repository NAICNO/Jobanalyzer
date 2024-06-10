package common

import (
	"fmt"
	"io"
)

// Invoke thunk repeatedly, protecting it against panics.  Panic messages are printed to `log`
// (though if converting the message to string throws, we'll actually exit out of the loop).  The
// outer loop is not necessarily cheap so it may be best if the thunk also has internal looping to
// defray the cost in the common case.

func Forever(thunk func(), log io.Writer) {
	// Manually lambda-lift to avoid allocation overhead in the loop
	d := func() {
		if msg := recover(); msg != nil {
			fmt.Fprintln(log, msg)
		}
	}
	t2 := func() {
		defer d()
		thunk()
	}
	// Catch any problems created by the Fprintln above, to at least tag them sensibly
	defer func() {
		recover()
		panic("PANIC IN CONVERSION OF PANIC MSG; PANICKING!")
	}()
	for {
		t2()
	}
}
