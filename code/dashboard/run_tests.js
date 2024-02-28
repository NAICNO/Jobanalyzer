// This test program is intended to be run with the Firefox JS shell, it tests the dashboard JS code
// as far as possible outside the browser.
//
// To download the compiled JS shell, go to eg
// https://ftp.mozilla.org/pub/firefox/releases/112.0.2/jsshell/ and grab the one you need.  Here,
// 112.0.2 was the latest Release version of Firefox at the time I downloaded.
//
// To run, just execute `js run_tests.js` in this directory, it will load all the files and run
// all the tests.
//
// To get help about the shell, start the shell as `js` and then evaluate `help()`.
//
// Also see run_tests.sh in this directory.

function assertEq(a, b) {
    if (a === b) {
        return true
    }
    throw new Error(`Failed assertion: ${a} ${b}`)
}

function assertEqual(a, b) {
    if (a == b) {
        return true
    }
    if (a instanceof Array && b instanceof Array) {
        if (a.length == b.length) {
            for ( let i = 0 ; i < a.length ; i++ )
                assertEqual(a[i], b[i])
            return true
        }
        throw new Error(`Wrong lengths`)
    }
    throw new Error(`Failed assertion: ${a} ${b}`)
}

function assertTrue(x) {
    assertEq(x, true)
}

function assertFalse(x) {
    assertEq(x, false)
}

function assertExcept(thunk, s) {
    try {
        thunk()
        throw new Error(`Expected exception with payload "${s}"`)
    } catch (e) {
        if (String(e).indexOf(s) == -1) {
            throw new Error(`Expected payload "${s}" got "${String(e)}"`)
        }
    }
}

load("hostglob.js")
load("hostglob_test.js")
testHostGlobber()
testMultiPatternSplitter()

load("bitset.js")
load("bitset_test.js")
testBitset()

load("queryterm.js") // requires hostglob.js and bitset.js
load("queryterm_test.js")
testQueryterm()

console.log("JS selftests OK!")
