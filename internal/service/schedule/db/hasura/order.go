package hasura

import (
	"fmt"

	exceptionsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// exceptionSortFields is the allowlist of order_by fields ListAvailabilityExceptions
// accepts, mapped to the GraphQL columns.
var exceptionSortFields = map[string]graphql.StringField{
	"kind":        exceptionsql.Kind,
	"create_time": exceptionsql.CreateTime,
}

// exceptionOrderTerms turns an AIP-132 order_by string into GraphQL order terms.
func exceptionOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	parsed, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}
	terms := make([]graphql.OrderTerm, 0, len(parsed))
	for _, t := range parsed {
		field, ok := exceptionSortFields[t.Field]
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
