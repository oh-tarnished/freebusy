package hasura

import (
	"fmt"

	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

var bookingSortFields = map[string]graphql.StringField{
	"name":        resourceql.Name,
	"state":       resourceql.State,
	"create_time": resourceql.CreateTime,
	"update_time": resourceql.UpdateTime,
}

// bookingOrderTerms turns an AIP-132 order_by string into GraphQL order terms.
func bookingOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	parsed, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}
	terms := make([]graphql.OrderTerm, 0, len(parsed))
	for _, t := range parsed {
		field, ok := bookingSortFields[t.Field]
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
