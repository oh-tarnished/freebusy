package types

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Pagination defaults applied by PageBounds when ListParams omits or overshoots
// the page size.
const (
	defaultPageSize = 50
	maxPageSize     = 1000
)

// ListParams carries pagination and ordering for List calls. PageToken is an
// opaque cursor produced by a prior List call; an empty token requests the first
// page. OrderBy is an AIP-132 order_by string validated by the adapter against a
// sortable-field allowlist.
type ListParams struct {
	PageSize  int32
	PageToken string
	OrderBy   string
	// Filter is the parsed, AND-combined set of filter conditions (AIP-160).
	// Empty means no filtering. Adapters validate each condition's Field against
	// the columns they can filter and translate the rest to their backend query.
	Filter []FilterCondition
}

// OrderTerm is a validated sort instruction: a backend-neutral field name plus
// direction. Adapters map Field to their own physical column (GORM) or GraphQL
// order field (Hasura) through a closed allowlist, which keeps user input out of
// the SQL/GraphQL order clause.
type OrderTerm struct {
	Field string
	Desc  bool
}

// ParseOrderBy parses an AIP-132 order_by value — a comma-separated list of
// fields, each optionally suffixed with "asc" or "desc" (default asc), e.g.
// "create_time desc, code". It validates only syntax and direction; callers must
// validate each Field against their sortable allowlist. A malformed value yields
// ErrInvalidArgument.
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

// PageBounds resolves the limit/offset window for a List call from params. It
// clamps the page size to [1, maxPageSize] (defaulting when unset) and decodes
// the opaque page token into an offset; a malformed token decodes to offset 0.
func PageBounds(params ListParams) (limit, offset int) {
	limit = int(params.PageSize)
	switch {
	case limit <= 0:
		limit = defaultPageSize
	case limit > maxPageSize:
		limit = maxPageSize
	}
	return limit, decodeOffset(params.PageToken)
}

// EncodeOffset produces the opaque page token addressing the row at the given
// absolute offset. Adapters call it with offset+limit after fetching one extra
// row to detect whether a further page exists.
func EncodeOffset(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeOffset(token string) int {
	if token == "" {
		return 0
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}
