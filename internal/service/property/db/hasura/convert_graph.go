package hasura

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/postaladdressql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/feesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/losdiscountsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/mediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/policiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/rateoverridesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/taxesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitapplicablepromocodesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitmediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const stateActive = "ACTIVE"

// --- Property graph ----------------------------------------------------------

type propertyGraph struct {
	property propertiesql.CreateInput
	address  *postaladdressql.CreateInput
	policy   *policiesql.CreateInput
	medias   []mediasql.CreateInput
}

func buildPropertyGraph(p *propertypbv1.Property, now time.Time) *propertyGraph {
	g := &propertyGraph{}
	nowStr := tsToStr(timestamppb.New(now))
	g.property = propertiesql.CreateInput{
		Organisation: lastSegment(p.GetOrganisation()),
		DisplayName:  p.GetDisplayName(),
		Description:  p.GetDescription(),
		TimeZone:     p.GetTimeZone(),
		Tags:         strSliceToPtrs(p.GetTags()),
		Attributes:   structToJSON(p.GetAttributes()),
		State:        stateActive,
		CreateTime:   nowStr,
		UpdateTime:   nowStr,
	}
	if a := p.GetAddress(); a != nil {
		aID := ulid.GenerateString()
		ci := addressInput(aID, a)
		g.address = &ci
		g.property.AddressId = aID
	}
	if pol := p.GetPolicy(); pol != nil {
		pID := ulid.GenerateString()
		ci := policiesql.CreateInput{
			Id:           pID,
			CheckinTime:  todToStr(pol.GetCheckinTime()),
			CheckoutTime: todToStr(pol.GetCheckoutTime()),
			HouseRules:   strSliceToPtrs(pol.GetHouseRules()),
			Notes:        pol.GetNotes(),
		}
		g.policy = &ci
		g.property.PolicyId = pID
	}
	for _, m := range p.GetMedia() {
		g.medias = append(g.medias, mediasql.CreateInput{
			Id:          ulid.GenerateString(),
			Uri:         m.GetUri(),
			Type:        mediaTypeToStr(m.GetType()),
			Title:       m.GetTitle(),
			Description: m.GetDescription(),
			MimeType:    m.GetMimeType(),
			SortOrder:   m.GetSortOrder(),
			Primary:     m.GetPrimary(),
		})
	}
	return g
}

type propertyParts struct {
	res          *pschema.PropertyProperties
	address      *commonschema.CommonPostalAddress
	policy       *pschema.PropertyPolicies
	medias       []pschema.PropertyMedias
	unitNames    []string
	licenceNames []string
}

func propertyFromParts(p propertyParts) *propertypbv1.Property {
	res := p.res
	out := &propertypbv1.Property{
		Name:         res.Name,
		Organisation: orgName(res.Organisation),
		DisplayName:  res.DisplayName,
		Description:  deref(res.Description),
		Address:      addressFromModel(p.address),
		TimeZone:     res.TimeZone,
		Policy:       policyFromModel(p.policy),
		Tags:         strPtrsToSlice(res.Tags),
		Attributes:   jsonToStruct(jsonBytes(res.Attributes)),
		State:        propertyStateFromStr(res.State),
		CreateTime:   strToTS(res.CreateTime),
		UpdateTime:   strToTS(res.UpdateTime),
		Etag:         deref(res.Etag),
		Units:        p.unitNames,
		Licences:     p.licenceNames,
	}
	for i := range p.medias {
		out.Media = append(out.Media, mediaFromModel(&p.medias[i]))
	}
	return out
}

func policyFromModel(p *pschema.PropertyPolicies) *propertypbv1.Policy {
	if p == nil {
		return nil
	}
	return &propertypbv1.Policy{
		CheckinTime:  strToTOD(p.CheckinTime),
		CheckoutTime: strToTOD(p.CheckoutTime),
		HouseRules:   strPtrsToSlice(p.HouseRules),
		Notes:        deref(p.Notes),
	}
}

