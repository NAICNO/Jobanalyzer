// -*- mode: go -*-

%{
//go:generate goyacc -o parser.go parser.y

package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Generate-table syntax tree.  Names are never "" unless noted.

type TableBlock struct {
	TableName string
	Prefix    []string			// completely unedited lines
	Fields    FieldSect
	Generate  string			// "" if no such section
	Summary   *HelpSect			// nil if no such section
	Help      *HelpSect			// nil if no such section
	Aliases   []Alias		    // empty if no such section
	Defaults  []string			// empty if no such section
}

type FieldSect struct {
	Type   string
	Fields []Field				// empty if no fields
}

type Field struct {
	Name  string
	Type  string
	Attrs []NV				    // quotes removed but escape codes intact in values
}

type HelpSect struct {
	Command string				// "" if no command name was present
	Text    []string			// including leading and trailing blank lines
}

type Alias struct {
	Name string
	Fields []string				// never empty
}

type NV struct {
	Name  string
	Value string
}

// Parser interface

func NewParser(inputName string, input io.Reader) *Parser {
	return &Parser{
		inputName: inputName,
		input: bufio.NewScanner(input),
	}
}

// Returns nil, nil at EOF

func (parser *Parser) Parse() (*TableBlock, error) {
	parser.inTable = false
	parser.doTokenize = false
	parser.errtxt = ""
	parser.qfront = nil
	parser.qtail = nil
	parser.table = nil
	r := yyParse(parser)
	if r != 0 {
		return nil, errors.New(parser.errtxt)
	}
	return parser.table, nil
}

%}

%union {
    s           string
	ss          []string
	nv          NV
	nvs         []NV
	as          []Alias
    fieldSect   FieldSect
	fields      []Field
	field       Field
	helpSect    *HelpSect
	summarySect *HelpSect
}

%start Table

%token <s> tIdent tText tString
%token tTable tFields tGenerate tHelp tSummary tAliases tDefaults tElbat
%token tStar tColon tComma tDot tEol tMark tLbrace tRbrace
%token tBad

%type <s> TypeName TextLine IdentOpt Table IdentPath GenerateOpt
%type <ss> TextLines Prefix IdentList DefaultsOpt
%type <nvs> Attributes
%type <as> AliasesOpt AliasLines
%type <nv> Attribute
%type <fieldSect> Fields
%type <fields> FieldLines
%type <field> FieldLine
%type <helpSect> HelpOpt SummaryOpt

%%

Table : { yylex.(*Parser).table = nil }
      | tTable tIdent tEol
        Prefix
		Mark
        Fields
        GenerateOpt
        SummaryOpt
        HelpOpt
        AliasesOpt
        DefaultsOpt
		tElbat
        {
			yylex.(*Parser).table = &TableBlock{
		 		TableName: $2,
				Prefix:    $4,
				Fields:    $6,
				Generate:  $7,
				Summary:   $8,
				Help:      $9,
				Aliases:   $10,
				Defaults:  $11,
			}
		}
		;

Prefix     : TextLines ;

Mark       : tMark { yylex.(*Parser).doTokenize = true } tEol ;

Fields     : tFields { yylex.(*Parser).doTokenize = true } TypeName tEol
             FieldLines { $$ = FieldSect{$3, $5} }
           ;
FieldLines : { $$ = make([]Field, 0) }
           | FieldLines FieldLine {	$$ = append($1, $2) }
           ;
FieldLine  : tIdent TypeName Attributes tEol { $$ = Field{ $1, $2, $3 } } ;
Attributes : { $$ = make([]NV, 0) }
           | Attributes Attribute { $$ = append($1, $2) }
           ;
Attribute  : tIdent tColon tString { $$ = NV{$1, $3 } } ;

GenerateOpt : { $$ = "" }
            | tGenerate { yylex.(*Parser).doTokenize = true } tIdent tEol { $$ = $3 }
            ;

HelpOpt     : { $$ = nil }
            | tHelp { yylex.(*Parser).doTokenize = false } IdentOpt tEol
              TextLines { $$ = &HelpSect{$3, $5} }
            ;

SummaryOpt  : { $$ = nil }
            | tSummary { yylex.(*Parser).doTokenize = false } tIdent tEol
              TextLines { $$ = &HelpSect{$3, $5} }
            ;

AliasesOpt  : { $$ = make([]Alias, 0) }
            | tAliases { yylex.(*Parser).doTokenize = true } tEol
              AliasLines { $$ = $4 }
            ;
AliasLines  : { $$ = make([]Alias, 0) }
            | AliasLines tIdent IdentList tEol { $$ = append($1, Alias{$2, $3}) }
            ;

DefaultsOpt : { $$ = nil }
            | tDefaults { yylex.(*Parser).doTokenize = true } IdentList tEol { $$ = $3 }
            ;

TypeName   : tStar TypeName { $$ = "*" + $2 }
           | tLbrace tRbrace TypeName { $$ = "[]" + $3 }
           | IdentPath
           ;

TextLines  : { $$ = make([]string, 0) }
           | TextLines TextLine { $$ = append($1, $2) }
           ;
TextLine   : tText tEol ;

IdentOpt  : { $$ = "" }
          | tIdent
          ;
IdentList : tIdent { $$ = []string{$1} }
          | IdentList tComma tIdent { $$ = append($1, $3) }
          ;
