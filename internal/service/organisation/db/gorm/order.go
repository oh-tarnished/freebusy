package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
)

var orgSortColumns = map[string]string{
	"name":         "name",
	"display_name": "display_name",
	"slug":         "slug",
	"create_time":  "create_time",
	"update_time":  "update_time",
}

var memberSortColumns = map[string]string{
	"email":       "email",
	"role":        "role",
	"state":       "state",
	"create_time": "create_time",
	"update_time": "update_time",
}

func orgOrderClause(orderBy string) (string, error) { return orderClause(orderBy, orgSortColumns) }
func memberOrderClause(orderBy string) (string, error) {
	return orderClause(orderBy, memberSortColumns)
}

// orderClause turns an AIP-132 order_by string into a safe "col DIR, ..." clause
// using the given column allowlist. Unknown fields are rejected.
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
