package gorm

import (
	"testing"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/genproto/googleapis/type/postaladdress"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/durationpb"
)

// roundTripProperty materializes a proto Property into its row graph, wires the
// preloaded associations back onto the property model (as the repository's
// preload would), and re-hydrates the proto — exercising build* and *fromModel
// without a database.
func roundTripProperty(in *propertypbv1.Property) *propertypbv1.Property {
	g := buildPropertyGraph(in)
	g.property.ID = "p1"
	g.property.Name = in.GetName()
	g.property.Address = g.address
	g.property.Policy = g.policy
	for _, m := range g.medias {
		m.PropertyID = "p1"
		g.property.Medias = append(g.property.Medias, *m)
	}
	return propertyFromModel(g.property)
}

func roundTripUnit(in *propertypbv1.Unit) *propertypbv1.Unit {
	g := buildUnitGraph(in, "p1")
	g.unit.ID = "u1"
	g.unit.Name = in.GetName()

	moneyByID := map[string]*common.Money{}
	if g.price != nil {
		moneyByID[g.price.ID] = g.price
	}
	for _, m := range g.moneys {
		moneyByID[m.ID] = m
	}
	dateByID := map[string]*shared.DateRange{}
	for _, d := range g.dates {
		dateByID[d.ID] = d
	}

	g.unit.Price = g.price
	for i := range g.rateOverrides {
		r := g.rateOverrides[i]
		r.Price = moneyByID[r.PriceID]
		if r.DateRangeID != nil {
			r.DateRange = dateByID[*r.DateRangeID]
		}
		g.unit.RateOverrides = append(g.unit.RateOverrides, *r)
	}
	for i := range g.losDiscounts {
		l := g.losDiscounts[i]
		if l.AmountOffID != nil {
			l.AmountOff = moneyByID[*l.AmountOffID]
		}
		g.unit.LosDiscounts = append(g.unit.LosDiscounts, *l)
	}
	for i := range g.fees {
		f := g.fees[i]
		if f.AmountID != nil {
			f.Amount = moneyByID[*f.AmountID]
		}
		g.unit.Fees = append(g.unit.Fees, *f)
	}
	for i := range g.taxes {
		g.unit.Taxes = append(g.unit.Taxes, *g.taxes[i])
	}
	for i := range g.medias {
		g.unit.UnitMedias = append(g.unit.UnitMedias, *g.medias[i])
	}
	for i := range g.promoCodes {
		g.unit.UnitApplicablePromoCodes = append(g.unit.UnitApplicablePromoCodes, *g.promoCodes[i])
	}
	return unitFromModel(g.unit)
}

func TestPropertyRoundTrip(t *testing.T) {
	in := &propertypbv1.Property{
		Name:         "properties/p1",
		Organisation: "organisations/acme",
		DisplayName:  "Grand Beach Resort",
		Description:  "Beachfront property in Goa",
		TimeZone:     "Asia/Kolkata",
		Address:      &postaladdress.PostalAddress{RegionCode: "IN", Locality: "Goa", AddressLines: []string{"Beach Rd"}},
		Policy: &propertypbv1.Policy{
			CheckinTime:  &timeofday.TimeOfDay{Hours: 14},
			CheckoutTime: &timeofday.TimeOfDay{Hours: 11},
			HouseRules:   []string{"No smoking", "No pets"},
			Notes:        "Quiet hours after 10pm",
		},
		Tags: []string{"beachfront", "5-star"},
		Media: []*propertypbv1.Media{
			{Uri: "s3://bucket/hero.jpg", Type: sharedpbv1.MediaType_MEDIA_TYPE_IMAGE, Title: "Hero", Primary: true},
			{Uri: "s3://bucket/facts.pdf", Type: sharedpbv1.MediaType_MEDIA_TYPE_DOCUMENT},
		},
	}
	out := roundTripProperty(in)

	if out.GetOrganisation() != "organisations/acme" || out.GetDisplayName() != "Grand Beach Resort" || out.GetTimeZone() != "Asia/Kolkata" {
		t.Fatalf("scalars not preserved: %+v", out)
	}
	if got := out.GetAddress(); got.GetRegionCode() != "IN" || got.GetLocality() != "Goa" || len(got.GetAddressLines()) != 1 {
		t.Fatalf("address not preserved: %+v", got)
	}
	if pol := out.GetPolicy(); pol.GetCheckinTime().GetHours() != 14 || pol.GetCheckoutTime().GetHours() != 11 || len(pol.GetHouseRules()) != 2 {
		t.Fatalf("policy not preserved: %+v", pol)
	}
	if len(out.GetTags()) != 2 {
		t.Fatalf("tags not preserved: %v", out.GetTags())
	}
	if m := out.GetMedia(); len(m) != 2 || m[0].GetUri() != "s3://bucket/hero.jpg" || !m[0].GetPrimary() || m[1].GetType() != sharedpbv1.MediaType_MEDIA_TYPE_DOCUMENT {
		t.Fatalf("media not preserved: %+v", m)
	}
}