func mediaFromModel(m *pschema.PropertyMedias) *propertypbv1.Media {
	return &propertypbv1.Media{
		Uri:         m.Uri,
		Type:        mediaTypeFromStr(m.Type),
		Title:       deref(m.Title),
		Description: deref(m.Description),
		MimeType:    deref(m.MimeType),
		SortOrder:   deref(m.SortOrder),
		Primary:     deref(m.Primary),
	}
}

func propertyStateFromStr(s *string) propertypbv1.PropertyState {
	if s == nil || *s == "" {
		return propertypbv1.PropertyState_PROPERTY_STATE_UNSPECIFIED
	}
	return propertypbv1.PropertyState(propertypbv1.PropertyState_value["PROPERTY_STATE_"+*s])
}

// --- Unit graph --------------------------------------------------------------

type unitGraph struct {
	unit          unitsql.CreateInput
	price         *moneysql.CreateInput
	moneys        []moneysql.CreateInput
	dates         []daterangesql.CreateInput
	rateOverrides []rateoverridesql.CreateInput
	losDiscounts  []losdiscountsql.CreateInput
	fees          []feesql.CreateInput
	taxes         []taxesql.CreateInput
	medias        []unitmediasql.CreateInput
	promoCodes    []unitapplicablepromocodesql.CreateInput
}

