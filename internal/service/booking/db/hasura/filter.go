package hasura

import (
	"fmt"
	"strings"

	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// bookingFilterPredicate translates the parsed filter conditions (AND-combined)
// into a GraphQL where predicate. Supported fields: state (`=`/`!=` BOOKING_STATE_*),
// customer (`=`/`!=`, a users/{user} name), and unit (`=`/`!=`, a unit name). The
// unit/customer FKs store bare ids, so the filter compares against the id segment.
func bookingFilterPredicate(conds []types.FilterCondition) (graphql.Predicate, bool, error) {
	preds := make([]graphql.Predicate, 0, len(conds))
	for _, c := range conds {
		p, err := bookingCond(c)
		if err != nil {
			return graphql.Predicate{}, false, err
		}
		preds = append(preds, p)
	}
	switch len(preds) {
	case 0:
		return graphql.Predicate{}, false, nil
	case 1:
		return preds[0], true, nil
	default:
		return graphql.And(preds...), true, nil
	}
}

func bookingCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "state":
		return enumCond(resourceql.State, c, "BOOKING_STATE_")
	case "customer":
		return idRefCond(resourceql.Customer, c)
	case "unit":
		return idRefCond(resourceql.Unit, c)
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// idRefCond matches a bare-id FK column against the id segment of a resource name
// value ("users/7" -> "7").
func idRefCond(f graphql.StringField, c types.FilterCondition) (graphql.Predicate, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, c.Field)
	}
	id := c.Value
	if i := strings.LastIndex(id, "/"); i >= 0 {
		id = id[i+1:]
	}
	if c.Op == types.FilterNeq {
		return f.Neq(id), nil
	}
	return f.Eq(id), nil
}

// enumCond matches a stored enum column, accepting the bare value or the
// fully-qualified proto name.
func enumCond(f graphql.StringField, c types.FilterCondition, prefix string) (graphql.Predicate, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, c.Field)
	}
	val := strings.TrimPrefix(strings.ToUpper(c.Value), prefix)
	if c.Op == types.FilterNeq {
		return f.Neq(val), nil
	}
	return f.Eq(val), nil
}
