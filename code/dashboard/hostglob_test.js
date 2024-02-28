// For console or JS shell use, see run_tests.js
//
// Dependencies: assertions, defined in run_tests.js

function testHostGlobber() {
    let g = new HostGlobber("a*[10-30].b[2,4]")
    assertTrue(g.match("a15.b2"))
    assertTrue(g.match("abc27.b4"))
    assertFalse(g.match("a31.b2"))

    g = new HostGlobber("a*[1-5]c.b[10-20]d")
    assertTrue(g.match("a3c.b12d"))
    assertFalse(g.match("a3c.b21d"))
    assertFalse(g.match("a3x.b12d"))

    g = new HostGlobber("c[1-3]-[2,4]")
    assertTrue(g.match("c1-2"))
    assertTrue(g.match("c2-2"))
    assertFalse(g.match("c2-3"))

    assertExcept(_ => new HostGlobber("a[]"), "Invalid number")
    assertExcept(_ => new HostGlobber("a[1-123456789012]"), "Number out of range")
    assertExcept(_ => new HostGlobber("a[1000000-1009999]"), "Expression too large")
    assertExcept(_ => new HostGlobber("a[4-3]"), "Invalid range")
    assertExcept(_ => new HostGlobber("a[1-50000]"), "Range too large")
    assertExcept(_ => new HostGlobber("a[1-5x]"), "Expected ','")
    assertExcept(_ => new HostGlobber("a[1"), "Expected ','")
    assertExcept(_ => new HostGlobber("a,b"), "',' not allowed here")
}

