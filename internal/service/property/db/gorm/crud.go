// Property and unit reads/creates over the generated stores.
package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

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
