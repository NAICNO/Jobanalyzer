// Parser for query expressions.
//
// The grammar is simple:
//
//   expr ::= ident binop string | expr logop expr | unop expr | "(" expr ")"
//
// Idents are the usual [a-zA-Z_][a-zA-Z0-9_]* thing except operator names (and, or, not).  Idents
// always denote fields in a table row.
//
// Strings are either idents, operator names, numbers, durations, or quoted things.  Strings are always
// literal, idents and operator names never denote fields or operators in a string context.
//
// Numbers are full signed floating-point numbers.  Durations are of the form n[wW]m[dD]o[hH]p[mM]
// where all four elements are optional but at least one must be present.
//
// Quoted things can be quoted '...', "...", `...`, or /.../, the quote cannot appear in the quoted
// string.
//
// In primitive binops <, <=, >, >=, =, =~ the first five require the string rhs to be convertible
// to the type of the field given by the ident lhs, and for the last the string is a regular
// expression and the field is formatted(!) to string before matching.  The regex is not augmented
// at all; if you want ^ or $ say, you must add them yourself.
//
// Logical ops "and", "or", and "not" combine other expressions.
//
// TODO: It would have been nice to have an even more permissive string syntax for user convenience.
// This can maybe be done if the parser is made to feed back to the lexer to allow a more permissive
// lexer to lex strings.  Another possibility is to expand the character set allowed by identifiers.

%{
//go:generate goyacc -o queryexpr.go queryexpr.y

package table

import (
	"fmt"
	"regexp"
	"strings"
)

%}

%union {
    text string
    node PNode
}

%start Query

%token <text> tIdent tString
%left <text> tOr
%left <text> tAnd
%nonassoc tEq tLt tLe tGt tGe tMatch
%right <text> tNot
%token tLparen tRparen

%type <node> Expr Query
%type <text> String

%%

Query : Expr { yylex.(*queryParser).expr = $1 } ;

Expr : tNot Expr             { $$ = &unaryOp{opNot, $2} }
     | Expr tOr Expr         { $$ = &logicalOp{opOr, $1, $3} }
     | Expr tAnd Expr        { $$ = &logicalOp{opAnd, $1, $3} }
     | tIdent tEq String     { $$ = &binaryOp{opEq, $1, $3} }
     | tIdent tLt String     { $$ = &binaryOp{opLt, $1, $3} }
     | tIdent tLe String     { $$ = &binaryOp{opLe, $1, $3} }
     | tIdent tGt String     { $$ = &binaryOp{opGt, $1, $3} }
     | tIdent tGe String     { $$ = &binaryOp{opGe, $1, $3} }
     | tIdent tMatch String  { $$ = &binaryOp{opMatch, $1, $3} }
     | tLparen Expr tRparen  { $$ = $2 }
     ;

String : tString | tIdent | tOr | tAnd | tNot ;

%%

type token struct {
	tok  int
	text string
}

type queryParser struct {
	input  string
	tokens []token
	errtxt string
	expr   PNode
}

func (q *queryParser) Lex(lval *yySymType) (tok int) {
	if len(q.tokens) == 0 {
		tok = -1
	} else {
		tok = q.tokens[0].tok
		lval.text = q.tokens[0].text
		q.tokens = q.tokens[1:]
	}
	return
}

func (q *queryParser) Error(s string) {
	if q.errtxt == "" {
		q.errtxt = s
	}
}

func (q *queryParser) Parse() (PNode, error) {
	r := yyParse(q)
	if r != 0 {
		return nil, fmt.Errorf("Can't parse %s: %s", q.input, q.errtxt)
	}
	return q.expr, nil
}

// Notes about Go regexes:
//
// - they do not have full lookahead/lookbehind.  Here \b serves to ensure that the keywords are not
//   part of a longer identifier.
// - they are greedy and the order of clauses matters, as does the order of disjuncts within the
//   punctuation case: longer matches must precede shorter, eg 5m must precede 5, <= must precede <.
//
// In general, this is a cute solution but it feels brittle.

var tokenRe = regexp.MustCompile(
	strings.Join([]string{
		`(\s+)`,
		`(<=|<|>=|>|=~|=|and\b|or\b|not\b|\(|\))`,
		`([a-zA-Z_][a-zA-Z0-9_]*)`,
		`"([^"]*)"`,
		`'([^']*)'`,
		`/([^/]*)/`,
		"`([^`]*)`",
		`(\d+[wW](?:\d+[dD])?(?:\d+[hH])?(?:\d+[mM])?)`,
		`(\d+[dD](?:\d+[hH])?(?:\d+[mM])?)`,
		`(\d+[hH](?:\d+[mM])?)`,
		`(\d+[mM])`,
		`(-?\d+(?:\.\d+)?(?:[eE][-+]?\d+)?)`,
		`(.)`,
	}, "|"))

const (
	spaces      = 1
	punctuation = 2
	ident       = 3
	firstString = 4
	lastString  = 12
	bad         = 13
)

var punct = map[string]int{
	"<":   tLt,
	"<=":  tLe,
	">":   tGt,
	">=":  tGe,
	"=":   tEq,
	"=~":  tMatch,
	"and": tAnd,
	"or":  tOr,
	"not": tNot,
	"(":   tLparen,
	")":   tRparen,
}

func newQueryParser(input string) (*queryParser, error) {
	m := tokenRe.FindAllStringSubmatch(input, -1)
	if m == nil {
		// This shouldn't actually happen: the regex should match every possible string.
		return nil, fmt.Errorf("Can't lex %s", input)
	}
	tokens := make([]token, 0)
	for _, tm := range m {
		var t int
		var text string
		switch {
		case tm[spaces] != "":
			continue
		case tm[ident] != "":
			text = tm[ident]
			t = tIdent
		case tm[punctuation] != "":
			text = tm[punctuation]
			t = punct[text]
		case tm[bad] != "":
			return nil, fmt.Errorf("Bad character: %s", tm[bad])
		default:
			for i := firstString ; i <= lastString ; i++ {
				if tm[i] != "" {
					text = tm[i]
					t = tString
					break
				}
			}
			if t == 0 {
				panic("Bad match")
			}
		}
		tokens = append(tokens, token{t, text})
	}
	return &queryParser{
		input: input,
		tokens: tokens,
	}, nil
}
