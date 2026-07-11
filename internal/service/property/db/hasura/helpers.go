package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/feesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/losdiscountsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/mediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/rateoverridesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/taxesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitapplicablepromocodesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitmediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	"github.com/the-protobuf-project/runtime-go/network/runtime"
)

// List-request constructors, wrapped so the repository reads uniformly and the
// per-entity ql imports stay in one place.
func unitsList() *unitsql.ListRequest                 { return unitsql.List() }
func mediasList() *mediasql.ListRequest               { return mediasql.List() }
func unitMediasList() *unitmediasql.ListRequest       { return unitmediasql.List() }
func rateOverridesList() *rateoverridesql.ListRequest { return rateoverridesql.List() }
func losDiscountsList() *losdiscountsql.ListRequest   { return losdiscountsql.List() }
func feesList() *feesql.ListRequest                   { return feesql.List() }
func taxesList() *taxesql.ListRequest                 { return taxesql.List() }
func unitPromoCodesList() *unitapplicablepromocodesql.ListRequest {
	return unitapplicablepromocodesql.List()
}

// propertyRefs captures the ids of a property's belongs-to children (address,
// policy) and its media rows so a writer can delete the superseded ones.
type propertyRefs struct {
	addressID *string
	policyID  *string
	mediaIDs  []string
}

// unitRefs captures the ids of a unit's value-objects (Money, DateRange) and its
// pricing/media/promo child rows, for deletion on update/delete.
type unitRefs struct {
	moneyIDs []string
	dateIDs  []string
	rateIDs  []string
	losIDs   []string
	feeIDs   []string
	taxIDs   []string
	mediaIDs []string
	promoIDs []string
}

// queueUnitInserts appends a unit graph's inserts to tx in foreign-key order: the
// price Money (referenced by the unit) first, then the unit, then the value-object
// Money/DateRange rows, then the pricing/media/promo children (stamped with the
// unit id) that reference them.
func queueUnitInserts(tx *runtime.Tx, r *PropertyRepository, g *unitGraph, unitID string) {
	if g.price != nil {
		var res moneysql.InsertCommonMoneysResponse
		tx.Add(r.svc.Mutation.Common.Moneys.CreateOp(*g.price, &res))
	}
	var uRes unitsql.InsertPropertyUnitsResponse
	tx.Add(r.svc.Mutation.Property.Units.CreateOp(g.unit, &uRes))
	queueUnitChildInserts(tx, r, g, unitID)
}

// queueUnitChildInserts appends only the value-object and child-row inserts (not
// the price Money or the unit row) — used by update, where the price is inserted
// separately and the unit is patched rather than created.
func queueUnitChildInserts(tx *runtime.Tx, r *PropertyRepository, g *unitGraph, unitID string) {
	moneyRes := make([]moneysql.InsertCommonMoneysResponse, len(g.moneys))
	for i := range g.moneys {
		tx.Add(r.svc.Mutation.Common.Moneys.CreateOp(g.moneys[i], &moneyRes[i]))
	}
	dateRes := make([]daterangesql.InsertSharedDateRangesResponse, len(g.dates))
	for i := range g.dates {
		tx.Add(r.svc.Mutation.Shared.DateRanges.CreateOp(g.dates[i], &dateRes[i]))
	}
	roRes := make([]rateoverridesql.InsertPropertyRateOverridesResponse, len(g.rateOverrides))
	for i := range g.rateOverrides {
		g.rateOverrides[i].UnitId = unitID
		tx.Add(r.svc.Mutation.Property.RateOverrides.CreateOp(g.rateOverrides[i], &roRes[i]))
	}
	ldRes := make([]losdiscountsql.InsertPropertyLosDiscountsResponse, len(g.losDiscounts))
	for i := range g.losDiscounts {
		g.losDiscounts[i].UnitId = unitID
		tx.Add(r.svc.Mutation.Property.LosDiscounts.CreateOp(g.losDiscounts[i], &ldRes[i]))
	}
	feeRes := make([]feesql.InsertPropertyFeesResponse, len(g.fees))
	for i := range g.fees {
		g.fees[i].UnitId = unitID
		tx.Add(r.svc.Mutation.Property.Fees.CreateOp(g.fees[i], &feeRes[i]))
	}
	taxRes := make([]taxesql.InsertPropertyTaxesResponse, len(g.taxes))
	for i := range g.taxes {
		g.taxes[i].UnitId = unitID
		tx.Add(r.svc.Mutation.Property.Taxes.CreateOp(g.taxes[i], &taxRes[i]))
	}
	mediaRes := make([]unitmediasql.InsertPropertyUnitMediasResponse, len(g.medias))
	for i := range g.medias {
		g.medias[i].UnitId = unitID
		tx.Add(r.svc.Mutation.Property.UnitMedias.CreateOp(g.medias[i], &mediaRes[i]))
	}
	codeRes := make([]unitapplicablepromocodesql.InsertPropertyUnitApplicablePromoCodesResponse, len(g.promoCodes))
	for i := range g.promoCodes {
		g.promoCodes[i].UnitId = unitID
		tx.Add(r.svc.Mutation.Property.UnitApplicablePromoCodes.CreateOp(g.promoCodes[i], &codeRes[i]))
	}
}