func TestUnitRoundTrip(t *testing.T) {
	in := &propertypbv1.Unit{
		Name:         "properties/p1/units/u1",
		DisplayName:  "Deluxe King",
		Type:         propertypbv1.UnitType_UNIT_TYPE_ROOM,
		BookingMode:  sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY,
		Capacity:     10,
		MaxOccupancy: 3,
		TimeZone:     "Asia/Kolkata",
		Price:        &money.Money{CurrencyCode: "INR", Units: 5000},
		PricingUnit:  propertypbv1.PricingUnit_PRICING_UNIT_PER_NIGHT,
		Duration:     durationpb.New(0),
		RateOverrides: []*propertypbv1.RateOverride{{
			DateRange: &sharedpbv1.DateRange{
				StartDate: &date.Date{Year: 2026, Month: 12, Day: 24},
				EndDate:   &date.Date{Year: 2026, Month: 12, Day: 27},
			},
			Weekdays: []sharedpbv1.Weekday{sharedpbv1.Weekday_WEEKDAY_FRIDAY, sharedpbv1.Weekday_WEEKDAY_SATURDAY},
			Price:    &money.Money{CurrencyCode: "INR", Units: 8000},
		}},
		LosDiscounts:         []*propertypbv1.LosDiscount{{MinNights: 3, PercentOff: 10}},
		Fees:                 []*propertypbv1.Fee{{Code: "cleaning_fee", Amount: &money.Money{CurrencyCode: "INR", Units: 500}, Taxable: true}},
		Taxes:                []*propertypbv1.Tax{{Code: "gst", Percent: 12}},
		Media:                []*propertypbv1.UnitMedia{{Uri: "s3://bucket/room.jpg", Type: sharedpbv1.MediaType_MEDIA_TYPE_IMAGE}},
		ApplicablePromoCodes: []string{"promo-codes/SUMMER25"},
		Tags:                 []string{"sea-view"},
	}
	out := roundTripUnit(in)

	if out.GetType() != propertypbv1.UnitType_UNIT_TYPE_ROOM || out.GetBookingMode() != sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY {
		t.Fatalf("enums not preserved: %+v", out)
	}
	if out.GetCapacity() != 10 || out.GetMaxOccupancy() != 3 {
		t.Fatalf("capacity/occupancy not preserved: %+v", out)
	}
	if out.GetPrice().GetUnits() != 5000 || out.GetPricingUnit() != propertypbv1.PricingUnit_PRICING_UNIT_PER_NIGHT {
		t.Fatalf("price not preserved: %+v", out.GetPrice())
	}
	if ro := out.GetRateOverrides(); len(ro) != 1 || ro[0].GetPrice().GetUnits() != 8000 ||
		ro[0].GetDateRange().GetStartDate().GetDay() != 24 || len(ro[0].GetWeekdays()) != 2 {
		t.Fatalf("rate override not preserved: %+v", ro)
	}
	if ld := out.GetLosDiscounts(); len(ld) != 1 || ld[0].GetMinNights() != 3 || ld[0].GetPercentOff() != 10 {
		t.Fatalf("los discount not preserved: %+v", ld)
	}
	if f := out.GetFees(); len(f) != 1 || f[0].GetCode() != "cleaning_fee" || f[0].GetAmount().GetUnits() != 500 || !f[0].GetTaxable() {
		t.Fatalf("fee not preserved: %+v", f)
	}
	if tx := out.GetTaxes(); len(tx) != 1 || tx[0].GetCode() != "gst" || tx[0].GetPercent() != 12 {
		t.Fatalf("tax not preserved: %+v", tx)
	}
	if m := out.GetMedia(); len(m) != 1 || m[0].GetUri() != "s3://bucket/room.jpg" {
		t.Fatalf("unit media not preserved: %+v", m)
	}
	if pc := out.GetApplicablePromoCodes(); len(pc) != 1 || pc[0] != "promo-codes/SUMMER25" {
		t.Fatalf("applicable promo codes not preserved: %v", pc)
	}
}
