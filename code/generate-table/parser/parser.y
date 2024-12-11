// -*- mode: go -*-
//
// TODO: Recognize more keywords
// TODO: AliasesOpt and mAliases
// TODO: Joining lines / line stream
// TODO: Leaky token queue

%{
//go:generate goyacc -o parser.go parser.y

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// Syntax tree

type TableBlock struct {
	TableName string
	Prefix    []string
	Fields    FieldSect
	Generate  *string
	Help      *HelpSect
	Aliases   []Pair
	Defaults  *string
}

type FieldSect struct {
	Type   string
	Fields []Field
}

type Field struct {
	Name  string
	Type  string
	Attrs []Pair
}

type HelpSect struct {
	Command string
	Text    []string
}

type Pair struct {
	Name  string
	Value string
}

%}

%union {
    s         string
	ps        *string
	ss        []string
	pair      Pair
	pairs     []Pair
    fieldSect FieldSect
	fields    []Field
	field     Field
	helpSect  *HelpSect
}

%start Table

%token <s> tIdent tSectAttr tText tAttrName tAttrText
%token tTable tEol tMark tFields tGenerate tHelp tAliases tDefaults tElbat tBad

%type <s> TextLine Table CommandName
%type <ss> TextLines
%type <pairs> Attributes AliasesOpt
%type <pair> Attribute
%type <fieldSect> Fields
%type <fields> FieldLines
%type <field> FieldLine
%type <ps> GenerateOpt DefaultsOpt
%type <helpSect> HelpOpt

%%

Table : tTable tIdent tEol
        { yylex.(*tableParser).mode=mText }
        TextLines
		tMark tEol
        { yylex.(*tableParser).mode=mFields }
        Fields
        { yylex.(*tableParser).mode=mText }
        GenerateOpt
        HelpOpt
        { yylex.(*tableParser).mode=mAliases }
        AliasesOpt
        { yylex.(*tableParser).mode=mText }
        DefaultsOpt
		tElbat
        {
			yylex.(*tableParser).table = &TableBlock{
		 		TableName: $2,
				Prefix:    $5,
				Fields:    $9,
			}
		}
		;

TextLines  : { $$ = make([]string, 0) }
           | TextLines TextLine { $$ = append($1, $2) }
           ;
TextLine   : tText tEol { $$ = $1 } ;

Fields     : tFields tSectAttr tEol FieldLines { $$ = FieldSect{$2, $4} } ;
FieldLines : { $$ = make([]Field, 0) }
           | FieldLines FieldLine {	$$ = append($1, $2) }
           ;
FieldLine  : tIdent tIdent Attributes tEol { $$ = Field{ $1, $2, $3 } } ;
Attributes : { $$ = make([]Pair, 0) }
           | Attributes Attribute { $$ = append($1, $2) }
           ;
Attribute  : tAttrName tAttrText { $$ = Pair{$1, $2 } } ;

GenerateOpt : { $$ = nil }
            | tGenerate tSectAttr tEol { $$ = strdup($2) }
            ;

HelpOpt     : { $$ = nil }
            | tHelp CommandName tEol TextLines { $$ = &HelpSect{$2, $4} }
            ;
CommandName : { $$ = "" }
            | tIdent { $$ = $1 }
            ;

DefaultsOpt : { $$ = nil }
            | tDefaults tSectAttr tEol { $$ = strdup($2) }
            ;

%%

const (
	mText = iota
	mFields
	mAliases
)

type token struct {
	tok int
	text string
}

type tableParser struct {
	input   *bufio.Scanner
	inTable bool
	mode    int
	errtxt  string
	queue   []token // this will leak, fix later
	table   *tableT
}

func Parse(input io.Reader) (*TableBlock, error) {
	p := tableParser{
		input: bufio.NewScanner(input),
		queue: make([]token, 0),
	}
	r := yyParse(&p)
	if r != 0 {
		return nil, fmt.Errorf("Parse error: %s", p.errtxt)
	}
	return p.table, nil
}

func (p *tableParser) enqueue(t token) {
	p.queue = append(p.queue, t)
}

func (p *tableParser) dequeue() (token, bool) {
	if len(p.queue) == 0 {
		return token{}, false
	}
	t := p.queue[0]
	p.queue = p.queue[1:]
	return t, true
}

func (p *tableParser) Error(s string) {
	if p.errtxt == "" {
		p.errtxt = s
	}
}

var (
	tablePrefix     = regexp.MustCompile(`^/\*TABLE\s+(\S+)\s*$`)
	tableSuffix     = regexp.MustCompile(`^ELBAT\*/`)
	prefixEndMarker = regexp.MustCompile(`^%%\s*$`)
	blankOrCommentLine = regexp.MustCompile(`^\s*(?:#.*)?$`)
	fieldsLine   = regexp.MustCompile(`^FIELDS\s+(\S+)\s*$`)
	fieldLine = regexp.MustCompile(`^\s*(\S+)\s+(\S+)(?:\s+(.*))?`)
	fieldAttr   = regexp.MustCompile(`^\s*([a-z]+):"([^"]*)"`)
)

func (p *tableParser) Lex(lval *yySymType) (tok int) {
	if !p.inTable {
		for {
			if !p.input.Scan() {
				return -1
			}
			s := p.input.Text()
			if m := tablePrefix.FindStringSubmatch(s); m != nil {
				p.inTable = true
				p.enqueue(token{tIdent, m[1]})
				p.enqueue(token{tEol, ""})
				return tTable
			}
		}
	}
	if t, ok := p.dequeue(); ok {
		lval.s = t.text
		return t.tok
	}
Scanner:
	for {
		if !p.input.Scan() {
			p.inTable = false
			return -1
		}
		s := p.input.Text()
		if tableSuffix.MatchString(s) {
			p.inTable = false
			return tElbat
		}
		if prefixEndMarker.MatchString(s) {
			p.enqueue(token{tEol, ""})
			return tMark
		}
		if blankOrCommentLine.MatchString(s) {
			continue Scanner
		}
		if m := fieldsLine.FindStringSubmatch(s); m != nil {
			p.enqueue(token{tSectAttr, m[1]})
			p.enqueue(token{tEol, ""})
			return tFields
		}
		// TODO: more of those
		switch p.mode {
		case mText:
			lval.s = s
			p.enqueue(token{tEol, ""})
			return tText
		case mFields:
			if m := fieldLine.FindStringSubmatch(s); m != nil {
				p.enqueue(token{tIdent, m[2]})
				attrs := p.parseAttrs(m[3])
				for k, v := range attrs {
					p.enqueue(token{tAttrName, k})
					p.enqueue(token{tAttrText, v})
				}
				p.enqueue(token{tEol, ""})
				lval.s = m[1]
				return tIdent
			}
			return tBad
		case mAliases:
			// TODO: Do this
		default:
			panic("Bad")
		}
	}
}

var validAttr = map[string]bool{
	"desc":     true,
	"alias":    true,
	"field":    true,
	"indirect": true,
	"config":   true,
}

func (p *tableParser) parseAttrs(s string) map[string]string {
	attrs := make(map[string]string)
	for s != "" {
		m := fieldAttr.FindStringSubmatch(s)
		if m == nil || !validAttr[m[1]] {
			p.Error("Invalid attribute staring at " + s)
			continue
		}
		attrs[m[1]] = m[2]
		s = s[len(m[0]):]
	}
	return attrs
}
