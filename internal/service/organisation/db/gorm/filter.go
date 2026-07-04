package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// applyOrgFilter narrows q with the parsed conditions (AND-combined). Supported
// fields: display_name (`=`/`!=`/`:`), slug (`=`/`!=`), state (`=`/`!=`
// ACTIVE|SUSPENDED); a bareword term matches display_name.
func applyOrgFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := orgCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func orgCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "":
		return `"display_name" ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
	case "display_name":
		return textCondition("display_name", c)
	case "slug":
		return eqCondition("slug", c)
	case "state":
		return enumCondition("state", c, "ORGANISATION_STATE_")
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// applyMemberFilter narrows q. Supported fields: email (`=`/`!=`/`:`), role
// (`=`/`!=`), state (`=`/`!=`); a bareword term matches email.
func applyMemberFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := memberCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func memberCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "":
		return `"email" ILIKE ? ESCAPE '\'`, []any{likeContains(c.Value)}, nil
	case "email":
		return textCondition("email", c)
	case "role":
		return enumCondition("role", c, "ORGANISATION_ROLE_")
	case "state":
		return enumCondition("state", c, "MEMBER_STATE_")
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

func eqCondition(col string, c types.FilterCondition) (string, []any, error) {
	q := `"` + col + `"`
	switch c.Op {
	case types.FilterEq:
		return q + " = ?", []any{c.Value}, nil
	case types.FilterNeq:
		return q + " <> ?", []any{c.Value}, nil
	default:
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, col)
	}
}

// enumCondition matches a stored enum column, accepting the bare value ("ADMIN")
// or the fully-qualified proto name ("ORGANISATION_ROLE_ADMIN").
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
