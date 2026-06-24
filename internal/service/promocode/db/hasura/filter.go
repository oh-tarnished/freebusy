package hasura

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// filterPredicate translates the parsed filter conditions (AND-combined) into a
// GraphQL where predicate. The second return is false when there are no
// conditions (graphql.Predicate is a value type and has no nil). Supported
// fields: code and display_name (`=`, `!=`, `:` substring), disabled (`=`/`!=`
// bool), and a bareword free-text term across code and display_name. The derived
// `state` is rejected here — it has no durable column the engine can filter on
// (the gorm provider derives it in SQL); an unknown field is also rejected.
func filterPredicate(conds []types.FilterCondition) (graphql.Predicate, bool, error) {
	preds := make([]graphql.Predicate, 0, len(conds))
	for _, c := range conds {
		p, err := condPredicate(c)
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

func condPredicate(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		pat := "%" + c.Value + "%"
		return graphql.Or(resourceql.Code.ILike(pat), resourceql.DisplayName.ILike(pat)), nil
	case "code", "display_name":
		field := resourceql.Code
		if c.Field == "display_name" {
			field = resourceql.DisplayName
		}
		switch c.Op {
		case types.FilterEq:
			return field.Eq(c.Value), nil
		case types.FilterNeq:
			return field.Neq(c.Value), nil
		case types.FilterHas:
			return field.ILike("%" + c.Value + "%"), nil
		}
		return graphql.Predicate{}, fmt.Errorf("%w: unsupported operator for %q", types.ErrInvalidArgument, c.Field)
	case "disabled":
		if c.Op == types.FilterHas {
			return graphql.Predicate{}, fmt.Errorf("%w: operator `:` is not valid for disabled", types.ErrInvalidArgument)
		}
		var want bool
		switch strings.ToLower(c.Value) {
		case "true":
			want = true
		case "false":
			want = false
		default:
			return graphql.Predicate{}, fmt.Errorf("%w: disabled must be true or false, got %q", types.ErrInvalidArgument, c.Value)
		}
		if c.Op == types.FilterNeq {
			want = !want
		}
		return resourceql.Disabled.Eq(want), nil
	case "state":
		return graphql.Predicate{}, fmt.Errorf("%w: filtering by state is not supported on the hasura provider (use the gorm provider)", types.ErrInvalidArgument)
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}
