// For console or JS shell use, see run_tests.js
//
// Dependencies: assertions, defined in run_tests.js

function testBitset() {
    let bs = new Bitset(33)
    assertEq(bs.length, 33)
    assertEq(bs.elts.length, 2)
    assertEq(bs.elts[0], 0)
    assertEq(bs.elts[1], 0)

    bs.fill()
    assertEq(bs.elts[0], 0xFFFFFF)
    assertEq(bs.elts[1], 0x0001FF)

    bs = new Bitset(33)
    bs.setBit(1)
    bs.setBit(8)
    bs.setBit(27)
    assertEq(bs.elts[0], (1 << 1) | (1 << 8))
    assertEq(bs.elts[1], (1 << (27 - 24)))

    let bs2 = new Bitset(32)
    bs2.setBit(2)
    bs2.setBit(8)
    bs2.setBit(24)
    assertEqual(bs2.toArray(), [0,0,1,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0])
    assertEqual(bs2.toString(), "2,8,24")

    let bs3 = bitsetUnion(bs, bs2)
    assertEq(bs3.length, 33)
    assertEq(bs3.elts[0], (1 << 1) | (1 << 2) | (1 << 8))
    assertEq(bs3.elts[1], (1 << (24 - 24)) | (1 << (27 - 24)))
    assertEq(bs3.isSet(1), true)
    assertEq(bs3.isSet(2), true)
    assertEq(bs3.isSet(8), true)
    assertEq(bs3.isSet(24), true)
    assertEq(bs3.isSet(27), true)
    assertEq(bs3.isSet(0), false)
    assertEq(bs3.isSet(3), false)
    assertEq(bs3.isSet(25), false)

    let bs4 = new Bitset(16)
    bs4.setBit(1)
    bs4.setBit(2)
    let bs5 = bitsetIntersection(bs3, bs4)
    assertEq(bs5.length, 33)
    assertEq(bs5.elts[0], (1 << 1) | (1 << 2))

    let res = []
    bs3.enumerate(function (x) { res.push(x) })
    assertEq(res.length, 5)
    assertEq(res[0], 1)
    assertEq(res[1], 2)
    assertEq(res[2], 8)
    assertEq(res[3], 24)
    assertEq(res[4], 27)

    bs3.clearBit(1)
    bs3.clearBit(27)
    assertEq(bs3.isSet(1), false)
    assertEq(bs3.isSet(2), true)
    assertEq(bs3.isSet(8), true)
    assertEq(bs3.isSet(24), true)
    assertEq(bs3.isSet(27), false)
    assertEq(bs3.isSet(0), false)
    assertEq(bs3.isSet(3), false)
    assertEq(bs3.isSet(25), false)

    let bs6 = new Bitset(4, true)
    assertTrue(bs6.isSet(0))
    assertTrue(bs6.isSet(1))
    assertTrue(bs6.isSet(2))
    assertTrue(bs6.isSet(3))
    assertEqual(bs6.toArray(), [1,1,1,1])
    assertEqual(bs6.toString(), "0,1,2,3")
}
