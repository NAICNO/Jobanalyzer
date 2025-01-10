// Simple query compiler.

package table

import (
	"fmt"
	"regexp"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Predicate table.
//
// The table generator will generate a table of converters and predicates for every field.
//
// The Convert function, if not nil, converts a string value supplied as part of the query text to a
// value of the appropriate type, but represented as an `any`.  The value will be converted once,
// during compilation.
//
// The Compare predicate takes a table row of the appropriate type, and the converted value, and
// returns -1, 0, or 1 depending on the value of the field in relation to the argument value.
//
// The SetCompare predicate takes a table row of the appropriate type, and the converted value, and
// an operation from the set =, <, <=, >, >=, encoded as 1, 2, 3, 4, 5 respectively, and returns
// true if the field and the value have the corresponding set relation (equal, proper subset,
// improper subset, proper superset, improper superset).

type Predicate[T any] struct {
	Convert    func(d string) (any, error)
	Compare    func(d T, v any) int
	SetCompare func(d T, v any, op int) bool
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Syntax trees.
//
// Parsed queries are represented as PNode instances, all of them are tagged with a POpXx.

const (
	// The value 0 is never a valid opcode.
	opEq = 1 + iota
	opLt
	opLe
	opGt
	opGe
	opMatch
	opAnd
	opOr
	opNot
)

var pop2op = [...]string{
	"*BAD*",
	"=",
	"<",
	"<=",
	">",
	">=",
	"=~",
	"and",
	"or",
	"not",
}

type PNode fmt.Stringer

type unaryOp struct {
	op  int
	opd PNode
}

func (b *unaryOp) String() string {
	return fmt.Sprintf("(%s %s)", pop2op[b.op], b.opd)
}

type logicalOp struct {
	op       int
	lhs, rhs PNode
}

func (b *logicalOp) String() string {
	return fmt.Sprintf("(%s %s %s)", pop2op[b.op], b.lhs, b.rhs)
}

type binaryOp struct {
	op           int
	field, value string
}

func (b *binaryOp) String() string {
	return fmt.Sprintf("(%s %s %s)", pop2op[b.op], b.field, b.value)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Query parsing

func ParseQuery(input string) (PNode, error) {
	parser, err := newQueryParser(input)
	if err != nil {
		return nil, err
	}
	return parser.Parse()
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Query generator.
//
// The compilers take both predicates and formatters because the =~ operator needs to format the lhs
// to be able to match it against the rhs.

// The returned predicate returns true iff the test passed.

func CompileQuery[T any](
	formatters map[string]Formatter[T],
	predicates map[string]Predicate[T],
	q PNode,
) (func(d T) bool, error) {
	return compileQuery(formatters, predicates, q)
}

// The returned predicate returns false iff the test passed.
//
// TODO: It's possible to avoid wrapping the predicate here.

func CompileQueryNeg[T any](
	formatters map[string]Formatter[T],
	predicates map[string]Predicate[T],
	q PNode,
) (func(d T) bool, error) {
	query, err := compileQuery(formatters, predicates, q)
	if err != nil {
		return nil, err
	}
	return func(d T) bool { return !query(d) }, nil
}

func compileQuery[T any](
	formatters map[string]Formatter[T],
	predicates map[string]Predicate[T],
	q PNode,
) (func(d T) bool, error) {
	switch l := q.(type) {
	case *logicalOp:
		lhs, err := compileQuery(formatters, predicates, l.lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := compileQuery(formatters, predicates, l.rhs)
		if err != nil {
			return nil, err
		}
		switch l.op {
		case opAnd:
			return func(d T) bool { return lhs(d) && rhs(d) }, nil
		case opOr:
			return func(d T) bool { return lhs(d) || rhs(d) }, nil
		default:
			panic("Unknown op")
		}
	case *unaryOp:
		opd, err := compileQuery(formatters, predicates, l.opd)
		if err != nil {
			return nil, err
		}
		switch l.op {
		case opNot:
			return func(d T) bool { return !opd(d) }, nil
		default:
			panic("Unknown op")
		}
	case *binaryOp:
		if l.op == opMatch {
			format, found := formatters[l.field]
			if !found {
				return nil, fmt.Errorf("Field not found: %s", l.field)
			}
			re, err := regexp.Compile(l.value)
			if err != nil {
				return nil, err
			}
			formatter := format.Fmt
			return func(d T) bool { return re.MatchString(formatter(d, 0)) }, nil
		}

		p, found := predicates[l.field]
		if !found {
			return nil, fmt.Errorf("Field not found: %s", l.field)
		}
		var value any
		if p.Convert != nil {
			v, err := p.Convert(l.value)
			if err != nil {
				return nil, err
			}
			value = v
		} else {
			value = l.value
		}
		if p.SetCompare != nil {
			setCompare := p.SetCompare
			op := l.op
			switch op {
			case opEq, opLt, opLe, opGt, opGe:
				return func(d T) bool { return setCompare(d, value, op) }, nil
			default:
				panic("Unknown op")
			}
		}
		compare := p.Compare
		switch l.op {
		case opEq:
			return func(d T) bool { return compare(d, value) == 0 }, nil
		case opLt:
			return func(d T) bool { return compare(d, value) < 0 }, nil
		case opLe:
			return func(d T) bool { return compare(d, value) <= 0 }, nil
		case opGt:
			return func(d T) bool { return compare(d, value) > 0 }, nil
		case opGe:
			return func(d T) bool { return compare(d, value) >= 0 }, nil
		default:
			panic("Unknown op")
		}
	default:
		panic("Bad operator type")
	}
}
