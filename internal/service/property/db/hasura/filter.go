package hasura

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// propertyFilterPredicate translates the parsed filter conditions (AND-combined)
// into a GraphQL where predicate. Supported fields: organisation (`=`/`!=`),
// display_name (`=`/`!=`/`:`), state (`=`/`!=` ACTIVE|ARCHIVED), and a bareword
// term matching display_name. The bool is false when there are no conditions.
func propertyFilterPredicate(conds []types.FilterCondition) (graphql.Predicate, bool, error) {
	preds := make([]graphql.Predicate, 0, len(conds))
	for _, c := range conds {
		p, err := propertyCond(c)
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

func propertyCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		return propertiesql.DisplayName.ILike("%" + c.Value + "%"), nil
	case "display_name":
		return textPred(propertiesql.DisplayName, c)
	case "organisation":
		return eqPred(propertiesql.Organisation, c)
	case "state":
		return enumPred(propertiesql.State, c, "PROPERTY_STATE_")
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// unitFilterPredicate builds a where predicate scoped to parent property
// (property_id = propertyID) AND the parsed conditions. Supported fields:
// display_name (`=`/`!=`/`:`), type (`=`/`!=`), state (`=`/`!=`), and a bareword
// term matching display_name.
func unitFilterPredicate(conds []types.FilterCondition, propertyID string) (graphql.Predicate, error) {
	preds := []graphql.Predicate{unitsql.PropertyId.Eq(propertyID)}
	for _, c := range conds {
		p, err := unitCond(c)
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

func unitCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		return unitsql.DisplayName.ILike("%" + c.Value + "%"), nil
	case "display_name":
		return textPred(unitsql.DisplayName, c)
	case "type":
		return enumPred(unitsql.Type, c, "UNIT_TYPE_")
	case "state":
		return enumPred(unitsql.State, c, "UNIT_STATE_")
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

func textPred(field graphql.StringField, c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Op {
	case types.FilterEq:
		return field.Eq(c.Value), nil
	case types.FilterNeq:
		return field.Neq(c.Value), nil
	case types.FilterHas:
		return field.ILike("%" + c.Value + "%"), nil
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator", types.ErrInvalidArgument)
	}
}

func eqPred(field graphql.StringField, c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Op {
	case types.FilterEq:
		return field.Eq(c.Value), nil
	case types.FilterNeq:
		return field.Neq(c.Value), nil
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator", types.ErrInvalidArgument)
	}
}

// enumPred matches a stored enum column, accepting the bare value ("ROOM") or the
// fully-qualified proto name ("UNIT_TYPE_ROOM").
func enumPred(field graphql.StringField, c types.FilterCondition, prefix string) (graphql.Predicate, error) {
	if c.Op != types.FilterEq && c.Op != types.FilterNeq {
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator", types.ErrInvalidArgument)
	}
	val := strings.TrimPrefix(strings.ToUpper(c.Value), prefix)
	if c.Op == types.FilterNeq {
		return field.Neq(val), nil
	}
	return field.Eq(val), nil
}
