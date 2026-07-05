// Package gorm provides the GORM-backed implementation of the identity
// persistence contract (internal/service/identity/db.UserRepository). A user is a
// flat resource ("users/{user}"); only profile preferences are writable — email
// and identity are IdP-owned and read-only here.
package gorm

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/identity"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// userFromModel builds the protobuf User from a stored row.
func userFromModel(m *identity.User) *identitypbv1.User {
	return &identitypbv1.User{
		Name:        m.Name,
		Email:       deref(m.Email),
		DisplayName: deref(m.DisplayName),
		AvatarUrl:   deref(m.AvatarURL),
		Locale:      deref(m.Locale),
		TimeZone:    deref(m.TimeZone),
		CreateTime:  timeToTS(&m.CreateTime),
		UpdateTime:  timeToTS(&m.UpdateTime),
		Etag:        deref(m.Etag),
	}
}
