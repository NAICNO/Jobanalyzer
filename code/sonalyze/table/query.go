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
	PKindShift = 20

	PKindBinop   = 1
	PKindLogical = 2
	PKindUnop    = 3
)

const (
	// The value 0 is never a valid opcode
	POpEq    = 1 | (PKindBinop << PKindShift)
	POpLt    = 2 | (PKindBinop << PKindShift)
	POpLe    = 3 | (PKindBinop << PKindShift)
	POpGt    = 4 | (PKindBinop << PKindShift)
	POpGe    = 5 | (PKindBinop << PKindShift)
	POpMatch = 6 | (PKindBinop << PKindShift)

	POpAnd = 20 | (PKindLogical << PKindShift)
	POpOr  = 21 | (PKindLogical << PKindShift)

	POpNot = 31 | (PKindUnop << PKindShift)
)

var pop2op = map[int]string{
	POpEq:    "=",
	POpLt:    "<",
	POpLe:    "<=",
	POpGt:    ">",
	POpGe:    ">=",
	POpMatch: "=~",
	POpAnd:   "and",
	POpOr:    "or",
	POpNot:   "not",
}

type PNode interface {
	Op() int
	String() string
}

type opField struct {
	op int
}

func (q *opField) Op() int {
	return q.op
}

type unaryOp struct {
	opField
	opd PNode
}

func (b *unaryOp) String() string {
	return fmt.Sprintf("(%s %s)", pop2op[b.op], b.opd)
}

func NewUnop(pop int, opd PNode) PNode {
	return &unaryOp{
		opField: opField{pop},
		opd:     opd,
	}
}

type logicalOp struct {
	opField
	lhs, rhs PNode
}

func (b *logicalOp) String() string {
	return fmt.Sprintf("(%s %s %s)", pop2op[b.op], b.lhs, b.rhs)
}

func NewLogical(pop int, lhs, rhs PNode) PNode {
	return &logicalOp{
		opField: opField{op: pop},
		lhs:     lhs,
		rhs:     rhs,
	}
}

type binaryOp struct {
	opField
	field, value string
}

func (b *binaryOp) String() string {
	return fmt.Sprintf("(%s %s %s)", pop2op[b.op], b.field, b.value)
}

func NewBinop(pop int, field, value string) PNode {
	return &binaryOp{
		opField: opField{pop},
		field:   field,
		value:   value,
	}
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
	switch q.Op() >> PKindShift {
	case PKindLogical:
		l := q.(*logicalOp)
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
	case PKindUnop:
		l := q.(*unaryOp)
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
	case PKindBinop:
		l := q.(*binaryOp)
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
