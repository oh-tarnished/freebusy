package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

const (
	propertiesTable = `"property"."properties"`
	unitsTable      = `"property"."units"`
)

// applyPropertyFilter narrows q with the parsed conditions (AND-combined).
// Supported fields: organisation (`=`/`!=`), display_name (`=`/`!=`/`:`),
// state (`=`/`!=` ACTIVE|ARCHIVED), and tags (`:` membership). A bareword term
// with no field is a case-insensitive search on display_name.
func applyPropertyFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := propertyCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func propertyCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "":
		pat := likeContains(c.Value)
		return propertiesTable + `."display_name" ILIKE ? ESCAPE '\'`, []any{pat}, nil
	case "display_name":
		return textCondition(propertiesTable, c)
	case "organisation":
		return eqCondition(propertiesTable, "organisation", c)
	case "state":
		return stateCondition(propertiesTable, c, "PROPERTY_STATE_")
	case "tags":
		return propertiesTable + `."tags" @> ARRAY[?]`, []any{c.Value}, nil
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// applyUnitFilter narrows q with the parsed conditions (AND-combined). Supported
// fields: display_name (`=`/`!=`/`:`), type (`=`/`!=`), state (`=`/`!=`
// ACTIVE|ARCHIVED), and tags (`:` membership). A bareword term searches
// display_name.
func applyUnitFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
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
		pat := likeContains(c.Value)
		return unitsTable + `."display_name" ILIKE ? ESCAPE '\'`, []any{pat}, nil
	case "display_name":
		return textCondition(unitsTable, c)
	case "type":
		return enumCondition(unitsTable, "type", c, "UNIT_TYPE_")
	case "state":
		return stateCondition(unitsTable, c, "UNIT_STATE_")
	case "tags":
		return unitsTable + `."tags" @> ARRAY[?]`, []any{c.Value}, nil
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// textCondition handles a text column with exact, negated, and substring (`:`)
// matching.
func textCondition(table string, c types.FilterCondition) (string, []any, error) {
	col := table + `."` + c.Field + `"`
	switch c.Op {
	case types.FilterEq:
		return col + " = ?", []any{c.Value}, nil
	case types.FilterNeq:
		return col + " <> ?", []any{c.Value}, nil
	case types.FilterHas:
		return col + " ILIKE ? ESCAPE '\\'", []any{likeContains(c.Value)}, nil
	default:
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, c.Field)
	}
}

// eqCondition handles an exact-match column (`=`/`!=` only).
func eqCondition(table, field string, c types.FilterCondition) (string, []any, error) {
	col := table + `."` + field + `"`
	switch c.Op {
	case types.FilterEq:
		return col + " = ?", []any{c.Value}, nil
	case types.FilterNeq:
		return col + " <> ?", []any{c.Value}, nil
	default:
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, field)
	}
}

// enumCondition matches a stored enum column, accepting either the bare value
// ("ROOM") or the fully-qualified proto name ("UNIT_TYPE_ROOM").
func enumCondition(table, field string, c types.FilterCondition, prefix string) (string, []any, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, field)
	}
	val := strings.TrimPrefix(strings.ToUpper(c.Value), prefix)
	op := "="
	if c.Op == types.FilterNeq {
		op = "<>"
	}
	return table + `."` + field + `" ` + op + " ?", []any{val}, nil
}

// stateCondition matches the stored `state` column against ACTIVE|ARCHIVED,
// accepting the bare or fully-qualified form.
func stateCondition(table string, c types.FilterCondition, prefix string) (string, []any, error) {
	return enumCondition(table, "state", c, prefix)
}

// likeContains builds a case-insensitive "contains" pattern, escaping the LIKE
// wildcards in the user value so they match literally.
func likeContains(v string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(v) + "%"
}
