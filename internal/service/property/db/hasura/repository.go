// Package hasura provides the Hasura/GraphQL-backed implementation of the
// property persistence contract (internal/service/property/db.PropertyRepository).
// It adapts the generated freebusyql handlers to that contract, converting between
// protobuf domain types and the normalized GraphQL schema (the address, policy,
// media, and a unit's pricing child tables, plus the common Money/PostalAddress
// and shared DateRange value-objects).
package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/feesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/losdiscountsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/mediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/rateoverridesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/taxesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitapplicablepromocodesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitmediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// PropertyRepository is the Hasura-backed property repository. Hasura exposes no
// client-side transactions across calls, but its mutation API can run several
// mutations as one atomic GraphQL document (svc.Mutation.Tx()); writes use that
// so a property/unit and its child rows commit together or not at all. Each read
// hydrates the row's children with follow-up queries.
type PropertyRepository struct {
	svc *freebusyql.Service
}

// NewPropertyRepository returns a Hasura-backed PropertyRepository bound to svc.
func NewPropertyRepository(svc *freebusyql.Service) *PropertyRepository {
	return &PropertyRepository{svc: svc}
}

// --- Property ----------------------------------------------------------------

func (r *PropertyRepository) CreateProperty(ctx context.Context, p *propertypbv1.Property) (*propertypbv1.Property, error) {
	id, name, err := types.ResolvePropertyName(p.GetName())
	if err != nil {
		return nil, err
	}
	g := buildPropertyGraph(p, time.Now().UTC())
	g.property.Id = id
	g.property.Name = name
	g.property.Etag = ulid.GenerateString()

	tx := r.svc.Mutation.Tx()
	if g.address != nil {
		var res commonschema.InsertCommonPostalAddressResponse
		tx.Add(r.svc.Mutation.Common.PostalAddress.CreateOp(*g.address, &res))
	}
	if g.policy != nil {
		var res pschema.InsertPropertyPoliciesResponse
		tx.Add(r.svc.Mutation.Property.Policies.CreateOp(*g.policy, &res))
	}
	var propRes pschema.InsertPropertyPropertiesResponse
	tx.Add(r.svc.Mutation.Property.Properties.CreateOp(g.property, &propRes))
	mediaRes := make([]pschema.InsertPropertyMediasResponse, len(g.medias))
	for i := range g.medias {
		g.medias[i].PropertyId = id
		tx.Add(r.svc.Mutation.Property.Medias.CreateOp(g.medias[i], &mediaRes[i]))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetProperty(ctx, name)
}

func (r *PropertyRepository) GetProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.Properties.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	parts, _, err := r.fetchPropertyParts(ctx, res)
	if err != nil {
		return nil, err
	}
	return propertyFromParts(parts), nil
}

func (r *PropertyRepository) ListProperties(ctx context.Context, in repox.ListInput) ([]*propertypbv1.Property, string, error) {
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[pschema.PropertyProperties](property.PropertyFilterSpec, r.svc.Query.Property.Properties).
		List(ctx, fin)
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Property, 0, len(rows))
	for i := range rows {
		parts, _, err := r.fetchPropertyParts(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, propertyFromParts(parts))
	}
	return items, next, nil
}

// fetchPropertyParts loads a property row's address, policy, media, and child
// unit names, and returns the ids of the deletable child rows.
func (r *PropertyRepository) fetchPropertyParts(ctx context.Context, res *pschema.PropertyProperties) (propertyParts, propertyRefs, error) {
	p := propertyParts{res: res}
	refs := propertyRefs{addressID: res.AddressId, policyID: res.PolicyId}

	if res.AddressId != nil {
		a, err := r.svc.Query.Common.PostalAddress.Get(ctx, *res.AddressId)
		if err != nil {
			return propertyParts{}, propertyRefs{}, mapHasuraErr(err)
		}
		p.address = a
	}
	if res.PolicyId != nil {
		pol, err := r.svc.Query.Property.Policies.Get(ctx, *res.PolicyId)
		if err != nil {
			return propertyParts{}, propertyRefs{}, mapHasuraErr(err)
		}
		p.policy = pol
	}
	medias, err := r.svc.Query.Property.Medias.List(ctx, mediasList().Where(mediasql.PropertyId.Eq(res.Id)))
	if err != nil {
		return propertyParts{}, propertyRefs{}, mapHasuraErr(err)
	}
	p.medias = medias
	for i := range medias {
		refs.mediaIDs = append(refs.mediaIDs, medias[i].Id)
	}
	units, err := r.svc.Query.Property.Units.List(ctx, unitsList().Where(unitsql.PropertyId.Eq(res.Id)))
	if err != nil {
		return propertyParts{}, propertyRefs{}, mapHasuraErr(err)
	}
	for i := range units {
		p.unitNames = append(p.unitNames, units[i].Name)
	}
	licences, err := r.svc.Query.Property.Licences.List(ctx, licencesql.List().Where(licencesql.PropertyId.Eq(res.Id)))
	if err != nil {
		return propertyParts{}, propertyRefs{}, mapHasuraErr(err)
	}
	for i := range licences {
		p.licenceNames = append(p.licenceNames, licences[i].Name)
	}
	return p, refs, nil
}

// --- Unit --------------------------------------------------------------------

