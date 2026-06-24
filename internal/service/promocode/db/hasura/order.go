package hasura

import (
	"fmt"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// promoSortFields is the allowlist of order_by fields the PromoCode List accepts,
// mapped to the generated GraphQL order field. Restricting to this closed set is
// what keeps the user-supplied order_by from naming an arbitrary column.
var promoSortFields = map[string]graphql.StringField{
	"name":        resourceql.Name,
	"code":        resourceql.Code,
	"state":       resourceql.State,
	"create_time": resourceql.CreateTime,
	"update_time": resourceql.UpdateTime,
}

// orderTerms turns an AIP-132 order_by string into the GraphQL order terms the
// resource List accepts. Unknown fields are rejected with types.ErrInvalidArgument.
func orderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	parsed, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}
	terms := make([]graphql.OrderTerm, 0, len(parsed))
	for _, t := range parsed {
		field, ok := promoSortFields[t.Field]
		if !ok {
			return nil, fmt.Errorf("%w: cannot sort by %q", types.ErrInvalidArgument, t.Field)
		}
		if t.Desc {
			terms = append(terms, field.Desc())
		} else {
			terms = append(terms, field.Asc())
		}
	}
	return terms, nil
}
