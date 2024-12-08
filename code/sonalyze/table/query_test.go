package table

import (
	_ "fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	p, err := newQueryParser(
		" = <= < >= > =~ (and) or not andor \"and\" 'or' /not/ `zappa` hi1 ho2 " +
			"10 10.5 -10.5e+7 -10e8 5w 4d 3h 2m 5w2m 3d2m",
	)
	assertNotErr(t, err)
	ts := p.tokens
	//fmt.Println(ts)
	toks := []int{
		tEq, tLe, tLt, tGe, tGt, tMatch, tLparen, tAnd,
		tRparen, tOr, tNot, tIdent, tString, tString, tString, tString, tIdent, tIdent,
		tString, tString, tString, tString, tString, tString, tString, tString, tString, tString,
	}
	strs := []string{
		"andor", "and", "or", "not", "zappa", "hi1", "ho2", "10", "10.5", "-10.5e+7", "-10e8",
		"5w", "4d", "3h", "2m", "5w2m", "3d2m",
	}
	j := 0
	assertEq(t, len(toks), len(ts))
	for i := range toks {
		assertEq(t, ts[i].tok, toks[i])
		if toks[i] == tString || toks[i] == tIdent {
			assertEq(t, ts[i].text, strs[j])
			j++
		}
	}
}

func TestParser(t *testing.T) {
	// Basic expr + check that andor is not split as and and or
	n, err := ParseQuery(`a=andor`)
	assertNotErr(t, err)
	bin := n.(*binaryOp)
	assertEq(t, bin.op, opEq)
	assertEq(t, bin.field, "a")
	assertEq(t, bin.value, "andor")

	// Operators + some space and quote stuff
	n, err = ParseQuery(`a< 10`)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opLt)
	assertEq(t, bin.value, "10")

	n, err = ParseQuery(`a <= "="`)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opLe)
	assertEq(t, bin.value, "=")

	// Identifiers are strings, in the right context
	n, err = ParseQuery(`a <= abracadabra`)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opLe)
	assertEq(t, bin.value, "abracadabra")

	n, err = ParseQuery(`abc >'10.('`)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opGt)
	assertEq(t, bin.field, "abc")
	assertEq(t, bin.value, "10.(")

	n, err = ParseQuery(`abc0 >=/hi ho/ `)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opGe)
	assertEq(t, bin.field, "abc0")
	assertEq(t, bin.value, "hi ho")

	n, err = ParseQuery(` abc0 >= 37.5`)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opGe)
	assertEq(t, bin.field, "abc0")
	assertEq(t, bin.value, "37.5")

	// The rightparen is not part of the string literal.  The + is required so as to
	// not interpret the = as an operator.
	n, err = ParseQuery("(ab <= `+=`)")
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opLe)
	assertEq(t, bin.field, "ab")
	assertEq(t, bin.value, "+=")

	n, err = ParseQuery(`User =~ /ec-[x-z]*/`)
	assertNotErr(t, err)
	bin = n.(*binaryOp)
	assertEq(t, bin.op, opMatch)
	assertEq(t, bin.field, "User")
	assertEq(t, bin.value, "ec-[x-z]*")

	// The not binds to the =~ binop and then the and groups that tree and the > binop.
	n, err = ParseQuery(`not User =~ /root|toor|zabbix/ and Duration > 1h`)
	assertNotErr(t, err)
	log := n.(*logicalOp)
	assertEq(t, log.op, opAnd)
	un := log.lhs.(*unaryOp)
	assertEq(t, un.op, opNot)
	bin = un.opd.(*binaryOp)
	assertEq(t, bin.op, opMatch)
	assertEq(t, bin.field, "User")
	assertEq(t, bin.value, "root|toor|zabbix")
	bin = log.rhs.(*binaryOp)
	assertEq(t, bin.op, opGt)
	assertEq(t, bin.field, "Duration")
	assertEq(t, bin.value, "1h")

	// and binds tighter than or
	n, err = ParseQuery(`User = u1 or User = u2 and Duration > 1h`)
	assertNotErr(t, err)
	log = n.(*logicalOp)
	assertEq(t, log.op, opOr)
	log = log.rhs.(*logicalOp)
	assertEq(t, log.op, opAnd)

	// same
	n, err = ParseQuery(`Duration > 1h and User = u1 or User = u2`)
	assertNotErr(t, err)
	log = n.(*logicalOp)
	assertEq(t, log.op, opOr)
	log = log.lhs.(*logicalOp)
	assertEq(t, log.op, opAnd)
}

func assertEq[T comparable](t *testing.T, a, b T) {
	if a != b {
		t.Fatalf("Unequal: %v %v", a, b)
	}
}

func assertNotErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
