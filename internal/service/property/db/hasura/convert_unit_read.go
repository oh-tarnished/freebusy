// The unit read side: rows back to the Unit proto.
package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
)

type unitParts struct {
	res           *pschema.PropertyUnits
	price         *commonschema.CommonMoneys
	rateOverrides []pschema.PropertyRateOverrides
	losDiscounts  []pschema.PropertyLosDiscounts
	fees          []pschema.PropertyFees
	taxes         []pschema.PropertyTaxes
	medias        []pschema.PropertyUnitMedias
	promoCodes    []pschema.PropertyUnitApplicablePromoCodes
	moneyByID     map[string]*commonschema.CommonMoneys
	dateByID      map[string]*sharedschema.SharedDateRanges
	licenceNames  []string
}

func unitFromParts(p unitParts) *propertypbv1.Unit {
	res := p.res
	out := &propertypbv1.Unit{
		Name:         res.Name,
		DisplayName:  res.DisplayName,
		Description:  repox.Deref(res.Description),
		Type:         unitTypeFromStr(res.Type),
		BookingMode:  bookingModeFromStr(res.BookingMode),
		Capacity:     repox.Deref(res.Capacity),
		MaxOccupancy: repox.Deref(res.MaxOccupancy),
		TimeZone:     res.TimeZone,
		Price:        moneyFromModel(p.price),
		PricingUnit:  pricingUnitFromStr(res.PricingUnit),
		Duration:     strToDuration(res.Duration),
		Tags:         strPtrsToSlice(res.Tags),
		Attributes:   jsonToStruct(jsonBytes(res.Attributes)),
		State:        unitStateFromStr(res.State),
		CreateTime:   strToTS(res.CreateTime),
		UpdateTime:   strToTS(res.UpdateTime),
		Etag:         repox.Deref(res.Etag),
	}
	for i := range p.rateOverrides {
		r := &p.rateOverrides[i]
		ro := &propertypbv1.RateOverride{
			Weekdays: weekdaysFromStr(r.Weekdays),
			Price:    moneyFromModel(p.moneyByID[r.PriceId]),
		}
		if r.DateRangeId != nil {
			ro.DateRange = dateRangeFromModel(p.dateByID[*r.DateRangeId])
		}
		out.RateOverrides = append(out.RateOverrides, ro)
	}
	for i := range p.losDiscounts {
		l := &p.losDiscounts[i]
		ld := &propertypbv1.LosDiscount{MinNights: l.MinNights, PercentOff: repox.Deref(l.PercentOff)}
		if l.AmountOffId != nil {
			ld.AmountOff = moneyFromModel(p.moneyByID[*l.AmountOffId])
		}
		out.LosDiscounts = append(out.LosDiscounts, ld)
	}
	for i := range p.fees {
		f := &p.fees[i]
		fee := &propertypbv1.Fee{
			Code:        f.Code,
			DisplayName: repox.Deref(f.DisplayName),
			Percent:     repox.Deref(f.Percent),
			PricingUnit: pricingUnitFromStr(f.PricingUnit),
			Taxable:     repox.Deref(f.Taxable),
		}
		if f.AmountId != nil {
			fee.Amount = moneyFromModel(p.moneyByID[*f.AmountId])
		}
		out.Fees = append(out.Fees, fee)
	}
	for i := range p.taxes {
		out.Taxes = append(out.Taxes, &propertypbv1.Tax{
			Code:        p.taxes[i].Code,
			DisplayName: repox.Deref(p.taxes[i].DisplayName),
			Percent:     p.taxes[i].Percent,
		})
	}
	for i := range p.medias {
		out.Media = append(out.Media, unitMediaFromModel(&p.medias[i]))
	}
	for i := range p.promoCodes {
		out.ApplicablePromoCodes = append(out.ApplicablePromoCodes, p.promoCodes[i].PromoCodeId)
	}
	return out
}

func unitMediaFromModel(m *pschema.PropertyUnitMedias) *propertypbv1.UnitMedia {
	return &propertypbv1.UnitMedia{
		Uri:         m.Uri,
		Type:        mediaTypeFromStr(m.Type),
		Title:       repox.Deref(m.Title),
		Description: repox.Deref(m.Description),
		MimeType:    repox.Deref(m.MimeType),
		SortOrder:   repox.Deref(m.SortOrder),
		Primary:     repox.Deref(m.Primary),
	}
}

func unitStateFromStr(s *string) propertypbv1.UnitState {
	if s == nil || *s == "" {
		return propertypbv1.UnitState_UNIT_STATE_UNSPECIFIED
	}
	return propertypbv1.UnitState(propertypbv1.UnitState_value["UNIT_STATE_"+*s])
}
