// Unit reads/creates and their row hydration.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/feesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/losdiscountsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/rateoverridesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/taxesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitapplicablepromocodesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitmediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

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
		return nil, dbutil.MapHasuraErr(err)
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
		return nil, dbutil.MapHasuraErr(err)
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
	rows, next, err := filterx.Hasura(property.UnitFilterSpec, r.svc.Query.Property.Units).
		Scope(unitsql.PropertyId.Eq(propertyID)).
		List(ctx, fin)
	if err != nil {
		return nil, "", dbutil.MapHasuraErr(repox.MapFilterxErr(err))
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
func (r *PropertyRepository) fetchUnitParts(ctx context.Context, res *unitsql.PropertyUnits) (unitParts, unitRefs, error) {
	p := unitParts{res: res, moneyByID: map[string]*moneysql.CommonMoneys{}, dateByID: map[string]*daterangesql.SharedDateRanges{}}
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
		return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
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
				return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
			}
			p.dateByID[*ro[i].DateRangeId] = d
			refs.dateIDs = append(refs.dateIDs, *ro[i].DateRangeId)
		}
	}

	ld, err := r.svc.Query.Property.LosDiscounts.List(ctx, losDiscountsList().Where(losdiscountsql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
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
		return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
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
		return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
	}
	p.taxes = taxes
	for i := range taxes {
		refs.taxIDs = append(refs.taxIDs, taxes[i].Id)
	}

	medias, err := r.svc.Query.Property.UnitMedias.List(ctx, unitMediasList().Where(unitmediasql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
	}
	p.medias = medias
	for i := range medias {
		refs.mediaIDs = append(refs.mediaIDs, medias[i].Id)
	}

	codes, err := r.svc.Query.Property.UnitApplicablePromoCodes.List(ctx, unitPromoCodesList().Where(unitapplicablepromocodesql.UnitId.Eq(res.Id)))
	if err != nil {
		return unitParts{}, unitRefs{}, dbutil.MapHasuraErr(err)
	}
	p.promoCodes = codes
	for i := range codes {
		refs.promoIDs = append(refs.promoIDs, codes[i].Id)
	}

	return p, refs, nil
}
