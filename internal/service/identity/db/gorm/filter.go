package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// applyUserFilter narrows q with the parsed conditions (AND-combined). Supported
// fields: display_name and email (`=`/`!=`/`:`); a bareword term matches
// display_name.
func applyUserFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := userCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func userCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "":
		return `"display_name" ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
	case "display_name":
		return textCondition("display_name", c)
	case "email":
		return textCondition("email", c)
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

func textCondition(col string, c types.FilterCondition) (string, []any, error) {
	q := `"` + col + `"`
	switch c.Op {
	case types.FilterEq:
		return q + " = ?", []any{c.Value}, nil
	case types.FilterNeq:
		return q + " <> ?", []any{c.Value}, nil
	case types.FilterHas:
		return q + ` ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
	default:
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, col)
	}
}

func likeContains(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(v) + "%"
}
