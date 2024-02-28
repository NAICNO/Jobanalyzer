// Simple dense set-of-numbers as bitvector.
//
// For some reason I'm using 24 bits per element, this is probably not optimal.  (32 is dangerous
// due to how browsers represent integers, and 16 seemed too little but might be better.)
//
// See bitset_test.js for test cases.

function Bitset(size, allSet) {
    this.elts = new Array(Math.ceil(size/24))
    this.length = size
    for ( let i=0 ; i < this.elts.length ; i++ )
        this.elts[i] = 0
    if (allSet)
        this.fill()
}

Bitset.prototype.fill = function() {
    let full = Math.floor(this.length/24)
    for (let i=0 ; i < full ; i++) {
        this.elts[i] = 0xFFFFFF
    }
    if ((this.length % 24) != 0) {
        this.elts[this.elts.length-1] = (1 << (this.length % 24)) - 1
    }
}

Bitset.prototype.enumerate = function(f) {
    for (let k = 0 ; k < this.elts.length ; k++) {
        let v = this.elts[k]
        if (v != 0) {
            for (let j = 0 ; j < 24 ; j++) {
                if (v & (1 << j)) {
                    f((k * 24) + j)
                }
            }
        }
    }
}

Bitset.prototype.setBit = function (n) {
    this.elts[(n / 24)|0] |= 1 << (n % 24)
}

Bitset.prototype.clearBit = function (n) {
    this.elts[(n / 24)|0] &= ~(1 << (n % 24))
}

Bitset.prototype.isSet = function (n) {
    return (this.elts[(n / 24)|0] & (1 << (n %24))) != 0
}

Bitset.prototype.toString = function () {
    let xs = []
    this.enumerate(k => xs.push(k))
    return xs.toString()
}

Bitset.prototype.toArray = function () {
    let xs = new Array(this.length)
    for ( let i = 0 ; i < xs.length ; i++ )
        xs[i] = 0
    this.enumerate(k => xs[k] = 1)
    return xs
}

function bitsetUnion(a, b) {
    let result = new Bitset(Math.max(a.length, b.length))
    for (let i=0 ; i < a.length ; i++) {
        result.elts[i] = a.elts[i]
    }
    for (let i=0 ; i < b.length ; i++) {
        result.elts[i] |= b.elts[i]
    }
    return result
}

function bitsetIntersection(a, b) {
    // The reason the length is the max of the two lengths is because it simplifies the client for
    // all sets to have the same length, notably bitset complement requires knowing the set of the
    // universe.
    let result = new Bitset(Math.max(a.length, b.length))
    for (let i=0 ; i < result.elts.length; i++) {
        let v0 = a.elts.length >= i ? a.elts[i] : 0
        let v1 = b.elts.length >= i ? b.elts[i] : 0
        result.elts[i] = v0 & v1
    }
    return result
}
