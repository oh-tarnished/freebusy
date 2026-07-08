package gorm

import (
	"context"
	"errors"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/identity"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// UserRepository is the GORM-backed user repository.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a GORM-backed UserRepository bound to db.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// mapGormErr translates GORM sentinel errors into the provider-neutral errors in
// internal/types.
func mapGormErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return types.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return types.ErrAlreadyExists
	default:
		return err
	}
}

// GetUser returns the user addressed by its resource name.
func (r *UserRepository) GetUser(ctx context.Context, name string) (*identitypbv1.User, error) {
	id, err := types.UserID(name)
	if err != nil {
		return nil, err
	}
	var m identity.User
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, mapGormErr(err)
	}
	return userFromModel(&m), nil
}

// ListUsers returns a page of users ordered by params.OrderBy.
func (r *UserRepository) ListUsers(ctx context.Context, params types.ListParams) ([]*identitypbv1.User, string, error) {
	models, next, err := filterx.Gorm[identity.User](identity.UserFilterSpec).
		List(ctx, r.db, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapGormErr(types.MapFilterxErr(err))
	}
	items := make([]*identitypbv1.User, 0, len(models))
	for i := range models {
		items = append(items, userFromModel(&models[i]))
	}
	return items, next, nil
}

// UpdateUser applies the masked profile fields of u and returns the result. Email
// and identity are IdP-owned and never written; an empty mask replaces every
// mutable profile field. u.Etag guards against concurrent writes.
func (r *UserRepository) UpdateUser(ctx context.Context, u *identitypbv1.User, paths []string) (*identitypbv1.User, error) {
	id, err := types.UserID(u.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var m identity.User
		if e := tx.WithContext(ctx).First(&m, "id = ?", id).Error; e != nil {
			return e
		}
		if u.GetEtag() != "" && m.Etag != nil && u.GetEtag() != *m.Etag {
			return types.ErrConflict
		}
		applyUserMask(&m, u, paths)
		m.Etag = ptr(ulid.GenerateString())
		return identity.NewUserStore(tx).Update(ctx, &m)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetUser(ctx, u.GetName())
}

// applyUserMask overwrites the mutable profile fields selected by paths (an empty
// mask selects them all). Only display_name, avatar_url, locale, and time_zone
// are writable.
func applyUserMask(m *identity.User, u *identitypbv1.User, paths []string) {
	if fieldSelected(paths, "display_name") {
		m.DisplayName = strOrNil(u.GetDisplayName())
	}
	if fieldSelected(paths, "avatar_url") {
		m.AvatarURL = strOrNil(u.GetAvatarUrl())
	}
	if fieldSelected(paths, "locale") {
		m.Locale = strOrNil(u.GetLocale())
	}
	if fieldSelected(paths, "time_zone") {
		m.TimeZone = strOrNil(u.GetTimeZone())
	}
}

// fieldSelected reports whether an update mask selects field. An empty mask
// selects every field.
func fieldSelected(paths []string, field string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, p := range paths {
		if p == field {
			return true
		}
	}
	return false
}
