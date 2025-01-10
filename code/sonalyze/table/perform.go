// Query abstraction

// The query runs on the generated table rows.  The *names* are always relative to the printed table
// rows.  And the *input* to the query predicate is such a table row.  So the query can't be applied
// until those rows have been computed.  This makes most sense from the user's point of view also.
// But it may limit applicability.

package table

import (
	"fmt"
	"slices"
)

func ApplyQuery[T any](
	q PNode,
	formatters map[string]Formatter[T],
	predicates map[string]Predicate[T],
	records []T,
) ([]T, error) {
	if q != nil {
		queryNeg, err := CompileQueryNeg(formatters, predicates, q)
		if err != nil {
			return nil, fmt.Errorf("Could not compile query: %v", err)
		}
		records = slices.DeleteFunc(records, queryNeg)
	}
	return records, nil
}
