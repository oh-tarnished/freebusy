// Package dbutil holds the small helpers shared by the hand-written Hasura
// repository code across domains — the pieces with no generated (repox)
// equivalent. Anything that grows a generated counterpart should move there
// and leave this package smaller.
package dbutil

import (
	"errors"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapHasuraErr translates the GraphQL client's storage errors onto the shared
// sentinels: an optimistic-concurrency conflict keeps its meaning, and
// unique/duplicate constraint messages surface as AlreadyExists. Everything
// else passes through unchanged.
func MapHasuraErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, graphql.ErrConflict):
		return types.ErrConflict
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") {
		return types.ErrAlreadyExists
	}
	return err
}

// TsToStr renders a timestamp in the RFC3339 UTC form the Hasura timestamptz
// scalar expects; nil renders empty (callers null it via NullableStr).
func TsToStr(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339)
}

// NullableStr wraps a string for a nullable GraphQL input field: empty maps to
// null, anything else to its value.
func NullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}
