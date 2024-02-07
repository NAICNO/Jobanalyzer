// General-ish query engine for node data.
//
// The purpose of this is to allow large and undisplayable sets of data rows (typically one row per
// node but it could also be used to query jobs) to be culled to something more manageable.
//
// It works by compiling a simple boolean/relational query under an environment into a query matcher
// object, that can then be applied to data rows.
//
// The query matcher takes as its input an immutable table of data objects (rows) and a
// representation of the set of rows (indices) from that set that are currently considered.  It
// returns a new set, the result of the filter, a subset of the input set.  (It is based on a set of
// rows rather than a single row because culling by eg node names in terms evaluated early will tend
// to quickly reduce the number of data rows considered by terms evaluated later, but that
// optimization is not currently implemented.)
//
// For example, the primitive query "(mem% > 50)" constructs a set that comprises those elements in
// the input set whose mem% value is greater than 50.  For another example, the primitive query
// "c1-*" constructs a set that comprises those elements in the input set whose node names match
// that pattern.  The query "c1-* and mem% > 50" combines them.  "login and down" shows the login
// nodes that are currently believed to be down.
//
// Note that the actual defined variable terms ("mem%" etc) and predefined operators for node
// selection ("login", "down", etc) are defined by the client of this library.

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Query compiler.
//
// The `query` is the query expression.  The `knownFields` is a map from known field names to either
// `true` or to a canonical field name (allowing for aliases).  The `builtinOperations` is a map
// from expression aliases (essentially subroutines) to their matcher expandings.
//
// Returns either a query matcher object, or a string describing the error as precisely as possible.
//
// The expression grammar is simple:
//
//   expr ::= token binop token
//          | unop token
//          | token
//          | (expr)
//
//   binop ::= "and" | "or" | "<" | ">" | "<=" | ">=" | "="
//   unop ::= "~"
//   token ::= string of non-whitespace chars other than punctuation, except "and" and "or"
//   punctuation ::= "<", ">", "=", "~", "(", ")"
//
// Everything is case-sensitive.
//
// Tokens represent either known field references, numbers, or node name wildcard matchers.  The
// interpretation of a token is contextual: a field name or number is interpreted as such only in
// the context of a binary operator that requires that interpretation; in all other contexts they
// are interpreted as node name matchers.  That is, "c1* > 5 and 37.5" is a legal and meaningful
// instruction if there is a known field called "c1*" and a node called "37.5".
//
// Binary operator precedence from low to high: OR, AND, relationals.  Unary ops bind tighter
// than binary ops.
//
// For a relational operator, the first argument must be a field name and the second must be a
// number.
//
// The meaning of an expression is that each node name matcher or relational operation induces a
// subset of the data rows, "and" is set intersection, "or" is set union, and "~" is set complement
// (where the full set of nodes is the universe).