func (r *PropertyRepository) CreateUnit(ctx context.Context, parent string, u *propertypbv1.Unit) (*propertypbv1.Unit, error) {
	propertyID, id, name, err := types.ResolveUnitName(parent, u.GetName())
	if err != nil {
		return nil, err
	}
	g := buildUnitGraph(u, propertyID, time.Now().UTC())
	g.unit.Id = id
	g.unit.Name = name
	g.unit.Etag = ulid.GenerateString()

	tx := r.svc.Mutation.Tx()
	queueUnitInserts(tx, r, g, id)
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetUnit(ctx, name)
}

func (r *PropertyRepository) GetUnit(ctx context.Context, name string) (*propertypbv1.Unit, error) {
	id, err := types.UnitID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.Units.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	parts, _, err := r.fetchUnitParts(ctx, res)
	if err != nil {
		return nil, err
	}
	return unitFromParts(parts), nil
}

func (r *PropertyRepository) ListUnits(ctx context.Context, parent string, in repox.ListInput) ([]*propertypbv1.Unit, string, error) {
	propertyID, err := types.PropertyID(parent)
	if err != nil {
		return nil, "", err
	}
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[pschema.PropertyUnits](property.UnitFilterSpec, r.svc.Query.Property.Units).
		Scope(unitsql.PropertyId.Eq(propertyID)).
		List(ctx, fin)
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Unit, 0, len(rows))
	for i := range rows {
		parts, _, err := r.fetchUnitParts(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, unitFromParts(parts))
	}
	return items, next, nil
}

// fetchUnitParts loads a unit's price and pricing children (rate overrides, LOS
// discounts, fees, taxes), media, and applicable-promo-code rows, resolving the
// Money/DateRange value-objects each references, and returns the deletable ids.
func (r *PropertyRepository) fetchUnitParts(ctx context.Context, res *pschema.PropertyUnits) (unitParts, unitRefs, error) {
	p := unitParts{res: res, moneyByID: map[string]*commonschema.CommonMoneys{}, dateByID: map[string]*sharedschema.SharedDateRanges{}}
	var refs unitRefs

	if res.PriceId != nil {
		m, err := r.money(ctx, *res.PriceId)
		if err != nil {
			return unitParts{}, unitRefs{}, err
		}
		p.price = m
		refs.moneyIDs = append(refs.moneyIDs, *res.PriceId)
	}

	ro, err := r.svc.Query.Property.RateOverrides.List(ctx, rateOverridesList().Where(rateoverridesql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, mapHasuraErr(err)
	}
	p.rateOverrides = ro
	for i := range ro {
		refs.rateIDs = append(refs.rateIDs, ro[i].Id)
		if ro[i].PriceId != "" {
			if m, err := r.money(ctx, ro[i].PriceId); err != nil {
				return unitParts{}, unitRefs{}, err
			} else {
				p.moneyByID[ro[i].PriceId] = m
				refs.moneyIDs = append(refs.moneyIDs, ro[i].PriceId)
			}
		}
		if ro[i].DateRangeId != nil {
			d, err := r.svc.Query.Shared.DateRanges.Get(ctx, *ro[i].DateRangeId)
			if err != nil {
				return unitParts{}, unitRefs{}, mapHasuraErr(err)
			}
			p.dateByID[*ro[i].DateRangeId] = d
			refs.dateIDs = append(refs.dateIDs, *ro[i].DateRangeId)
		}
	}

	ld, err := r.svc.Query.Property.LosDiscounts.List(ctx, losDiscountsList().Where(losdiscountsql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, mapHasuraErr(err)
	}
	p.losDiscounts = ld
	for i := range ld {
		refs.losIDs = append(refs.losIDs, ld[i].Id)
		if ld[i].AmountOffId != nil {
			if m, err := r.money(ctx, *ld[i].AmountOffId); err != nil {
				return unitParts{}, unitRefs{}, err
			} else {
				p.moneyByID[*ld[i].AmountOffId] = m
				refs.moneyIDs = append(refs.moneyIDs, *ld[i].AmountOffId)
			}
		}
	}

	fees, err := r.svc.Query.Property.Fees.List(ctx, feesList().Where(feesql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, mapHasuraErr(err)
	}
	p.fees = fees
	for i := range fees {
		refs.feeIDs = append(refs.feeIDs, fees[i].Id)
		if fees[i].AmountId != nil {
			if m, err := r.money(ctx, *fees[i].AmountId); err != nil {
				return unitParts{}, unitRefs{}, err
			} else {
				p.moneyByID[*fees[i].AmountId] = m
				refs.moneyIDs = append(refs.moneyIDs, *fees[i].AmountId)
			}
		}
	}

	taxes, err := r.svc.Query.Property.Taxes.List(ctx, taxesList().Where(taxesql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, mapHasuraErr(err)
	}
	p.taxes = taxes
	for i := range taxes {
		refs.taxIDs = append(refs.taxIDs, taxes[i].Id)
	}

	medias, err := r.svc.Query.Property.UnitMedias.List(ctx, unitMediasList().Where(unitmediasql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, mapHasuraErr(err)
	}
	p.medias = medias
	for i := range medias {
		refs.mediaIDs = append(refs.mediaIDs, medias[i].Id)
	}

	codes, err := r.svc.Query.Property.UnitApplicablePromoCodes.List(ctx, unitPromoCodesList().Where(unitapplicablepromocodesql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, mapHasuraErr(err)
	}
	p.promoCodes = codes
	for i := range codes {
		refs.promoIDs = append(refs.promoIDs, codes[i].Id)
	}

	return p, refs, nil
}

func (r *PropertyRepository) money(ctx context.Context, id string) (*commonschema.CommonMoneys, error) {
	m, err := r.svc.Query.Common.Moneys.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	return m, nil
}
