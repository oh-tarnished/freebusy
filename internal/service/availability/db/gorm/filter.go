package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// applyUnitFilter narrows the unit query with the parsed SearchAvailability filter
// (AIP-160, AND-combined). Supported fields: type (`=`/`!=` UNIT_TYPE_*),
// display_name (`=`/`!=`/`:`), tags (`:` membership), and a bareword term matching
// display_name.
func applyUnitFilter(q *gorm.DB, filter string) (*gorm.DB, error) {
	conds, err := types.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	for _, c := range conds {
		clause, args, err := unitCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func unitCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "":
		return `"display_name" ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
	case "display_name":
		switch c.Op {
		case types.FilterEq:
			return `"display_name" = ?`, []any{c.Value}, nil
		case types.FilterNeq:
			return `"display_name" <> ?`, []any{c.Value}, nil
		case types.FilterHas:
			return `"display_name" ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
		default:
			return "", nil, fmt.Errorf("%w: unsupported operator for display_name", types.ErrInvalidArgument)
		}
	case "type":
		if c.Op != types.FilterEq && c.Op != types.FilterNeq {
			return "", nil, fmt.Errorf("%w: unsupported operator for type", types.ErrInvalidArgument)
		}
		val := strings.TrimPrefix(strings.ToUpper(c.Value), "UNIT_TYPE_")
		op := "="
		if c.Op == types.FilterNeq {
			op = "<>"
		}
		return `"type" ` + op + " ?", []any{val}, nil
	case "tags":
		if c.Op != types.FilterHas && c.Op != types.FilterEq {
			return "", nil, fmt.Errorf("%w: tags supports only membership (tags:\"x\")", types.ErrInvalidArgument)
		}
		return "? = ANY(tags)", []any{c.Value}, nil
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

func likeContains(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(v) + "%"
}
