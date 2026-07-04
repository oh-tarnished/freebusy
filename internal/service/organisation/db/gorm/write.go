package gorm

import (
	"context"
	"errors"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

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

// UpdateOrganisation applies the masked fields of o and returns the result.
// An empty paths slice replaces every mutable field; o.Etag guards against
// concurrent writes.
func (r *OrganisationRepository) UpdateOrganisation(ctx context.Context, o *orgpbv1.Organisation, paths []string) (*orgpbv1.Organisation, error) {
	id, err := types.OrganisationID(o.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing organisation.Organisation
		if e := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if o.GetEtag() != "" && existing.Etag != nil && o.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}
		merged := orgFromModel(&existing)
		applyOrgMask(merged, o, paths)
		m := orgToModel(merged)
		existing.DisplayName = m.DisplayName
		existing.Slug = m.Slug
		existing.BillingEmail = m.BillingEmail
		existing.Settings = m.Settings
		existing.Etag = ptr(ulid.GenerateString())
		return organisation.NewOrganisationStore(tx).Update(ctx, &existing)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetOrganisation(ctx, o.GetName())
}

// DeleteOrganisation removes an organisation. Without force it fails when the
// organisation still has members; with force the members cascade in the DB.
func (r *OrganisationRepository) DeleteOrganisation(ctx context.Context, name string, force bool) error {
	id, err := types.OrganisationID(name)
	if err != nil {
		return err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing organisation.Organisation
		if e := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if !force {
			var count int64
			if e := tx.WithContext(ctx).Model(&organisation.Member{}).Where("organisation_id = ?", id).Count(&count).Error; e != nil {
				return e
			}
			if count > 0 {
				return types.ErrConflict
			}
		}
		return organisation.NewOrganisationStore(tx).DeleteByID(ctx, id)
	})
	return mapGormErr(err)
}

// UpdateMember applies the masked fields of mem (role is the only mutable field)
// and returns the result.
func (r *OrganisationRepository) UpdateMember(ctx context.Context, mem *orgpbv1.Member, paths []string) (*orgpbv1.Member, error) {
	id, err := types.MemberID(mem.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing organisation.Member
		if e := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if mem.GetEtag() != "" && existing.Etag != nil && mem.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}
		if inMask(paths, "role") {
			existing.Role = roleToModel(mem.GetRole())
		}
		existing.Etag = ptr(ulid.GenerateString())
		return organisation.NewMemberStore(tx).Update(ctx, &existing)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetMember(ctx, mem.GetName())
}

// DeleteMember removes a member by resource name.
func (r *OrganisationRepository) DeleteMember(ctx context.Context, name string) error {
	id, err := types.MemberID(name)
	if err != nil {
		return err
	}
	if err := organisation.NewMemberStore(r.db.WithContext(ctx)).DeleteByID(ctx, id); err != nil {
		return mapGormErr(err)
	}
	return nil
}
