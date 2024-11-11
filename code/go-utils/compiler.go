// (This started as a clone of ../dashboard/queryterm.js.)
//
// It works by compiling a simple boolean/relational query under an environment into a query matcher
// object, that can then be applied to the data rows of a table.
//
// The query matcher takes as its input an immutable table of data objects (rows) and a
// representation of the set of rows (indices) from that set that are currently considered.  It
// returns a new set, the result of the filter, a subset of the input set.  (It is based on a set of
// rows rather than a single row because culling by eg node names in terms evaluated early will tend
// to quickly reduce the number of data rows considered by terms evaluated later, but that
// optimization is not currently implemented.)
//
// For example, the primitive query "(mem% > 50)" constructs a set that comprises those rows in the
// input set whose mem% column is greater than 50.  For another example, the primitive query "host =
// c1-*" constructs a set that comprises those elements in the input set whose node names match that
// pattern.  The query "host = c1-* and mem% > 50" combines them.
//
// Note that the actual defined variable terms ("mem%" etc) are defined by the client of this
// library.

// Query compiler.
//
// The `query` is the query expression.  The `knownFields` is a map from known field names to either
// itself or to a canonical field name (allowing for aliases).  The `builtinOperations` is a map
// from expression aliases (essentially subroutines) to their matcher expandings.
//
// Returns either a query matcher object, or a string describing the error as precisely as possible.
//
// The expression grammar is simple:
//
//   expr ::= token binop token
//          | unop token
//          | token
//          | (expr)
//
//   binop ::= "and" | "or" | "<" | ">" | "<=" | ">=" | "="
//   unop ::= "~"
//   token ::= string of non-whitespace chars other than punctuation, except "and" and "or"
//   punctuation ::= "<", ">", "=", "~", "(", ")"
//
// Everything is case-sensitive.
//
// Tokens represent either known field references, numbers, or node name wildcard matchers.  The
// interpretation of a token is contextual: a field name or number is interpreted as such only in
// the context of a binary operator that requires that interpretation; in all other contexts they
// are interpreted as node name matchers.  That is, "c1* > 5 and 37.5" is a legal and meaningful
// instruction if there is a known field called "c1*" and a node called "37.5".
//
// Binary operator precedence from low to high: OR, AND, relationals.  Unary ops bind tighter
// than binary ops.
//
// For a relational operator, the first argument must be a field name and the second must be a
// number.
//
// The meaning of an expression is that each node name matcher or relational operation induces a
// subset of the data rows, "and" is set intersection, "or" is set union, and "~" is set complement
// (where the full set of nodes is the universe).

import (
	"strings"
)

type Bitset struct {
}

type CompiledOp interface {
	fmt.Stringer
	Eval(...) *Bitset
}

type notOperation struct {
	e CompiledOp
}

func (n *notOperation) String() string {
	return fmt.Sprintf("(~ %s)", n.e.String())
}

func (n *notOperation) Eval(data, elems) *Bitset {
}

type compiler struct {
	input *strings.Reader
}

func CompileQuery(
	query string,
	knownFields map[string]string,
	builtinOperations map[string]CompiledQuery,
) (CompiledQuery, error) {
	var c = &compiler{
		input: strings.NewReader(query),
	}
	return c.expr()
}

func (c *compiler) exprPrim() (CompiledOp, error) {
	if c.eatToken("(") {
		e, err := c.exprOr()
		if err != nil {
			return nil, err
		}
		if !c.eatToken(")") {
			return nil, c.fail("Expected ')' here")
		}
		return e, nil
	}
	if c.eatToken("~") {
		e, err := c.exprPrim()
		if err != nil {
			return nil, err
		}
		return &notOperation{e}, nil
	}
	tok := c.get()
	if numval, err := strconv.ParseFloat(tok, 10, 64); err == nil && math.IsFinite(numval) {
		return &numVal{numval}
	}
	// More typed values here, notably for durations, dates 2024-10-31T12:00

	// Hm, this is dodgy - if we have a dead-end, tok will have been updated.  The JS
	// code has the same problem.
	for {
		probe, found := knownFields[tok]
		if !found {
			break
		}
		if probe == tok {
			return &lookupField{probe}
		}
		tok = probe				// alias
	}
	if probe, found := builtinOperations[tok]; found {
		return probe
	}
	

}
