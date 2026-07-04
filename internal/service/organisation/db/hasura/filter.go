package hasura

import (
	"fmt"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/membersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// orgFilterPredicate translates the parsed filter conditions (AND-combined) into
// a GraphQL where predicate. Supported: display_name (`=`/`!=`/`:`), slug
// (`=`/`!=`), state (`=`/`!=`); a bareword term matches display_name.
func orgFilterPredicate(conds []types.FilterCondition) (graphql.Predicate, bool, error) {
	preds := make([]graphql.Predicate, 0, len(conds))
	for _, c := range conds {
		p, err := orgCond(c)
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

func orgCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		return resourceql.DisplayName.ILike("%" + c.Value + "%"), nil
	case "display_name":
		return textPred(resourceql.DisplayName, c)
	case "slug":
		return eqPred(resourceql.Slug, c)
	case "state":
		return enumPred(resourceql.State, c, "ORGANISATION_STATE_")
	default:
		return graphql.Predicate{}, fmt.Errorf("%w: cannot filter by %q", types.ErrInvalidArgument, c.Field)
	}
}

// memberFilterPredicate scopes to parent organisation AND the parsed conditions.
// Supported: email (`=`/`!=`/`:`), role (`=`/`!=`), state (`=`/`!=`).
func memberFilterPredicate(conds []types.FilterCondition, orgID string) (graphql.Predicate, error) {
	preds := []graphql.Predicate{membersql.OrganisationId.Eq(orgID)}
	for _, c := range conds {
		p, err := memberCond(c)
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

func memberCond(c types.FilterCondition) (graphql.Predicate, error) {
	switch c.Field {
	case "":
		return membersql.Email.ILike("%" + c.Value + "%"), nil
	case "email":
		return textPred(membersql.Email, c)
	case "role":
		return enumPred(membersql.Role, c, "ORGANISATION_ROLE_")
	case "state":
		return enumPred(membersql.State, c, "MEMBER_STATE_")
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
