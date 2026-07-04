package hasura

import (
	"fmt"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

var propertySortFields = map[string]graphql.StringField{
	"name":         propertiesql.Name,
	"display_name": propertiesql.DisplayName,
	"state":        propertiesql.State,
	"create_time":  propertiesql.CreateTime,
	"update_time":  propertiesql.UpdateTime,
}

var unitSortFields = map[string]graphql.StringField{
	"name":         unitsql.Name,
	"display_name": unitsql.DisplayName,
	"type":         unitsql.Type,
	"state":        unitsql.State,
	"create_time":  unitsql.CreateTime,
	"update_time":  unitsql.UpdateTime,
}

func propertyOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	return orderTerms(orderBy, propertySortFields)
}

func unitOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	return orderTerms(orderBy, unitSortFields)
}

// orderTerms turns an AIP-132 order_by string into the GraphQL order terms a List
// accepts, using the given field allowlist. Unknown fields are rejected with
// types.ErrInvalidArgument.
func orderTerms(orderBy string, fields map[string]graphql.StringField) ([]graphql.OrderTerm, error) {
	parsed, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}
	terms := make([]graphql.OrderTerm, 0, len(parsed))
	for _, t := range parsed {
		field, ok := fields[t.Field]
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