function compileQuery(query, knownFields, builtinOperations) {
    if (knownFields === undefined) {
        knownFields = {}
    }
    if (builtinOperations === undefined) {
        builtinOperations = {}
    }

    // Character indices in `query`
    let i = 0
    let lim = query.length

    function isSpace(c) {
        return c == ' ' || c == '\t' || c == '\n' || c == '\r'
    }

    function isDelim(c) {
        return c == '(' || c == ')'
    }

    function isPunct(c) {
        return c == '<' || c == '>' || c == '=' || c == "~"
    }

    function isToken(c) {
        return !isDelim(c) && !isPunct(c) && !isSpace(c)
    }

    // Location of last token gotten, used for error reporting.
    let loc = 0

    // Token stream, with start-of-token location, -1 means "no token yet"
    let pending = ""
    let pending_i = -1

    var operatorNames = { "<":true, "<=":true, ">":true, ">=":true, "=":true, "and":true, "or":true, "~":true }

    // Place a token in pending if not yet there, return false if no more tokens could be extracted.
    function fill() {
        if (pending_i >= 0) {
            return true
        }
        while (i < lim && isSpace(query.charAt(i))) {
            i++
        }
        if (i == lim) {
            return false
        }
        pending_i = i
        let probe = query.charAt(i)
        if (isDelim(probe)) {
            pending = probe
            i++
            return true
        }
        let isOperator = false
        if (isPunct(probe)) {
            i++
            while (i < lim && isPunct(query.charAt(i))) {
                i++
            }
            isOperator = true
        } else if (isToken(probe)) {
            i++
            while (i < lim && isToken(query.charAt(i))) {
                i++
            }
        } else {
            fail(`Unexpected character '${probe}'`)
        }
        pending = query.substring(pending_i, i)
        if (isOperator && !(pending in operatorNames)) {
            fail(`Unknown operator '${pending}'`)
        }
        return true
    }

    function get() {
        if (!fill()) {
            loc = i
            fail("Unexpected end of expression")
        }
        let s = pending
        loc = pending_i
        pending_i = -1
        return s
    }

    function fail(irritant) {
        throw `Location ${loc+1}: ${irritant}`
    }

    function eatToken(t) {
        if (!fill()) {
            return false
        }
        if (pending == t) {
            get()
            return true
        }
        return false
    }

    function exprBin(next, constructor, ...ops) {
        return function () {
            let e = next()
        Outer:
            for (;;) {
                for (let op of ops) {
                    if (eatToken(op)) {
                        let e2 = next()
                        try {
                            e = new constructor(op, e, e2)
                        } catch (ex) {
                            fail(ex)
                        }
                        continue Outer
                    }
                }
                break
            }
            return e
        }
    }

    function exprPrim() {
        if (eatToken("(")) {
            let e = exprOr()
            if (!eatToken(")")) {
                fail("Expected ')' here")
            }
            return e
        }
        if (eatToken("~")) {
            let e = exprPrim()
            return new notOperation(e)
        }
        let t = get()
        let probe = parseFloat(t)
        if (isFinite(probe)) {
            return probe
        }
        while (knownFields.hasOwnProperty(t)) {
            let mapping = knownFields[t]
            if (mapping === true)
                return t
            if (typeof mapping != "string") {
                // something strange
                break
            }
            // alias
            t = mapping
        }
        if (t in builtinOperations) {
            return builtinOperations[t]
        }
        if (t in operatorNames || t == "(" || t == ")") {
            fail(`Misplaced operator or punctuation '${t}'`)
        }
        return new nodeglob(t)
    }

    var exprRelop = exprBin(exprPrim, relOperation, "<", "<=", ">", ">=", "=")
    var exprAnd = exprBin(exprRelop, setOperation, "and")
    var exprOr = exprBin(exprAnd, setOperation, "or")

    function expr() {
        let e = exprOr()
        if (fill()) {
            fail(`Junk at end of expression: ${get()}`)
        }
        return e
    }

    try {
        return expr()
    } catch (ex) {
        return String(ex)
    }
}

function relOperation(op, fld, n) {
    if (!(typeof fld == "string" && typeof n == "number")) {
        throw `Wrong type of arguments to relational operator ${op}`
    }
    this.op = op
    this.fld = fld
    this.n = n
    switch (op) {
    case "<":
        this.fn = (a) => a < n
        break
    case "<=":
        this.fn = (a) => a <= n
        break
    case ">":
        this.fn = (a) => a > n
        break
    case ">=":
        this.fn = (a) => a >= n
        break
    case "=":
        this.fn = (a) => a == n
        break
    default:
        throw `Internal error`
    }
}

relOperation.prototype.toString = function () {
    return `(${this.op} ${this.fld} ${this.n})`
}

relOperation.prototype.eval = function (data, elems) {
    let result = new bitset(elems.length)
    let fn = this.fn
    let fld = this.fld
    elems.enumerate(function (n) {
        if (fn(data[n][fld])) {
            result.setBit(n)
        }
    })
    return result
}

function notOperation(e) {
    this.e = e
}

notOperation.prototype.toString = function() {
    return `(~ ${this.e})`
}

// not(e) is (universe \ e)
notOperation.prototype.eval = function (data, elems) {
    let result = new bitset(data.length)
    result.fill()
    let xs = this.e.eval(data, elems)
    xs.enumerate(function (n) {
        result.clearBit(n)
    })
    return result
}

function nodeglob(g) {
    this.glob = g
    this.matcher = makeGlobber(g)
}

nodeglob.prototype.toString = function () {
    return `(node ${this.glob})`
}

