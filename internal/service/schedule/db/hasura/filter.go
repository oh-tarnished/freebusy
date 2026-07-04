package hasura

import (
	"fmt"
	"strings"

	exceptionsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// exceptionFilterPredicate builds the GraphQL where predicate for a unit's
// availability exceptions: always scoped to unitID, AND-combined with the parsed
// filter conditions. Supported fields: kind (`=`/`!=` CLOSURE|EXTRA_HOURS) and
// reason (`=`/`!=`/`:`); a bareword term matches reason.
func exceptionFilterPredicate(conds []types.FilterCondition, unitID string) (graphql.Predicate, error) {
	preds := []graphql.Predicate{exceptionsql.UnitId.Eq(unitID)}
	for _, c := range conds {
		p, err := exceptionCond(c)
		if err != nil {
			return graphql.Predicate{}, err
		}
		preds = append(preds, p)
	}
	if len(preds) == 1 {
		return preds[0], nil
	}
	return graphql.And(preds...), nil
}

func exceptionCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		return exceptionsql.Reason.ILike("%" + c.Value + "%"), nil
	case "reason":
		return textCond(exceptionsql.Reason, c)
	case "kind":
		return enumCond(exceptionsql.Kind, c, "EXCEPTION_KIND_")
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

func textCond(f graphql.StringField, c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Op {
	case types.FilterEq:
		return f.Eq(c.Value), nil
	case types.FilterNeq:
		return f.Neq(c.Value), nil
	case types.FilterHas:
		return f.ILike("%" + c.Value + "%"), nil
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, c.Field)
	}
}

// enumCond matches a stored enum column, accepting the bare value ("CLOSURE") or
// the fully-qualified proto name ("EXCEPTION_KIND_CLOSURE").
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
