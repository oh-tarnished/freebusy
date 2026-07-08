package gorm

import (
	"context"
	"fmt"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
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
		existing.Etag = ptr(ulid.GenerateString())
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
		return nil, mapGormErr(err)
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
		existing.Etag = ptr(ulid.GenerateString())
		return property.NewPropertyStore(tx).Update(ctx, &existing)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetProperty(ctx, name)
}

// UpdateUnit applies the masked fields of u to the stored unit and returns the
// result. The pricing children (price, rate overrides, LOS discounts, fees,
// taxes), media, and applicable-promo-code rows are rebuilt from the merged
// proto; the superseded child rows and their Money/DateRange value-objects are
// deleted once the unit no longer references them.
func (r *PropertyRepository) UpdateUnit(ctx context.Context, u *propertypbv1.Unit, paths []string) (*propertypbv1.Unit, error) {
	id, err := types.UnitID(u.GetName())
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing property.Unit
		if e := preloadUnit(tx.WithContext(ctx)).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if u.GetEtag() != "" && existing.Etag != nil && u.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}
		oldMoney, oldDates := collectUnitValueObjects(&existing)

		merged := unitFromModel(&existing)
		applyUnitMask(merged, u, paths)
		g := buildUnitGraph(merged, existing.PropertyID)
		g.unit.ID = id

		if g.price != nil {
			if e := common.NewMoneyStore(tx).Create(ctx, g.price); e != nil {
				return e
			}
		}

		existing.DisplayName = g.unit.DisplayName
		existing.Description = g.unit.Description
		existing.Type = g.unit.Type
		existing.Capacity = g.unit.Capacity
		existing.MaxOccupancy = g.unit.MaxOccupancy
		existing.TimeZone = g.unit.TimeZone
		existing.PricingUnit = g.unit.PricingUnit
		existing.Duration = g.unit.Duration
		existing.Tags = g.unit.Tags
		existing.Attributes = g.unit.Attributes
		existing.PriceID = g.unit.PriceID
		existing.Etag = ptr(ulid.GenerateString())
		existing.Price, existing.RateOverrides, existing.LosDiscounts = nil, nil, nil
		existing.Fees, existing.Taxes, existing.UnitMedias = nil, nil, nil
		existing.UnitApplicablePromoCodes, existing.UnitsLink, existing.Property = nil, nil, nil
		existing.Licences = nil
		if e := property.NewUnitStore(tx).Update(ctx, &existing); e != nil {
			return e
		}

		for _, model := range []any{
			&property.RateOverride{}, &property.LosDiscount{}, &property.Fee{},
			&property.Tax{}, &property.UnitMedia{}, &property.UnitApplicablePromoCodes{},
		} {
			if e := tx.WithContext(ctx).Where("unit_id = ?", id).Delete(model).Error; e != nil {
				return e
			}
		}
		if e := g.persistChildren(ctx, tx); e != nil {
			return e
		}
		return deleteValueObjects(ctx, tx, oldMoney, oldDates)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetUnit(ctx, u.GetName())
}

// DeleteUnit removes a unit, its pricing children (cascaded by the unit_id
// foreign keys), media, and applicable-promo-code rows, then the Money/DateRange
// value-objects those children referenced. Child licences block the delete
// unless force is set, in which case they (and their attachment rows) go too.
func (r *PropertyRepository) DeleteUnit(ctx context.Context, name string, force bool) error {
	id, err := types.UnitID(name)
	if err != nil {
		return err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var existing property.Unit
		if e := preloadUnit(tx.WithContext(ctx)).First(&existing, "id = ?", id).Error; e != nil {
			return e
		}
		if len(existing.Licences) > 0 && !force {
			return fmt.Errorf("%w: unit has %d licences; set force to delete them too",
				types.ErrInvalidArgument, len(existing.Licences))
		}
		money, dates := collectUnitValueObjects(&existing)
		if e := property.NewUnitStore(tx).DeleteByID(ctx, id); e != nil {
			return e
		}
		attachments := shared.NewAttachmentStore(tx)
		for i := range existing.Licences {
			if aid := existing.Licences[i].AttachmentID; aid != nil {
				if e := attachments.DeleteByID(ctx, *aid); e != nil {
					return e
				}
			}
		}
		return deleteValueObjects(ctx, tx, money, dates)
	})
	return mapGormErr(err)
}

// collectUnitValueObjects returns the ids of the Money and DateRange rows a
// unit's price and pricing children reference, so they can be deleted once the
// referencing rows are gone.
func collectUnitValueObjects(m *property.Unit) (moneyIDs, dateIDs []string) {
	if m.PriceID != nil {
		moneyIDs = append(moneyIDs, *m.PriceID)
	}
	for i := range m.RateOverrides {
		if m.RateOverrides[i].PriceID != "" {
			moneyIDs = append(moneyIDs, m.RateOverrides[i].PriceID)
		}
		if m.RateOverrides[i].DateRangeID != nil {
			dateIDs = append(dateIDs, *m.RateOverrides[i].DateRangeID)
		}
	}
	for i := range m.LosDiscounts {
		if m.LosDiscounts[i].AmountOffID != nil {
			moneyIDs = append(moneyIDs, *m.LosDiscounts[i].AmountOffID)
		}
	}
	for i := range m.Fees {
		if m.Fees[i].AmountID != nil {
			moneyIDs = append(moneyIDs, *m.Fees[i].AmountID)
		}
	}
	return moneyIDs, dateIDs
}

func deleteValueObjects(ctx context.Context, tx *gorm.DB, moneyIDs, dateIDs []string) error {
	moneys := common.NewMoneyStore(tx)
	for _, id := range moneyIDs {
		if e := moneys.DeleteByID(ctx, id); e != nil {
			return e
		}
	}
	dates := shared.NewDateRangeStore(tx)
	for _, id := range dateIDs {
		if e := dates.DeleteByID(ctx, id); e != nil {
			return e
		}
	}
	return nil
}
