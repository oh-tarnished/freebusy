package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
)

// exceptionSortColumns is the allowlist of order_by fields ListAvailabilityExceptions
// accepts, mapped to safe physical column names.
var exceptionSortColumns = map[string]string{
	"kind":        "kind",
	"create_time": "create_time",
}

func exceptionOrderClause(orderBy string) (string, error) {
	terms, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		col, ok := exceptionSortColumns[term.Field]
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
