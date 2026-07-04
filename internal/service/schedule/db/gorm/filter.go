package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// applyExceptionFilter narrows q with the parsed conditions (AND-combined).
// Supported fields: kind (`=`/`!=` CLOSURE|EXTRA_HOURS) and reason
// (`=`/`!=`/`:`); a bareword term matches reason.
func applyExceptionFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := exceptionCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func exceptionCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "":
		return `"reason" ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
	case "reason":
		return textCondition("reason", c)
	case "kind":
		return enumCondition("kind", c, "EXCEPTION_KIND_")
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
		return q + " ILIKE ? ESCAPE '\\'", []any{likeContains(c.Value)}, nil
	default:
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, col)
	}
}

// enumCondition matches a stored enum column, accepting the bare value ("CLOSURE")
// or the fully-qualified proto name ("EXCEPTION_KIND_CLOSURE").
func enumCondition(col string, c types.FilterCondition, prefix string) (string, []any, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, col)
	}
	val := strings.TrimPrefix(strings.ToUpper(c.Value), prefix)
	op := "="
	if c.Op == types.FilterNeq {
		op = "<>"
	}
	return `"` + col + `" ` + op + " ?", []any{val}, nil
}

func likeContains(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(v) + "%"
}
