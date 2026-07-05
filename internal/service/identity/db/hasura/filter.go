package hasura

import (
	"fmt"

	usersql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/usersql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// userFilterPredicate translates the parsed filter conditions (AND-combined) into
// a GraphQL where predicate. Supported fields: display_name and email
// (`=`/`!=`/`:`); a bareword term matches display_name. The bool is false when
// there are no conditions.
func userFilterPredicate(conds []types.FilterCondition) (graphql.Predicate, bool, error) {
	preds := make([]graphql.Predicate, 0, len(conds))
	for _, c := range conds {
		p, err := userCond(c)
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

func userCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		return usersql.DisplayName.ILike("%" + c.Value + "%"), nil
	case "display_name":
		return textCond(usersql.DisplayName, c)
	case "email":
		return textCond(usersql.Email, c)
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
