package gorm

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/types"
	"gorm.io/gorm"
)

// applyBookingFilter narrows q with the parsed conditions (AND-combined).
// Supported fields: state (`=`/`!=` BOOKING_STATE_*), customer (`=`/`!=`, a
// users/{user} name), and unit (`=`/`!=`, a unit name). The unit/customer FKs
// store bare ids, so the filter compares against the id segment of the value.
func applyBookingFilter(q *gorm.DB, conds []types.FilterCondition) (*gorm.DB, error) {
	for _, c := range conds {
		clause, args, err := bookingCondition(c)
		if err != nil {
			return nil, err
		}
		q = q.Where(clause, args...)
	}
	return q, nil
}

func bookingCondition(c types.FilterCondition) (string, []any, error) {
	switch c.Field {
	case "state":
		return enumCondition("state", c, "BOOKING_STATE_")
	case "customer":
		return idRefCondition("customer", c)
	case "unit":
		return idRefCondition("unit", c)
	default:
		return "", nil, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// idRefCondition matches a bare-id FK column against the id segment of a resource
// name value ("users/7" -> "7").
func idRefCondition(col string, c types.FilterCondition) (string, []any, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return "", nil, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, col)
	}
	id := c.Value
	if i := strings.LastIndex(id, "/"); i >= 0 {
		id = id[i+1:]
	}
	op := "="
	if c.Op == types.FilterNeq {
		op = "<>"
	}
	return `"` + col + `" ` + op + " ?", []any{id}, nil
}

// enumCondition matches a stored enum column, accepting the bare value or the
// fully-qualified proto name.
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
