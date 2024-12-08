// -*- mode: go -*-
//
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
// Logical ops "and", "or", and "not" combine other expressions; parens override precedence.
//
// TODO: It would have been nice to have an even more permissive string syntax for user convenience.
// This can maybe be done if the parser is made to feed back to the lexer to allow a more permissive
// lexer to lex strings.  Another possibility is to expand the character set allowed by identifiers.

%{
//go:generate goyacc -o queryexpr.go queryexpr.y

package table

import (
	"fmt"
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

type queryParser struct {
	errtxt string
	expr   PNode
	// lexer
	input  []byte
	i      int
}

func (q *queryParser) Error(s string) {
	if q.errtxt == "" {
		q.errtxt = s
	}
}

func (q *queryParser) Parse() (PNode, error) {
	r := yyParse(q)
	// Since the lexer returns -1 on syntax errors, the parser may find itself in a valid state and
	// may not return an error indicator.  But errtxt != "" will still indicate the lexer error.
	if r != 0 || q.errtxt != "" {
		return nil, fmt.Errorf("Can't parse %s: %s", q.input, q.errtxt)
	}
	return q.expr, nil
}

func newQueryParser(input string) (*queryParser, error) {
	return &queryParser{
		input: []byte(input),
		i:     0,
	}, nil
}
