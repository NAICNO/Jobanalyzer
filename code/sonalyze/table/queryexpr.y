%{
//go:generate goyacc -o queryexpr.go queryexpr.y

package table

import (
	"fmt"
	"regexp"
)

%}

%union {
    text string
    node PNode
}

%start TopExpr

%token <text> TIdent TString
%left <text> TOr
%left <text> TAnd
%left <text> TEqual TLess TLessOrEqual TGreater TGreaterOrEqual TMatch
%right <text> TNot
%token TLeftParen TRightParen

%type <node> Expr TopExpr
%type <text> String

%%

TopExpr : Expr { yylex.(*queryParser).expr = $$ } ;

Expr : TNot Expr { $$ = NewUnop(POpNot, $2) }
     | Expr TOr Expr { $$ = NewLogical(POpOr, $1, $3) }
     | Expr TAnd Expr  { $$ = NewLogical(POpAnd, $1, $3) }
     | TIdent TEqual String { $$ = NewBinop(POpEq, $1, $3) }
     | TIdent TLess String { $$ = NewBinop(POpLt, $1, $3) }
     | TIdent TLessOrEqual String { $$ = NewBinop(POpLe, $1, $3) }
     | TIdent TGreater String { $$ = NewBinop(POpGt, $1, $3) }
     | TIdent TGreaterOrEqual String { $$ = NewBinop(POpGe, $1, $3) }
     | TIdent TMatch String { $$ = NewBinop(POpMatch, $1, $3) }
     | TLeftParen Expr TRightParen { $$ = $2 }
     ;

// Keywords and idents can appear in the string position, to reduce the need for quoting
String : TString { $$ = $1 }
       | TIdent { $$ = $1 }
       | TOr { $$ = $1 }
       | TAnd { $$ = $1 }
       | TNot { $$ = $1 }
       ;

%%

type token struct {
	tok  int
	text string
}

type queryParser struct {
	input  string
	tokens []token
	errtxt string
	expr   PNode				// final result of successful parse
}

func (q *queryParser) Lex(lval *yySymType) (tok int) {
	if len(q.tokens) == 0 {
		tok = -1
	} else {
		tok = q.tokens[0].tok
		lval.text = q.tokens[0].text
		fmt.Printf("Consumed %d %s\n", tok, lval.text)
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

// This lexer is noncontextual, so a bit more quoting is needed than strictly desirable.  We should add
// some negative-lookahead assertions to the punctuation, esp to `and` and `or`, but this does not
// completely fix anything.

var tokenRe = regexp.MustCompile(`(\s+)|(<|<=|>|>=|=|and|or|not|\(|\))|([a-zA-Z][a-zA-Z0-9]*)|"([^"]*)"|'([^']*)'|/([^/]*)/|([^()\s]+)`)

const (
	spaces = 1
	punctuation = 2
	ident = 3
	dquoted = 4
	squoted = 5
	slashed = 6
	nonspace = 7
)

var punct = map[string]int{
	"<": TLess,
	"<=": TLessOrEqual,
	">": TGreater,
	">=": TGreaterOrEqual,
	"=": TEqual,
	"and": TAnd,
	"or": TOr,
	"not": TNot,
	"(": TLeftParen,
	")": TRightParen,
}

func newQueryParser(input string) (*queryParser, error) {
	m := tokenRe.FindAllStringSubmatch(input, -1)
	if m == nil {
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
			t = TIdent
		case tm[punctuation] != "":
			text = tm[punctuation]
			t = punct[text]
		case tm[dquoted] != "":
			text = tm[dquoted]
			t = TString
		case tm[squoted] != "":
			text = tm[squoted]
			t = TString
		case tm[slashed] != "":
			text = tm[slashed]
			t = TString
		case tm[nonspace] != "":
			text = tm[nonspace]
			t = TString
		default:
			panic("Bad match")
		}
		tokens = append(tokens, token{t, text})
	}
	return &queryParser{
		input: input,
		tokens: tokens,
	}, nil
}

