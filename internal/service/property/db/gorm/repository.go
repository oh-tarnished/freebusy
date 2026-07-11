// Package gorm provides the GORM-backed implementation of the property
// persistence contract (internal/service/property/db.PropertyRepository). It
// adapts the generated per-entity stores under
// internal/database/gorm/freebusy/property to that contract, converting between
// protobuf domain types and the relational storage models.
package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
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

// CreateProperty persists p and returns the stored record. The resource name is
// taken from p.Name when present, otherwise a fresh ULID id is assigned.
func (r *PropertyRepository) CreateProperty(ctx context.Context, p *propertypbv1.Property) (*propertypbv1.Property, error) {
	id, name, err := types.ResolvePropertyName(p.GetName())
	if err != nil {
		return nil, err
	}
	g := buildPropertyGraph(p)
	g.property.ID = id
	g.property.Name = name
	g.property.Etag = repox.Ptr(ulid.GenerateString())

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		return g.persist(ctx, tx)
	}); err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.GetProperty(ctx, name)
}

// GetProperty returns the property addressed by its resource name.
func (r *PropertyRepository) GetProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(name)
	if err != nil {
		return nil, err
	}
	var m property.Property
	if err := preloadProperty(r.db.WithContext(ctx)).First(&m, "id = ?", id).Error; err != nil {
		return nil, repox.MapGormErr(err)
	}
	return propertyFromModel(&m), nil
}

// ListProperties returns a page of properties ordered by in.OrderBy.
func (r *PropertyRepository) ListProperties(ctx context.Context, in repox.ListInput) ([]*propertypbv1.Property, string, error) {
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	models, next, err := filterx.Gorm[property.Property](property.PropertyFilterSpec).
		List(ctx, preloadProperty(r.db), fin)
	if err != nil {
		return nil, "", repox.MapGormErr(repox.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Property, 0, len(models))
	for i := range models {
		items = append(items, propertyFromModel(&models[i]))
	}
	return items, next, nil
}

// --- Unit reads/creates ------------------------------------------------------

// CreateUnit persists u under parent ("properties/{property}") and returns the
// stored record.
func (r *PropertyRepository) CreateUnit(ctx context.Context, parent string, u *propertypbv1.Unit) (*propertypbv1.Unit, error) {
	propertyID, id, name, err := types.ResolveUnitName(parent, u.GetName())
	if err != nil {
		return nil, err
	}
	g := buildUnitGraph(u, propertyID)
	g.unit.ID = id
	g.unit.Name = name
	g.unit.Etag = repox.Ptr(ulid.GenerateString())

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		return g.persist(ctx, tx)
	}); err != nil {
		return nil, repox.MapGormErr(err)
	}
	return r.GetUnit(ctx, name)
}

// GetUnit returns the unit addressed by its resource name.
func (r *PropertyRepository) GetUnit(ctx context.Context, name string) (*propertypbv1.Unit, error) {
	id, err := types.UnitID(name)
	if err != nil {
		return nil, err
	}
	var m property.Unit
	if err := preloadUnit(r.db.WithContext(ctx)).First(&m, "id = ?", id).Error; err != nil {
		return nil, repox.MapGormErr(err)
	}
	return unitFromModel(&m), nil
}

// ListUnits returns a page of units under parent ("properties/{property}").
func (r *PropertyRepository) ListUnits(ctx context.Context, parent string, in repox.ListInput) ([]*propertypbv1.Unit, string, error) {
	propertyID, err := types.PropertyID(parent)
	if err != nil {
		return nil, "", err
	}
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	models, next, err := filterx.Gorm[property.Unit](property.UnitFilterSpec).
		List(ctx, preloadUnit(r.db).Where("property_id = ?", propertyID), fin)
	if err != nil {
		return nil, "", repox.MapGormErr(repox.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Unit, 0, len(models))
	for i := range models {
		items = append(items, unitFromModel(&models[i]))
	}
	return items, next, nil
}
