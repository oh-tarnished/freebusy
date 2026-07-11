// Package db is the identity persistence seam. The read/update surface is the
// generated provider-agnostic repositories
// (internal/database/repository/freebusy/identity — GORM or Hasura behind one
// interface); this package narrows them to the service-facing contract.
// Identity is deliberately thin: users are provisioned by the IdP/auth layer
// (there is no CreateUser RPC), so this contract covers only reading and
// updating an existing user's profile.
package db

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/identity"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
)

// UserRepository provides read/update persistence for users. Errors are the
// repox sentinels (aliased in internal/types).
type UserRepository interface {
	// GetUser returns the user by resource name ("users/{user}"), or
	// repox.ErrNotFound. The "users/me" alias is resolved by the caller before
	// this is reached.
	GetUser(ctx context.Context, name string) (*identitypbv1.User, error)

	// ListUsers returns a page of users and an opaque next-page token.
	ListUsers(ctx context.Context, in repox.ListInput) (items []*identitypbv1.User, nextPageToken string, err error)

	// UpdateUser persists the profile fields named by paths (an AIP-134 field
	// mask); an empty mask replaces all mutable profile fields. Email and identity
	// are IdP-owned (OUTPUT_ONLY) and never written. u.Etag guards against
	// concurrent writes.
	UpdateUser(ctx context.Context, u *identitypbv1.User, paths []string) (*identitypbv1.User, error)
}

// New returns the UserRepository for the configured provider
// ([database].provider; GORM by default, Hasura opt-in), built on the generated
// repositories.
func New(conn *database.Connection) UserRepository {
	c := repox.Conn{Gorm: conn.PgSQLConn}
	if database.ProviderFromConfig() == database.ProviderHasura {
		c = repox.Conn{GraphQL: conn.Hasura}
	}
	return &repos{gen: identity.New(c)}
}

// repos maps the service contract onto the generated repository set.
type repos struct {
	gen identity.Repositories
}

func (r *repos) GetUser(ctx context.Context, name string) (*identitypbv1.User, error) {
	return r.gen.Users.Get(ctx, name)
}

func (r *repos) ListUsers(ctx context.Context, in repox.ListInput) ([]*identitypbv1.User, string, error) {
	return r.gen.Users.List(ctx, in)
}

func (r *repos) UpdateUser(ctx context.Context, u *identitypbv1.User, paths []string) (*identitypbv1.User, error) {
	return r.gen.Users.Update(ctx, u, paths)
}
