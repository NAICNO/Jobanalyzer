package common

const DEBUG = false

func Assert(c bool, s string) {
	if !c {
		panic(s)
	}
}
