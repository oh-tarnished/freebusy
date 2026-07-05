// Package db is the identity persistence layer. It defines the provider-agnostic
// UserRepository contract (spoken in protobuf domain types) and a factory that
// builds the implementation for the configured backend. Identity is deliberately
// thin: users are provisioned by the IdP/auth layer (there is no CreateUser RPC),
// so this contract covers only reading and updating an existing user's profile.
// Shared, provider-neutral vocabulary (errors, list params, names, field masks)
// lives in internal/types.
package db

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/service/identity/db/gorm"
	"github.com/oh-tarnished/freebusy/internal/service/identity/db/hasura"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
)

// UserRepository provides read/update persistence for users. Errors are the
// sentinels in internal/types (types.ErrNotFound, types.ErrConflict, …).
type UserRepository interface {
	// GetUser returns the user by resource name ("users/{user}"), or
	// types.ErrNotFound. The "users/me" alias is resolved by the caller before
	// this is reached.
	GetUser(ctx context.Context, name string) (*identitypbv1.User, error)

	// ListUsers returns a page of users and an opaque next-page token.
	ListUsers(ctx context.Context, params types.ListParams) (items []*identitypbv1.User, nextPageToken string, err error)

	// UpdateUser persists the profile fields named by paths (an AIP-134 field
	// mask); an empty mask replaces all mutable profile fields. Email and identity
	// are IdP-owned and never written. u.Etag guards against concurrent writes.
	UpdateUser(ctx context.Context, u *identitypbv1.User, paths []string) (*identitypbv1.User, error)
}

// Assert the provider implementations satisfy the contract here.
var (
	_ UserRepository = (*gorm.UserRepository)(nil)
	_ UserRepository = (*hasura.UserRepository)(nil)
)

// New returns the UserRepository for the configured provider, built over the
// matching handle on conn ([database].provider; GORM by default, Hasura opt-in).
func New(conn *database.Connection) UserRepository {
	if database.ProviderFromConfig() == database.ProviderHasura {
		return hasura.NewUserRepository(conn.Hasura)
	}
	return gorm.NewUserRepository(conn.PgSQLConn)
}
