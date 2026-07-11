package gorm

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// UpdateProperty applies the masked fields of p to the stored record and returns
// the result. An empty paths slice replaces every mutable field; p.Etag, when
// set, guards against concurrent writes. The merged proto is re-materialized into
// a fresh child graph (address, policy, media), the property row is repointed at
// it, and the superseded belongs-to rows are deleted once unreferenced.
func (r *PropertyRepository) UpdateProperty(ctx context.Context, p *propertypbv1.Property, paths []string) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(p.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing property.Property
		if e := preloadProperty(tx.WithContext(ctx)).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if p.GetEtag() != "" && existing.Etag != nil && p.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}
		oldAddr, oldPolicy := existing.AddressID, existing.PolicyID

		merged := propertyFromModel(&existing)
		applyPropertyMask(merged, p, paths)
		g := buildPropertyGraph(merged)
		g.property.ID = id

		if g.address != nil {
			if e := common.NewPostalAddressStore(tx).Create(ctx, g.address); e != nil {
				return e
			}
		}
		if g.policy != nil {
			if e := property.NewPolicyStore(tx).Create(ctx, g.policy); e != nil {
				return e
			}
		}

		existing.OrganisationID = g.property.OrganisationID
		existing.DisplayName = g.property.DisplayName
		existing.Description = g.property.Description
		existing.TimeZone = g.property.TimeZone
		existing.Tags = g.property.Tags
		existing.Attributes = g.property.Attributes
		existing.AddressID = g.property.AddressID
		existing.PolicyID = g.property.PolicyID
		existing.Etag = repox.Ptr(ulid.GenerateString())
		existing.Address, existing.Policy, existing.Medias, existing.Units, existing.UnitsLink = nil, nil, nil, nil, nil
		existing.Licences = nil
		if e := property.NewPropertyStore(tx).Update(ctx, &existing); e != nil {
			return e
		}

		if e := tx.WithContext(ctx).Where("property_id = ?", id).Delete(&property.Media{}).Error; e != nil {
			return e
		}
		medias := property.NewMediaStore(tx)
		for _, m := range g.medias {
			m.PropertyID = id
			if e := medias.Create(ctx, m); e != nil {
				return e
			}
		}

		if oldAddr != nil {
			if e := common.NewPostalAddressStore(tx).DeleteByID(ctx, *oldAddr); e != nil {
				return e
			}
		}
		if oldPolicy != nil {
			if e := property.NewPolicyStore(tx).DeleteByID(ctx, *oldPolicy); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.GetProperty(ctx, p.GetName())
}

// ArchiveProperty flips a property to the ARCHIVED state; UnarchiveProperty
// restores it to ACTIVE. Both bump the etag.
func (r *PropertyRepository) ArchiveProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	return r.setPropertyState(ctx, name, property.PropertyStateArchived)
}

func (r *PropertyRepository) UnarchiveProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	return r.setPropertyState(ctx, name, property.PropertyStateActive)
}

func (r *PropertyRepository) setPropertyState(ctx context.Context, name string, state property.PropertyState) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(name)
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing property.Property
		if e := tx.WithContext(ctx).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		existing.State = &state
		existing.Etag = repox.Ptr(ulid.GenerateString())
		return property.NewPropertyStore(tx).Update(ctx, &existing)
	})
	if err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.GetProperty(ctx, name)
}
