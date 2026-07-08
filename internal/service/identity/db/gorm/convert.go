// Package gorm provides the GORM-backed implementation of the identity
// persistence contract (internal/service/identity/db.UserRepository). A user is a
// flat resource ("users/{user}"); only profile preferences are writable — email
// and identity are IdP-owned and read-only here.
package gorm

import (
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/identity"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
)

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// userFromModel builds the protobuf User from a stored row via the generated
// converter (the models package's protobuf.go).
func userFromModel(m *identity.User) *identitypbv1.User {
	return identity.UserToProto(m)
}
