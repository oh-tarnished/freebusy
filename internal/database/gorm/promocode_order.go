package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
)

// promoSortColumns is the allowlist of order_by fields the PromoCode List accepts,
// mapped to safe physical column names. Restricting to this closed set is what
// prevents the user-supplied order_by from injecting SQL into the ORDER BY clause.
var promoSortColumns = map[string]string{
	"name":          "name",
	"code":          "code",
	"state":         "state",
	"discount_type": "discount_type",
	"create_time":   "create_time",
	"update_time":   "update_time",
}

// orderClause turns an AIP-132 order_by string into a safe "col DIR, ..." clause.
// Unknown fields are rejected with types.ErrInvalidArgument; only allowlisted
// columns and the literals ASC/DESC ever reach the SQL.
func orderClause(orderBy string) (string, error) {
	terms, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		col, ok := promoSortColumns[term.Field]
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
