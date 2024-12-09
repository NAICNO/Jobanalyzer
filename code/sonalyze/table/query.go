// Simple query compiler.

package table

import (
	"fmt"
)

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Predicate table.
//
// The table generator will generate a table of converters and predicates for every field.
//
// The converter converts a string value supplied as part of the query text to a value of the
// appropriate type, but represented as an `any`.  The value will be converted once, during
// compilation.
//
// The Compare predicate takes a table row of the appropriate type, and the converted value, and
// returns -1, 0, or 1 depending on the value of the field in relation to the argument value.
//
// (There will be more predicates.)

type Predicate[T any] struct {
	Convert func(d string) (any, error)
	Compare func(d T, v any) int
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Syntax trees.
//
// Parsed queries are represented as PNode instances, all of them are tagged with a POpXx.

const (
	// The value 0 is never a valid opcode
	POpEq    = iota+1
	POpLt
	POpLe
	POpGt
	POpGe
	POpMatch
	POpAnd
	POpOr
	POpNot
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
	op int
	opd PNode
}

func (b *unaryOp) String() string {
	return fmt.Sprintf("(%s %s)", pop2op[b.op], b.opd)
}

type logicalOp struct {
	op int
	lhs, rhs PNode
}

func (b *logicalOp) String() string {
	return fmt.Sprintf("(%s %s %s)", pop2op[b.op], b.lhs, b.rhs)
}

type binaryOp struct {
	op int
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

func CompileQuery[T any](predicates map[string]Predicate[T], q PNode) (func(d T) bool, error) {
	switch l := q.(type) {
	case *logicalOp:
		lhs, err := CompileQuery(predicates, l.lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := CompileQuery(predicates, l.rhs)
		if err != nil {
			return nil, err
		}
		switch l.op {
		case POpAnd:
			return func(d T) bool { return lhs(d) && rhs(d) }, nil
		case POpOr:
			return func(d T) bool { return lhs(d) || rhs(d) }, nil
		default:
			panic("Unknown op")
		}
	case *unaryOp:
		opd, err := CompileQuery(predicates, l.opd)
		if err != nil {
			return nil, err
		}
		switch l.op {
		case POpNot:
			return func(d T) bool { return !opd(d) }, nil
		default:
			panic("Unknown op")
		}
	case *binaryOp:
		// TODO: POpMatch, this has built-in rhs conversion and probably calls a field accessor
		// instead of a comparator.
		//
		// TODO: Misc stuff for set-like things (GpuSet, maybe host name sets).
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
		}
		compare := p.Compare
		switch l.op {
		case POpEq:
			return func(d T) bool { return compare(d, value) == 0 }, nil
		case POpLt:
			return func(d T) bool { return compare(d, value) < 0 }, nil
		case POpLe:
			return func(d T) bool { return compare(d, value) <= 0 }, nil
		case POpGt:
			return func(d T) bool { return compare(d, value) > 0 }, nil
		case POpGe:
			return func(d T) bool { return compare(d, value) >= 0 }, nil
		default:
			panic("Unknown op")
		}
	default:
		panic("Bad operator type")
	}
}