IdentPath : tIdent
          | tIdent tDot IdentPath { $$ = $1 + "." + $3 }
          ;

%%

type token struct {
	tok  int
	text string
	next *token
}

type Parser struct /* implements yyLexer */ {
	input      *bufio.Scanner
	inputName  string
	lineno     int
	currline   int
	inTable    bool
	doTokenize bool
	errtxt     string
	qfront     *token
	qtail      *token
	table      *TableBlock
}

func (p *Parser) getLine() (string, bool) {
	p.currline = p.lineno
	if !p.input.Scan() {
		return "", false
	}
	s := p.input.Text()
	p.lineno++
	for {
		before, found := strings.CutSuffix(s, "\\")
		if !found {
			break
		}
		s = before
		if !p.input.Scan() {
			break
		}
		s += p.input.Text()
		p.lineno++
	}
	return s, true
}

func (p *Parser) enqueue(t int, s string) {
	n := &token{t, s, nil}
	if p.qfront == nil {
		p.qfront = n
	} else {
		p.qtail.next = n
	}
	p.qtail = n
}

func (p *Parser) dequeue() (int, string, bool) {
	if p.qfront == nil {
		return 0, "", false
	}
	n := p.qfront
	p.qfront = n.next;
	if p.qfront == nil {
		p.qtail = nil
	}
	return n.tok, n.text, true
}

var (
	tablePrefix        = regexp.MustCompile(`^/\*TABLE(?:\s+(.*))?$`)
	blankOrCommentLine = regexp.MustCompile(`^\s*(?:#.*)?$`)
	sectionheader      = regexp.MustCompile(`^(ELBAT\*/|%%|FIELDS|GENERATE|SUMMARY|HELP|ALIASES|DEFAULTS)(?:\s+(.*))?$`)
)

var keyword = map[string]int{
	"ELBAT*/": tElbat,
	"%%": tMark,
	"FIELDS": tFields,
	"GENERATE": tGenerate,
	"SUMMARY": tSummary,
	"HELP": tHelp,
	"ALIASES": tAliases,
	"DEFAULTS": tDefaults,
}

func (p *Parser) Lex(lval *yySymType) (tok int) {
	if !p.inTable {
		for {
			s, haveLine := p.getLine()
			if !haveLine {
				return -1
			}
			if m := tablePrefix.FindStringSubmatch(s); m != nil {
				p.inTable = true
				p.qfront = nil
				p.qtail = nil
				p.tokenize(m[1])
				return tTable
			}
		}
	}
	if tok, text, ok := p.dequeue(); ok {
		lval.s = text
		return tok
	}
Scanner:
	for {
		s, haveLine := p.getLine()
		if !haveLine {
			p.inTable = false
			return -1
		}
		if m := sectionheader.FindStringSubmatch(s); m != nil {
			t := keyword[m[1]]
			switch t {
			case 0:
				t = tBad
			case tElbat:
				p.inTable = false
			case tMark:
				p.enqueue(tEol, "")
			default:
				p.tokenize(m[2])
			}
			return t
		}
		if !p.doTokenize {
			lval.s = s
			p.enqueue(tEol, "")
			return tText
		}
		if blankOrCommentLine.MatchString(s) {
			continue Scanner
		}
		p.tokenize(s)
		tok, text, _ := p.dequeue()
		lval.s = text
		return tok
	}
}

// tokenize() always pushes at least eol

func (p *Parser) tokenize(s string) {
	r := strings.NewReader(s)
Scanner:
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			break Scanner
		}
		switch c {
		case ' ', '\t', '\n', '\r':
			continue Scanner
		case '.':
			p.enqueue(tDot, "")
		case ':':
			p.enqueue(tColon, "")
		case ',':
			p.enqueue(tComma, "")
		case '*':
			p.enqueue(tStar, "")
		case '[':
			p.enqueue(tLbrace, "")
		case ']':
			p.enqueue(tRbrace, "")
		case '"':
			s := ""
		String:
			for {
				c, _, err := r.ReadRune()
				if err != nil {
					p.enqueue(tBad, "End of input in string")
					break Scanner
				}
				if c == '"' {
					break String
				}
				if c == '\\' {
					c, _, err = r.ReadRune()
					if err != nil {
						p.enqueue(tBad, "End of input in string")
						break Scanner
					}
				}
				s = s + string(c)
			}
			p.enqueue(tString, s)
		default:
			if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
				s := string(c)
			Ident:
				for {
					c, _, err := r.ReadRune()
					if err != nil {
						break Ident
					}
					// "Not whitespace and not comma" would also be an acceptable guard here, but
					// this covers the chars traditionally used in idents.
					if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
					   c == '/' || c == '_' || c == '-' || c == '%' {
						s += string(c)
					} else {
						r.UnreadRune()
						break Ident
					}
				}
				p.enqueue(tIdent, s)
			} else {
				p.Error(fmt.Sprintf("Unknown character %c", c))
				p.enqueue(tBad, "")
				break Scanner
			}
		}
	}
	p.enqueue(tEol, "")
}

func (p *Parser) Error(s string) {
	if p.errtxt == "" {
		p.errtxt = fmt.Sprintf("%s:%d: %s", p.inputName, p.currline, s)
	}
}
