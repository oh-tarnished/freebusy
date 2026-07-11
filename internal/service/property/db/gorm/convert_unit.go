package gorm

import (
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// unitGraph is the set of rows a single Unit materializes into: the unit row,
// its belongs-to price Money (created before it), and its has-many pricing
// children — rate overrides (each with its own Money + DateRange), LOS discounts
// and fees (each with a Money), taxes, media, and applicable-promo-code join
// rows (created after it, since they carry the unit_id FK).
type unitGraph struct {
	unit          *property.Unit
	price         *common.Money
	moneys        []*common.Money // child value-object Money rows (rate/los/fee)
	dates         []*shared.DateRange
	rateOverrides []*property.RateOverride
	losDiscounts  []*property.LosDiscount
	fees          []*property.Fee
	taxes         []*property.Tax
	medias        []*property.UnitMedia
	promoCodes    []*property.UnitApplicablePromoCodes
}

// buildUnitGraph turns a proto Unit into its row graph under propertyID. The
// generated converters carry each row's field mass; this wires fresh ULIDs and
// the value-object foreign keys. The unit's identity (ID/Name/Etag) and every
// child's unit_id FK are stamped by the repository.
func buildUnitGraph(u *propertypbv1.Unit, propertyID string) *unitGraph {
	g := &unitGraph{}
	g.unit = property.UnitFromProto(u)
	g.unit.Name = "" // identity is the repository's
	g.unit.Etag = nil
	g.unit.PropertyID = propertyID
	if p := moneyToModel(u.GetPrice()); p != nil {
		g.price = p
		g.unit.PriceID = &p.ID
	}

	for _, ro := range u.GetRateOverrides() {
		row := property.RateOverrideFromProto(ro)
		row.ID = ulid.GenerateString()
		if dr := dateRangeToModel(ro.GetDateRange()); dr != nil {
			g.dates = append(g.dates, dr)
			row.DateRangeID = &dr.ID
		}
		if price := moneyToModel(ro.GetPrice()); price != nil {
			g.moneys = append(g.moneys, price)
			row.PriceID = price.ID
		}
		g.rateOverrides = append(g.rateOverrides, row)
	}
	for _, ld := range u.GetLosDiscounts() {
		row := property.LosDiscountFromProto(ld)
		row.ID = ulid.GenerateString()
		if amt := moneyToModel(ld.GetAmountOff()); amt != nil {
			g.moneys = append(g.moneys, amt)
			row.AmountOffID = &amt.ID
		}
		g.losDiscounts = append(g.losDiscounts, row)
	}
	for _, f := range u.GetFees() {
		row := property.FeeFromProto(f)
		row.ID = ulid.GenerateString()
		if amt := moneyToModel(f.GetAmount()); amt != nil {
			g.moneys = append(g.moneys, amt)
			row.AmountID = &amt.ID
		}
		g.fees = append(g.fees, row)
	}
	for _, t := range u.GetTaxes() {
		row := property.TaxFromProto(t)
		row.ID = ulid.GenerateString()
		g.taxes = append(g.taxes, row)
	}
	for _, m := range u.GetMedia() {
		row := property.UnitMediaFromProto(m)
		row.ID = ulid.GenerateString()
		g.medias = append(g.medias, row)
	}
	for _, name := range u.GetApplicablePromoCodes() {
		g.promoCodes = append(g.promoCodes, &property.UnitApplicablePromoCodes{
			ID:          ulid.GenerateString(),
			PromoCodeID: name,
		})
	}
	return g
}

// unitFromModel assembles the protobuf Unit from a stored row and its preloaded
// associations. The generated converter covers the flat fields and the price
// Money; the pricing children, media, and promo-code join rows are layered on
// through their own generated converters.
func unitFromModel(m *property.Unit) *propertypbv1.Unit {
	u := property.UnitToProto(m)
	for i := range m.RateOverrides {
		u.RateOverrides = append(u.RateOverrides, property.RateOverrideToProto(&m.RateOverrides[i]))
	}
	for i := range m.LosDiscounts {
		u.LosDiscounts = append(u.LosDiscounts, property.LosDiscountToProto(&m.LosDiscounts[i]))
	}
	for i := range m.Fees {
		u.Fees = append(u.Fees, property.FeeToProto(&m.Fees[i]))
	}
	for i := range m.Taxes {
		u.Taxes = append(u.Taxes, property.TaxToProto(&m.Taxes[i]))
	}
	for i := range m.UnitMedias {
		u.Media = append(u.Media, property.UnitMediaToProto(&m.UnitMedias[i]))
	}
	for i := range m.UnitApplicablePromoCodes {
		u.ApplicablePromoCodes = append(u.ApplicablePromoCodes, m.UnitApplicablePromoCodes[i].PromoCodeID)
	}
	return u
}
