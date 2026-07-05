package hasura

import (
	"fmt"

	usersql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/usersql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

var userSortFields = map[string]graphql.StringField{
	"display_name": usersql.DisplayName,
	"email":        usersql.Email,
	"create_time":  usersql.CreateTime,
	"update_time":  usersql.UpdateTime,
}

func userOrderTerms(orderBy string) ([]graphql.OrderTerm, error) {
	parsed, err := types.ParseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}
	terms := make([]graphql.OrderTerm, 0, len(parsed))
	for _, t := range parsed {
		field, ok := userSortFields[t.Field]
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
