// Package gorm provides the GORM-backed implementation of the property
// persistence contract (internal/service/property/db.PropertyRepository). It
// adapts the generated per-entity stores under
// internal/database/gorm/freebusy/property to that contract, converting between
// protobuf domain types and the relational storage models.
package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"gorm.io/gorm"
)

// PropertyRepository is the GORM-backed property repository. Each write persists
// a property (or unit) together with its normalized child rows (address, policy,
// media; and for a unit its price, rate overrides, LOS discounts, fees, taxes,
// media, and applicable-promo-code join rows) inside a single transaction; each
// read re-hydrates them via preloads.
type PropertyRepository struct {
	db *gorm.DB
}

// NewPropertyRepository returns a GORM-backed PropertyRepository bound to db.
// The parent db package asserts it satisfies db.PropertyRepository.
func NewPropertyRepository(db *gorm.DB) *PropertyRepository {
	return &PropertyRepository{db: db}
}

// preloadProperty eager-loads a property's association graph: its address,
// policy, media gallery, and the child units and licences (names only surface
// on the proto).
func preloadProperty(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Address").
		Preload("Policy").
		Preload("Medias").
		Preload("Units").
		Preload("Licences")
}

// preloadUnit eager-loads a unit's association graph: its price and every
// pricing child (rate overrides with their Money + DateRange, LOS discounts and
// fees with their Money, taxes), media gallery, and applicable-promo-code rows.
func preloadUnit(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Price").
		Preload("RateOverrides.DateRange").
		Preload("RateOverrides.Price").
		Preload("LosDiscounts.AmountOff").
		Preload("Fees.Amount").
		Preload("Taxes").
		Preload("UnitMedias").
		Preload("UnitApplicablePromoCodes").
		Preload("Licences")
}

// persist inserts a property graph: its belongs-to children (address, policy)
// first, then the property row, then its has-many media (which carry the FK).
func (g *propertyGraph) persist(ctx context.Context, tx *gorm.DB) error {
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
	if e := property.NewPropertyStore(tx).Create(ctx, g.property); e != nil {
		return e
	}
	medias := property.NewMediaStore(tx)
	for _, m := range g.medias {
		m.PropertyID = g.property.ID
		if e := medias.Create(ctx, m); e != nil {
			return e
		}
	}
	return nil
}

// persist inserts a unit graph: the price Money (referenced by the unit) first,
// then the unit row, then its pricing children.
func (g *unitGraph) persist(ctx context.Context, tx *gorm.DB) error {
	if g.price != nil {
		if e := common.NewMoneyStore(tx).Create(ctx, g.price); e != nil {
			return e
		}
	}
	if e := property.NewUnitStore(tx).Create(ctx, g.unit); e != nil {
		return e
	}
	return g.persistChildren(ctx, tx)
}

// persistChildren inserts the value-object Money/DateRange rows the pricing
// children reference, then the children themselves (which carry the unit FK).
// The unit row must already exist. Used by both create and update.
func (g *unitGraph) persistChildren(ctx context.Context, tx *gorm.DB) error {
	moneys := common.NewMoneyStore(tx)
	for _, m := range g.moneys {
		if e := moneys.Create(ctx, m); e != nil {
			return e
		}
	}
	dates := shared.NewDateRangeStore(tx)
	for _, d := range g.dates {
		if e := dates.Create(ctx, d); e != nil {
			return e
		}
	}
	ro := property.NewRateOverrideStore(tx)
	for _, row := range g.rateOverrides {
		row.UnitID = g.unit.ID
		if e := ro.Create(ctx, row); e != nil {
			return e
		}
	}
	ld := property.NewLosDiscountStore(tx)
	for _, row := range g.losDiscounts {
		row.UnitID = g.unit.ID
		if e := ld.Create(ctx, row); e != nil {
			return e
		}
	}
	fee := property.NewFeeStore(tx)
	for _, row := range g.fees {
		row.UnitID = g.unit.ID
		if e := fee.Create(ctx, row); e != nil {
			return e
		}
	}
	tax := property.NewTaxStore(tx)
	for _, row := range g.taxes {
		row.UnitID = g.unit.ID
		if e := tax.Create(ctx, row); e != nil {
			return e
		}
	}
	medias := property.NewUnitMediaStore(tx)
	for _, row := range g.medias {
		row.UnitID = g.unit.ID
		if e := medias.Create(ctx, row); e != nil {
			return e
		}
	}
	codes := property.NewUnitApplicablePromoCodesStore(tx)
	for _, row := range g.promoCodes {
		row.UnitID = g.unit.ID
		if e := codes.Create(ctx, row); e != nil {
			return e
		}
	}
	return nil
}

// --- Property reads/creates --------------------------------------------------

// --- Unit reads/creates ------------------------------------------------------
