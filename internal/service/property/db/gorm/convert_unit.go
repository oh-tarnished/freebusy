package gorm

import (
	"time"

	"github.com/lib/pq"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/protobuf/types/known/durationpb"
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

// buildUnitGraph turns a proto Unit into its row graph under propertyID, minting
// a fresh ULID for every row. The unit's identity (ID/Name/Etag) and every
// child's unit_id FK are stamped by the repository.
func buildUnitGraph(u *propertypbv1.Unit, propertyID string) *unitGraph {
	g := &unitGraph{}
	state := property.UnitStateActive
	g.unit = &property.Unit{
		DisplayName:  u.GetDisplayName(),
		Description:  strOrNil(u.GetDescription()),
		Type:         unitTypeToModel(u.GetType()),
		BookingMode:  bookingModeToModel(u.GetBookingMode()),
		Capacity:     nilIfZeroInt32(u.GetCapacity()),
		MaxOccupancy: nilIfZeroInt32(u.GetMaxOccupancy()),
		TimeZone:     u.GetTimeZone(),
		PricingUnit:  pricingUnitToModel(u.GetPricingUnit()),
		Duration:     durationToStr(u.GetDuration()),
		Tags:         pq.StringArray(u.GetTags()),
		Attributes:   structToJSON(u.GetAttributes()),
		State:        &state,
		PropertyID:   propertyID,
	}
	if p := moneyToModel(u.GetPrice()); p != nil {
		g.price = p
		g.unit.PriceID = &p.ID
	}

	for _, ro := range u.GetRateOverrides() {
		row := &property.RateOverride{ID: ulid.GenerateString(), Weekdays: weekdaysToStr(ro.GetWeekdays())}
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
		row := &property.LosDiscount{ID: ulid.GenerateString(), MinNights: ld.GetMinNights(), PercentOff: nilIfZeroInt32(ld.GetPercentOff())}
		if amt := moneyToModel(ld.GetAmountOff()); amt != nil {
			g.moneys = append(g.moneys, amt)
			row.AmountOffID = &amt.ID
		}
		g.losDiscounts = append(g.losDiscounts, row)
	}
	for _, f := range u.GetFees() {
		row := &property.Fee{
			ID:          ulid.GenerateString(),
			Code:        f.GetCode(),
			DisplayName: strOrNil(f.GetDisplayName()),
			Percent:     nilIfZeroInt32(f.GetPercent()),
			PricingUnit: pricingUnitToModel(f.GetPricingUnit()),
			Taxable:     ptr(f.GetTaxable()),
		}
		if amt := moneyToModel(f.GetAmount()); amt != nil {
			g.moneys = append(g.moneys, amt)
			row.AmountID = &amt.ID
		}
		g.fees = append(g.fees, row)
	}
	for _, t := range u.GetTaxes() {
		g.taxes = append(g.taxes, &property.Tax{
			ID:          ulid.GenerateString(),
			Code:        t.GetCode(),
			DisplayName: strOrNil(t.GetDisplayName()),
			Percent:     t.GetPercent(),
		})
	}
	for _, m := range u.GetMedia() {
		g.medias = append(g.medias, unitMediaToModel(m))
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
// associations (price, rate overrides, LOS discounts, fees, taxes, media, and
// applicable-promo-code join rows).
func unitFromModel(m *property.Unit) *propertypbv1.Unit {
	u := &propertypbv1.Unit{
		Name:         m.Name,
		DisplayName:  m.DisplayName,
		Description:  deref(m.Description),
		Type:         unitTypeFromModel(m.Type),
		BookingMode:  bookingModeFromModel(m.BookingMode),
		Capacity:     deref(m.Capacity),
		MaxOccupancy: deref(m.MaxOccupancy),
		TimeZone:     m.TimeZone,
		Price:        moneyFromModel(m.Price),
		PricingUnit:  pricingUnitFromModel(m.PricingUnit),
		Duration:     durationFromStr(m.Duration),
		Tags:         []string(m.Tags),
		Attributes:   jsonToStruct(m.Attributes),
		State:        unitStateFromModel(m.State),
		CreateTime:   timeToTS(&m.CreateTime),
		UpdateTime:   timeToTS(&m.UpdateTime),
		Etag:         deref(m.Etag),
	}
	for i := range m.RateOverrides {
		u.RateOverrides = append(u.RateOverrides, rateOverrideFromModel(&m.RateOverrides[i]))
	}
	for i := range m.LosDiscounts {
		u.LosDiscounts = append(u.LosDiscounts, losDiscountFromModel(&m.LosDiscounts[i]))
	}
	for i := range m.Fees {
		u.Fees = append(u.Fees, feeFromModel(&m.Fees[i]))
	}
	for i := range m.Taxes {
		u.Taxes = append(u.Taxes, taxFromModel(&m.Taxes[i]))
	}
	for i := range m.UnitMedias {
		u.Media = append(u.Media, unitMediaFromModel(&m.UnitMedias[i]))
	}
	for i := range m.UnitApplicablePromoCodes {
		u.ApplicablePromoCodes = append(u.ApplicablePromoCodes, m.UnitApplicablePromoCodes[i].PromoCodeID)
	}
	return u
}

func rateOverrideFromModel(r *property.RateOverride) *propertypbv1.RateOverride {
	return &propertypbv1.RateOverride{
		DateRange: dateRangeFromModel(r.DateRange),
		Weekdays:  weekdaysFromStr(r.Weekdays),
		Price:     moneyFromModel(r.Price),
	}
}

func losDiscountFromModel(r *property.LosDiscount) *propertypbv1.LosDiscount {
	return &propertypbv1.LosDiscount{
		MinNights:  r.MinNights,
		PercentOff: deref(r.PercentOff),
		AmountOff:  moneyFromModel(r.AmountOff),
	}
}

func feeFromModel(r *property.Fee) *propertypbv1.Fee {
	return &propertypbv1.Fee{
		Code:        r.Code,
		DisplayName: deref(r.DisplayName),
		Amount:      moneyFromModel(r.Amount),
		Percent:     deref(r.Percent),
		PricingUnit: pricingUnitFromModel(r.PricingUnit),
		Taxable:     deref(r.Taxable),
	}
}

func taxFromModel(r *property.Tax) *propertypbv1.Tax {
	return &propertypbv1.Tax{
		Code:        r.Code,
		DisplayName: deref(r.DisplayName),
		Percent:     r.Percent,
	}
}

func unitMediaToModel(m *propertypbv1.UnitMedia) *property.UnitMedia {
	return &property.UnitMedia{
		ID:          ulid.GenerateString(),
		URI:         m.GetUri(),
		Type:        mediaTypeToModel(m.GetType()),
		Title:       strOrNil(m.GetTitle()),
		Description: strOrNil(m.GetDescription()),
		MimeType:    strOrNil(m.GetMimeType()),
		SortOrder:   ptr(m.GetSortOrder()),
		Primary:     ptr(m.GetPrimary()),
	}
}

func unitMediaFromModel(m *property.UnitMedia) *propertypbv1.UnitMedia {
	return &propertypbv1.UnitMedia{
		Uri:         m.URI,
		Type:        mediaTypeFromModel(m.Type),
		Title:       deref(m.Title),
		Description: deref(m.Description),
		MimeType:    deref(m.MimeType),
		SortOrder:   deref(m.SortOrder),
		Primary:     deref(m.Primary),
	}
}

func unitStateFromModel(s *property.UnitState) propertypbv1.UnitState {
	if s == nil {
		return propertypbv1.UnitState_UNIT_STATE_UNSPECIFIED
	}
	return propertypbv1.UnitState(propertypbv1.UnitState_value["UNIT_STATE_"+string(*s)])
}

func nilIfZeroInt32(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

// durationToStr serializes a proto Duration into the string column the ORM
// generates for it (Go duration syntax, e.g. "30m0s"), round-tripping exactly.
func durationToStr(d *durationpb.Duration) *string {
	if d == nil {
		return nil
	}
	return ptr(d.AsDuration().String())
}

func durationFromStr(s *string) *durationpb.Duration {
	if s == nil || *s == "" {
		return nil
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
}
