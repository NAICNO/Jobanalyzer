// Query compiler
//
// A query is an expression
//
//   e ::= id op val | e and e | e or e | !e | ( e )
//
// where val is a string, possibly quoted, which will be converted to a value appropriate for the id
// and/or the operator, eg, 1h5m is converted to 3605 for a "duration" field, 2024-10-21 to some
// timestamp for a date field, etc.
//
// where op is an operator < <= > >= != =~ 
//
// where =~ is regexp match
// where =? is string prefix match, same as =~ ^x

package table

type QOp int
const (
	QEquals QOp = iota
	QLess
	QGreater
)

// clearly the `d` type can be fixed...
// but what about v?  It's field-specific, so probably not.

type QFunc = func(d any, op QOp, v any) bool
type QCvt = func(v any) any

type QField struct {
	Convert Qcvt
	Filter QFunc
}

// To make this even faster there could be one function per operation: FilterEq, FilterLt, FilterGt.
// But there might be many operators, so it doesn't scale.  Not all would apply to all types though.
// But for number types, many enough.  And presumably we've just moved the switch out somewhere,
// where it has to figure out which function to call?  (But can maybe do that just once.)

var nodesQuery = map[string]QField{
	"MemGB": {
		Convert: func(v any) any {
			return ToInteger(v)
		},
		Filter: func(d any, op QOp, v any) bool {
			fv := d.(*config.NodeConfigRecord).MemGB
			fw := v.(int)
			switch op {
			case QEqual:
				return fv == fw
			case QLess:
				return fv < fw
			case QGreater:
				return fv > fw
			default:
				panic("NYI")
			}
		},
		FilterEq: func(d any, v any) bool {
			// Getting rid of one case is going to help here...
			return d.(*config.NodeConfigRecord).MemGB == v.(int)
		},
	},
}

func filterGen[T any](f *ast, filters map[string]QField) func(d T, v any) bool {
	switch f.op {
	case QEqual:
		filter
