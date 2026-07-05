// Package hasura provides the Hasura/GraphQL-backed implementation of the
// identity persistence contract (internal/service/identity/db.UserRepository).
// Only profile preferences are writable; email and identity are IdP-owned.
package hasura

import (
	"errors"
	"strings"
	"time"

	identityschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const rfc3339 = time.RFC3339

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func tsToStr(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(rfc3339)
}

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	t, err := time.Parse(rfc3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

// nullableStr maps an empty optional string to a SQL NULL update and a non-empty
// one to a value update, so clearing a profile field clears the column.
func nullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}

func userFromSchema(u *identityschema.IdentityUsers) *identitypbv1.User {
	return &identitypbv1.User{
		Name:        u.Name,
		Email:       deref(u.Email),
		DisplayName: deref(u.DisplayName),
		AvatarUrl:   deref(u.AvatarUrl),
		Locale:      deref(u.Locale),
		TimeZone:    deref(u.TimeZone),
		CreateTime:  strToTS(u.CreateTime),
		UpdateTime:  strToTS(u.UpdateTime),
		Etag:        deref(u.Etag),
	}
}

// mapHasuraErr translates GraphQL/runtime errors into the repository sentinels.
func mapHasuraErr(err error) error {
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