func buildUnitGraph(u *propertypbv1.Unit, propertyID string, now time.Time) *unitGraph {
	g := &unitGraph{}
	nowStr := tsToStr(timestamppb.New(now))
	g.unit = unitsql.CreateInput{
		DisplayName:  u.GetDisplayName(),
		Description:  u.GetDescription(),
		Type:         unitTypeToStr(u.GetType()),
		BookingMode:  bookingModeToStr(u.GetBookingMode()),
		Capacity:     u.GetCapacity(),
		MaxOccupancy: u.GetMaxOccupancy(),
		TimeZone:     u.GetTimeZone(),
		PricingUnit:  pricingUnitToStr(u.GetPricingUnit()),
		Duration:     durationToStr(u.GetDuration()),
		Tags:         strSliceToPtrs(u.GetTags()),
		Attributes:   structToJSON(u.GetAttributes()),
		State:        stateActive,
		PropertyId:   propertyID,
		CreateTime:   nowStr,
		UpdateTime:   nowStr,
	}
	if p := u.GetPrice(); p != nil {
		mID := ulid.GenerateString()
		ci := moneyInput(mID, p)
		g.price = &ci
		g.unit.PriceId = mID
	}
	for _, ro := range u.GetRateOverrides() {
		row := rateoverridesql.CreateInput{Id: ulid.GenerateString(), Weekdays: weekdaysToStr(ro.GetWeekdays())}
		if dr := ro.GetDateRange(); dr != nil {
			dID := ulid.GenerateString()
			g.dates = append(g.dates, dateRangeInput(dID, dr))
			row.DateRangeId = dID
		}
		if p := ro.GetPrice(); p != nil {
			mID := ulid.GenerateString()
			g.moneys = append(g.moneys, moneyInput(mID, p))
			row.PriceId = mID
		}
		g.rateOverrides = append(g.rateOverrides, row)
	}
	for _, ld := range u.GetLosDiscounts() {
		row := losdiscountsql.CreateInput{Id: ulid.GenerateString(), MinNights: ld.GetMinNights(), PercentOff: ld.GetPercentOff()}
		if amt := ld.GetAmountOff(); amt != nil {
			mID := ulid.GenerateString()
			g.moneys = append(g.moneys, moneyInput(mID, amt))
			row.AmountOffId = mID
		}
		g.losDiscounts = append(g.losDiscounts, row)
	}
	for _, f := range u.GetFees() {
		row := feesql.CreateInput{
			Id:          ulid.GenerateString(),
			Code:        f.GetCode(),
			DisplayName: f.GetDisplayName(),
			Percent:     f.GetPercent(),
			PricingUnit: pricingUnitToStr(f.GetPricingUnit()),
			Taxable:     f.GetTaxable(),
		}
		if amt := f.GetAmount(); amt != nil {
			mID := ulid.GenerateString()
			g.moneys = append(g.moneys, moneyInput(mID, amt))
			row.AmountId = mID
		}
		g.fees = append(g.fees, row)
	}
	for _, t := range u.GetTaxes() {
		g.taxes = append(g.taxes, taxesql.CreateInput{
			Id:          ulid.GenerateString(),
			Code:        t.GetCode(),
			DisplayName: t.GetDisplayName(),
			Percent:     t.GetPercent(),
		})
	}
	for _, m := range u.GetMedia() {
		g.medias = append(g.medias, unitmediasql.CreateInput{
			Id:          ulid.GenerateString(),
			Uri:         m.GetUri(),
			Type:        mediaTypeToStr(m.GetType()),
			Title:       m.GetTitle(),
			Description: m.GetDescription(),
			MimeType:    m.GetMimeType(),
			SortOrder:   m.GetSortOrder(),
			Primary:     m.GetPrimary(),
		})
	}
	for _, name := range u.GetApplicablePromoCodes() {
		g.promoCodes = append(g.promoCodes, unitapplicablepromocodesql.CreateInput{
			Id:          ulid.GenerateString(),
			PromoCodeId: name,
		})
	}
	return g
}

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
		Description:  deref(res.Description),
		Type:         unitTypeFromStr(res.Type),
		BookingMode:  bookingModeFromStr(res.BookingMode),
		Capacity:     deref(res.Capacity),
		MaxOccupancy: deref(res.MaxOccupancy),
		TimeZone:     res.TimeZone,
		Price:        moneyFromModel(p.price),
		PricingUnit:  pricingUnitFromStr(res.PricingUnit),
		Duration:     strToDuration(res.Duration),
		Tags:         strPtrsToSlice(res.Tags),
		Attributes:   jsonToStruct(jsonBytes(res.Attributes)),
		State:        unitStateFromStr(res.State),
		CreateTime:   strToTS(res.CreateTime),
		UpdateTime:   strToTS(res.UpdateTime),
		Etag:         deref(res.Etag),
		Licences:     p.licenceNames,
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
		ld := &propertypbv1.LosDiscount{MinNights: l.MinNights, PercentOff: deref(l.PercentOff)}
		if l.AmountOffId != nil {
			ld.AmountOff = moneyFromModel(p.moneyByID[*l.AmountOffId])
		}
		out.LosDiscounts = append(out.LosDiscounts, ld)
	}
	for i := range p.fees {
		f := &p.fees[i]
		fee := &propertypbv1.Fee{
			Code:        f.Code,
			DisplayName: deref(f.DisplayName),
			Percent:     deref(f.Percent),
			PricingUnit: pricingUnitFromStr(f.PricingUnit),
			Taxable:     deref(f.Taxable),
		}
		if f.AmountId != nil {
			fee.Amount = moneyFromModel(p.moneyByID[*f.AmountId])
		}
		out.Fees = append(out.Fees, fee)
	}
	for i := range p.taxes {
		out.Taxes = append(out.Taxes, &propertypbv1.Tax{
			Code:        p.taxes[i].Code,
			DisplayName: deref(p.taxes[i].DisplayName),
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
		Title:       deref(m.Title),
		Description: deref(m.Description),
		MimeType:    deref(m.MimeType),
		SortOrder:   deref(m.SortOrder),
		Primary:     deref(m.Primary),
	}
}

func unitStateFromStr(s *string) propertypbv1.UnitState {
	if s == nil || *s == "" {
		return propertypbv1.UnitState_UNIT_STATE_UNSPECIFIED
	}
	return propertypbv1.UnitState(propertypbv1.UnitState_value["UNIT_STATE_"+*s])
}
