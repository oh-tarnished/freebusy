package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
)

// userSortColumns is the allowlist of order_by fields ListUsers accepts, mapped
// to safe physical column names.
var userSortColumns = map[string]string{
	"display_name": "display_name",
	"email":        "email",
	"create_time":  "create_time",
	"update_time":  "update_time",
}

func userOrderClause(orderBy string) (string, error) {
	terms, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		col, ok := userSortColumns[term.Field]
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