nodeglob.prototype.eval = function (data, elems) {
    let result = new bitset(elems.length)
    let matcher = this.matcher
    elems.enumerate(function (n) {
        if (data[n].hostname.match(matcher)) {
            result.setBit(n)
        }
    })
    return result
}

function makeGlobber(g) {
    let r = "^"
    // FIXME: Escape special characters
    // FIXME: Ranges of digits, as per normal
    for ( let i=0 ; i < g.length ; i++ ) {
        if (g.charAt(i) == '*') {
            r += ".*"
        } else {
            r += g.charAt(i)
        }
    }
    r += "$"
    return new RegExp(r)
}

function setOperation(op, e1, e2) {
    this.op = op
    this.e1 = e1
    this.e2 = e2
}

setOperation.prototype.toString = function () {
    return `(${this.op} ${this.e1} ${this.e2})`
}

setOperation.prototype.eval = function (data, elems) {
    if (this.op == "or") {
        let v1 = this.e1.eval(data, elems)
        let v2 = this.e2.eval(data, elems)
        return bitsetUnion(v1, v2)
    }

    let v1 = this.e1.eval(data, elems)
    let v2 = this.e2.eval(data, elems)
    return bitsetIntersection(v1, v2)
}

///////////////////////////////////////////////////////////////////////////////////////////////
//
// Simple dense set-of-numbers as bitvector.
//
// For some reason I'm using 24 bits per element, this is probably not optimal.  (32 is dangerous
// due to how browsers represent integers, and 16 seemed too little but might be better.)

function bitset(size) {
    this.elts = new Array(Math.ceil(size/24))
    this.length = size
    for ( let i=0 ; i < this.elts.length ; i++ )
        this.elts[i] = 0
}

bitset.prototype.fill = function() {
    let full = Math.floor(this.length/24)
    for (let i=0 ; i < full ; i++) {
        this.elts[i] = 0xFFFFFF
    }
    if ((this.length % 24) != 0) {
        this.elts[this.elts.length-1] = (1 << (this.length % 24)) - 1
    }
}

bitset.prototype.enumerate = function(f) {
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

bitset.prototype.setBit = function (n) {
    this.elts[(n / 24)|0] |= 1 << (n % 24)
}

bitset.prototype.clearBit = function (n) {
    this.elts[(n / 24)|0] &= ~(1 << (n % 24))
}

bitset.prototype.isSet = function (n) {
    return (this.elts[(n / 24)|0] & (1 << (n %24))) != 0
}

function bitsetUnion(a, b) {
    let result = new bitset(Math.max(a.length, b.length))
    for (let i=0 ; i < a.length ; i++) {
        result.elts[i] = a.elts[i]
    }
    for (let i=0 ; i < b.length ; i++) {
        result.elts[i] |= b.elts[i]
    }
    return result
}

function bitsetIntersection(a, b) {
    let result = new bitset(Math.max(a.length, b.length))
    for (let i=0 ; i < result.elts.length; i++) {
        let v0 = a.elts.length >= i ? a.elts[i] : 0
        let v1 = b.elts.length >= i ? b.elts[i] : 0
        result.elts[i] = v0 & v1
    }
    return result
}

// For console use
function testBitset() {
    function assertEq(a, b) {
        if (a != b) {
            throw new Error(`Failed assertion: ${a} ${b}`)
        }
    }

    let bs = new bitset(33)
    assertEq(bs.length, 33)
    assertEq(bs.elts.length, 2)
    assertEq(bs.elts[0], 0)
    assertEq(bs.elts[1], 0)

    bs.fill()
    assertEq(bs.elts[0], 0xFFFFFF)
    assertEq(bs.elts[1], 0x0001FF)

    bs = new bitset(33)
    bs.setBit(1)
    bs.setBit(8)
    bs.setBit(27)
    assertEq(bs.elts[0], (1 << 1) | (1 << 8))
    assertEq(bs.elts[1], (1 << (27 - 24)))

    let bs2 = new bitset(32)
    bs2.setBit(2)
    bs2.setBit(8)
    bs2.setBit(24)
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

    let bs4 = new bitset(16)
    bs4.setBit(1)
    bs4.setBit(2)
    let bs5 = bitsetIntersection(bs3, bs4)
    assertEq(bs5.length, 16)
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
}
