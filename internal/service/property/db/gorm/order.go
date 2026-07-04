package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
)

// propertySortColumns is the allowlist of order_by fields ListProperties accepts,
// mapped to safe physical column names. Restricting to this closed set is what
// prevents a user-supplied order_by from injecting SQL into the ORDER BY clause.
var propertySortColumns = map[string]string{
	"name":         "name",
	"display_name": "display_name",
	"state":        "state",
	"create_time":  "create_time",
	"update_time":  "update_time",
}

// unitSortColumns is the equivalent allowlist for ListUnits.
var unitSortColumns = map[string]string{
	"name":         "name",
	"display_name": "display_name",
	"type":         "type",
	"state":        "state",
	"create_time":  "create_time",
	"update_time":  "update_time",
}

func propertyOrderClause(orderBy string) (string, error) {
	return orderClause(orderBy, propertySortColumns)
}

func unitOrderClause(orderBy string) (string, error) { return orderClause(orderBy, unitSortColumns) }

// orderClause turns an AIP-132 order_by string into a safe "col DIR, ..." clause
// using the given column allowlist. Unknown fields are rejected with
// types.ErrInvalidArgument; only allowlisted columns and ASC/DESC reach the SQL.
func orderClause(orderBy string, cols map[string]string) (string, error) {
	terms, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		col, ok := cols[term.Field]
		if !ok {
			return "", fmt.Errorf("%w: cannot sort by %q", types.ErrInvalidArgument, term.Field)
		}
		dir := "ASC"
		if term.Desc {
			dir = "DESC"
		}
		parts = append(parts, col+" "+dir)
	}
	return strings.Join(parts, ", "), nil
}
