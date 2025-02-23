package table

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	p, err := newQueryParser(
		" = <= < >= > =~ (and) or not andor \"and\" 'or' /not/ `zappa` hi1 ho2 " +
			"\n" +
			" 10 10.5\t-10.5e+7\n-10e8 5w 4d 3h 2m 5w2m 3d2m // '' `` \"\" " +
			"15w12d17h10m",
	)
	assertNotErr(t, err)
	toks := []int{
		tEq, tLe, tLt, tGe, tGt, tMatch, tLparen, tAnd,
		tRparen, tOr, tNot, tIdent, tString, tString, tString, tString, tIdent, tIdent,
		tString, tString, tString, tString, tString, tString, tString, tString, tString, tString,
		tString, tString, tString, tString, tString,
	}
	strs := []string{
		"andor", "and", "or", "not", "zappa", "hi1", "ho2", "10", "10.5", "-10.5e+7", "-10e8",
		"5w", "4d", "3h", "2m", "5w2m", "3d2m", "", "", "", "", "15w12d17h10m",
	}
	j := 0
	for i := range toks {
		var lval yySymType
		tok := p.Lex(&lval)
		assertEq(t, tok, toks[i])
		if toks[i] == tString || toks[i] == tIdent {
			assertEq(t, lval.text, strs[j])
			j++
		}
	}
	var lval yySymType
	assertEq(t, p.Lex(&lval), -1)

	// Error conditions

	type bad struct {
		input, expect string
	}
	for _, b := range []bad{
		bad{"!=", "Unexpected character"},
		bad{"'hi there", "End of input in string"},
		bad{"-", "Non-empty digit string required"},
		bad{"13x", "Token separator required after number"},
		bad{"12.", "Non-empty digit string required"},
		bad{"12.1f", "Token separator required after number"},
		bad{"12e+", "Non-empty digit string required"},
		bad{"12e", "Non-empty digit string required"},
		bad{"12d13", "Time interval designator required"},
		bad{"12h13d", "Time interval designators must appear"},
		bad{"12d13hh", "Token separator required after duration"},
	} {
		p, err := newQueryParser(b.input)
		assertNotErr(t, err)
		var lval yySymType
		p.Lex(&lval)
		if !strings.Contains(p.errtxt, b.expect) {
			t.Fatalf("Should have failed with /%s/: %s, got `%s`", b.expect, b.input, p.errtxt)
		}
	}
}
