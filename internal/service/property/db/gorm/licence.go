package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// Licence persistence follows the Unit repository pattern — the attachment
// (scanned certificate) is a belongs-to child in shared.attachments, persisted
// with the licence row in one transaction and replaced wholesale on update.
// One table holds property-wide and per-unit licences alike; target and the
// nullable unit reference say which is which.

// CreateLicence persists l under parent ("properties/{property}") and returns
// the stored record. The caller (runtime layer) has already validated that a
// unit reference, if set, belongs to the parent property.
func (r *PropertyRepository) CreateLicence(ctx context.Context, parent string, l *propertypbv1.Licence) (*propertypbv1.Licence, error) {
	propertyID, id, name, err := types.ResolveLicenceName(parent, l.GetName())
	if err != nil {
		return nil, err
	}
	var unitID *string
	if l.GetUnit() != "" {
		_, uid, err := types.ParseUnitParent(l.GetUnit())
		if err != nil {
			return nil, err
		}
		unitID = &uid
	}
	g := buildLicenceGraph(l, propertyID, unitID)
	g.licence.ID = id
	g.licence.Name = name
	g.licence.State = ptr(property.LicenceStateActive)
	g.licence.Etag = ptr(ulid.GenerateString())

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if g.attachment != nil {
			if e := shared.NewAttachmentStore(tx).Create(ctx, g.attachment); e != nil {
				return e
			}
		}
		return property.NewLicenceStore(tx).Create(ctx, g.licence)
	}); err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetLicence(ctx, name)
}

// GetLicence returns the licence addressed by its resource name.
func (r *PropertyRepository) GetLicence(ctx context.Context, name string) (*propertypbv1.Licence, error) {
	id, err := types.LicenceID(name)
	if err != nil {
		return nil, err
	}
	var m property.Licence
	if err := r.db.WithContext(ctx).Preload("Attachment").First(&m, "id = ?", id).Error; err != nil {
		return nil, mapGormErr(err)
	}
	return licenceFromModel(&m), nil
}

// ListLicences returns a page of licences under parent
// ("properties/{property}") — property-wide and per-unit ones alike; the
// filter narrows by target, unit, type, state, or expiry_date.
func (r *PropertyRepository) ListLicences(ctx context.Context, parent string, params types.ListParams) ([]*propertypbv1.Licence, string, error) {
	propertyID, err := types.PropertyID(parent)
	if err != nil {
		return nil, "", err
	}
	models, next, err := filterx.Gorm[property.Licence](property.LicenceFilterSpec).
		List(ctx, r.db.Preload("Attachment").Where("property_id = ?", propertyID), types.FilterxInput(params))
	if err != nil {
		return nil, "", mapGormErr(types.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Licence, 0, len(models))
	for i := range models {
		items = append(items, licenceFromModel(&models[i]))
	}
	return items, next, nil
}

// UpdateLicence applies the masked fields of l to the stored licence. The
// target and unit are immutable; the attachment is rebuilt from the merged
// proto, and the superseded attachment row is deleted once unreferenced.
func (r *PropertyRepository) UpdateLicence(ctx context.Context, l *propertypbv1.Licence, paths []string) (*propertypbv1.Licence, error) {
	id, err := types.LicenceID(l.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing property.Licence
		if e := tx.WithContext(ctx).Preload("Attachment").First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if l.GetEtag() != "" && existing.Etag != nil && l.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}
		oldAttachment := existing.AttachmentID

		merged := licenceFromModel(&existing)
		applyLicenceMask(merged, l, paths)
		g := buildLicenceGraph(merged, existing.PropertyID, existing.UnitID)

		if g.attachment != nil {
			if e := shared.NewAttachmentStore(tx).Create(ctx, g.attachment); e != nil {
				return e
			}
		}

		existing.Type = g.licence.Type
		existing.LicenceNumber = g.licence.LicenceNumber
		existing.IssuingAuthority = g.licence.IssuingAuthority
		existing.IssueDate = g.licence.IssueDate
		existing.ExpiryDate = g.licence.ExpiryDate
		existing.Notes = g.licence.Notes
		existing.AttachmentID = g.licence.AttachmentID
		existing.Etag = ptr(ulid.GenerateString())
		existing.Attachment, existing.Property, existing.Unit = nil, nil, nil
		if e := property.NewLicenceStore(tx).Update(ctx, &existing); e != nil {
			return e
		}

		if oldAttachment != nil {
			if e := shared.NewAttachmentStore(tx).DeleteByID(ctx, *oldAttachment); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetLicence(ctx, l.GetName())
}

// DeleteLicence removes a licence and its attachment row.
func (r *PropertyRepository) DeleteLicence(ctx context.Context, name string) error {
	id, err := types.LicenceID(name)
	if err != nil {
		return err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing property.Licence
		if e := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if e := property.NewLicenceStore(tx).DeleteByID(ctx, id); e != nil {
			return e
		}
		if existing.AttachmentID != nil {
			return shared.NewAttachmentStore(tx).DeleteByID(ctx, *existing.AttachmentID)
		}
		return nil
	})
	return mapGormErr(err)
}
