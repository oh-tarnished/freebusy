package hasura

import (
	"fmt"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// promoSortFields is the allowlist of order_by fields the PromoCode List accepts,
// mapped to their GraphQL order fields. It mirrors the GORM adapter's allowlist so
// both providers sort by the same set, and keeps user input out of the query.
var promoSortFields = map[string]graphql.StringField{
	"name":          resourceql.Name,
	"code":          resourceql.Code,
	"state":         resourceql.State,
	"discount_type": resourceql.DiscountType,
	"create_time":   resourceql.CreateTime,
	"update_time":   resourceql.UpdateTime,
}

// orderTerms turns an AIP-132 order_by string into GraphQL order terms. Unknown
// fields are rejected with repository.ErrInvalidArgument.
func orderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	terms, err := repository.ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}
	out := make([]graphql.OrderTerm, 0, len(terms))
	for _, term := range terms {
		field, ok := promoSortFields[term.Field]
		if !ok {
			return nil, fmt.Errorf("%w: cannot sort by %q", repository.ErrInvalidArgument, term.Field)
		}
		if term.Desc {
			out = append(out, field.Desc())
		} else {
			out = append(out, field.Asc())
		}
	}
	return out, nil
}
