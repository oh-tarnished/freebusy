package repository

import (
	"fmt"
	"strings"
)

// OrderTerm is a validated sort instruction: a backend-neutral field name plus
// direction. Adapters map Field to their own physical column (GORM) or GraphQL
// order field (Hasura) through a closed allowlist, which is what keeps user input
// out of the SQL/GraphQL order clause.
type OrderTerm struct {
	Field string
	Desc  bool
}

// ParseOrderBy parses an AIP-132 order_by value — a comma-separated list of
// fields, each optionally suffixed with "asc" or "desc" (default asc), e.g.
// "create_time desc, code". It validates only the syntax and direction; callers
// must validate each Field against their sortable allowlist before use. A
// malformed value yields ErrInvalidArgument.
func ParseOrderBy(orderBy string) ([]OrderTerm, error) {
	orderBy = strings.TrimSpace(orderBy)
	if orderBy == "" {
		return nil, nil
	}
	parts := strings.Split(orderBy, ",")
	terms := make([]OrderTerm, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(part)
		if len(fields) == 0 || len(fields) > 2 {
			return nil, fmt.Errorf("%w: malformed order_by term %q", ErrInvalidArgument, strings.TrimSpace(part))
		}
		term := OrderTerm{Field: fields[0]}
		if len(fields) == 2 {
			switch strings.ToLower(fields[1]) {
			case "asc":
			case "desc":
				term.Desc = true
			default:
				return nil, fmt.Errorf("%w: invalid sort direction %q", ErrInvalidArgument, fields[1])
			}
		}
		terms = append(terms, term)
	}
	return terms, nil
}
