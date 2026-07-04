package hasura

import (
	"fmt"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/membersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

var orgSortFields = map[string]graphql.StringField{
	"name":         resourceql.Name,
	"display_name": resourceql.DisplayName,
	"slug":         resourceql.Slug,
	"create_time":  resourceql.CreateTime,
	"update_time":  resourceql.UpdateTime,
}

var memberSortFields = map[string]graphql.StringField{
	"email":       membersql.Email,
	"role":        membersql.Role,
	"state":       membersql.State,
	"create_time": membersql.CreateTime,
	"update_time": membersql.UpdateTime,
}

func orgOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	return orderTerms(orderBy, orgSortFields)
}

func memberOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	return orderTerms(orderBy, memberSortFields)
}

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
