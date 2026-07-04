package gorm

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/genproto/googleapis/type/postaladdress"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure (side-effect-free) conversions between the protobuf
// Property/Unit domain types and the normalized GORM storage models. The
// protobuf API nests address, policy, media, and (on a unit) its pricing
// (rate overrides, LOS discounts, fees, taxes) as sub-messages; the schema
// stores each as its own belongs-to or has-many child table under the property
// schema, with Money/DateRange/PostalAddress normalized into the shared common
// tables. A build* function turns a proto into the row graph the repository
// persists in one transaction; a *fromModel function re-hydrates the proto from
// a preloaded model.

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// strOrNil maps an empty proto string (which cannot represent NULL) to a nil
// column pointer, so unset optional strings stay NULL in the database.
func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// lastSegment returns the final path component of an AIP resource name
// ("promoCodes/p1" -> "p1"), used to populate a join row's id column while the
// full name round-trips via a separate column.
func lastSegment(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// orgName rebuilds the "organisations/{id}" resource name from the bare id the
// property row's organisation FK column stores (the FK references
// organisations.id). Empty id yields the empty string.
func orgName(id string) string {
	if id == "" {
		return ""
	}
	return "organisations/" + id
}

// --- value-object conversions ------------------------------------------------

// moneyToModel builds a fresh common.Money row from a proto Money, or nil.
func moneyToModel(m *money.Money) *common.Money {
	if m == nil {
		return nil
	}
	return &common.Money{
		ID:           ulid.GenerateString(),
		CurrencyCode: strOrNil(m.GetCurrencyCode()),
		Units:        ptr(m.GetUnits()),
		Nanos:        ptr(m.GetNanos()),
	}
}

func moneyFromModel(m *common.Money) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: deref(m.CurrencyCode),
		Units:        deref(m.Units),
		Nanos:        deref(m.Nanos),
	}
}

// dateRangeToModel builds a fresh shared.DateRange row from a proto DateRange.
func dateRangeToModel(d *sharedpbv1.DateRange) *shared.DateRange {
	if d == nil {
		return nil
	}
	return &shared.DateRange{
		ID:        ulid.GenerateString(),
		StartDate: dateToTime(d.GetStartDate()),
		EndDate:   dateToTime(d.GetEndDate()),
	}
}

func dateRangeFromModel(d *shared.DateRange) *sharedpbv1.DateRange {
	if d == nil {
		return nil
	}
	return &sharedpbv1.DateRange{
		StartDate: timeToDate(d.StartDate),
		EndDate:   timeToDate(d.EndDate),
	}
}

func dateToTime(d *date.Date) time.Time {
	if d == nil {
		return time.Time{}
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC)
}

func timeToDate(t time.Time) *date.Date {
	if t.IsZero() {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

// todToTime maps a google.type.TimeOfDay onto a time-of-day time.Time (the zero
// date carrying only H:M:S), matching the `type:time` column.
func todToTime(t *timeofday.TimeOfDay) *time.Time {
	if t == nil {
		return nil
	}
	v := time.Date(0, 1, 1, int(t.GetHours()), int(t.GetMinutes()), int(t.GetSeconds()), int(t.GetNanos()), time.UTC)
	return &v
}

func timeToTOD(t *time.Time) *timeofday.TimeOfDay {
	if t == nil {
		return nil
	}
	return &timeofday.TimeOfDay{
		Hours:   int32(t.Hour()),
		Minutes: int32(t.Minute()),
		Seconds: int32(t.Second()),
		Nanos:   int32(t.Nanosecond()),
	}
}

// weekdaysToStr serializes a repeated Weekday enum into the single string column
// the ORM generates for it (comma-joined value names), round-tripping exactly.
func weekdaysToStr(days []sharedpbv1.Weekday) *string {
	if len(days) == 0 {
		return nil
	}
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, d.String())
	}
	return ptr(strings.Join(parts, ","))
}

func weekdaysFromStr(s *string) []sharedpbv1.Weekday {
	if s == nil || *s == "" {
		return nil
	}
	names := strings.Split(*s, ",")
	out := make([]sharedpbv1.Weekday, 0, len(names))
	for _, n := range names {
		out = append(out, sharedpbv1.Weekday(sharedpbv1.Weekday_value[strings.TrimSpace(n)]))
	}
	return out
}

func addressToModel(a *postaladdress.PostalAddress) *common.PostalAddress {
	if a == nil {
		return nil
	}
	return &common.PostalAddress{
		ID:                 ulid.GenerateString(),
		Revision:           ptr(a.GetRevision()),
		RegionCode:         strOrNil(a.GetRegionCode()),
		LanguageCode:       strOrNil(a.GetLanguageCode()),
		PostalCode:         strOrNil(a.GetPostalCode()),
		SortingCode:        strOrNil(a.GetSortingCode()),
		AdministrativeArea: strOrNil(a.GetAdministrativeArea()),
		Locality:           strOrNil(a.GetLocality()),
		Sublocality:        strOrNil(a.GetSublocality()),
		AddressLines:       a.GetAddressLines(),
		Recipients:         a.GetRecipients(),
		Organization:       strOrNil(a.GetOrganization()),
	}
}

func addressFromModel(a *common.PostalAddress) *postaladdress.PostalAddress {
	if a == nil {
		return nil
	}
	return &postaladdress.PostalAddress{
		Revision:           deref(a.Revision),
		RegionCode:         deref(a.RegionCode),
		LanguageCode:       deref(a.LanguageCode),
		PostalCode:         deref(a.PostalCode),
		SortingCode:        deref(a.SortingCode),
		AdministrativeArea: deref(a.AdministrativeArea),
		Locality:           deref(a.Locality),
		Sublocality:        deref(a.Sublocality),
		AddressLines:       a.AddressLines,
		Recipients:         a.Recipients,
		Organization:       deref(a.Organization),
	}
}

func structToJSON(s *structpb.Struct) []byte {
	if s == nil {
		return nil
	}
	b, err := s.MarshalJSON()
	if err != nil {
		return nil
	}
	return b
}

func jsonToStruct(b []byte) *structpb.Struct {
	if len(b) == 0 {
		return nil
	}
	s := &structpb.Struct{}
	if err := s.UnmarshalJSON(b); err != nil {
		return nil
	}
	return s
}

// --- enum conversions --------------------------------------------------------

func unitTypeToModel(t propertypbv1.UnitType) property.UnitType {
	return property.UnitType(strings.TrimPrefix(t.String(), "UNIT_TYPE_"))
}

func unitTypeFromModel(t property.UnitType) propertypbv1.UnitType {
	return propertypbv1.UnitType(propertypbv1.UnitType_value["UNIT_TYPE_"+string(t)])
}

func bookingModeToModel(m sharedpbv1.BookingMode) property.BookingMode {
	return property.BookingMode(strings.TrimPrefix(m.String(), "BOOKING_MODE_"))
}

func bookingModeFromModel(m property.BookingMode) sharedpbv1.BookingMode {
	return sharedpbv1.BookingMode(sharedpbv1.BookingMode_value["BOOKING_MODE_"+string(m)])
}

func pricingUnitToModel(p propertypbv1.PricingUnit) *property.PricingUnit {
	if p == propertypbv1.PricingUnit_PRICING_UNIT_UNSPECIFIED {
		return nil
	}
	return ptr(property.PricingUnit(strings.TrimPrefix(p.String(), "PRICING_UNIT_")))
}

func pricingUnitFromModel(p *property.PricingUnit) propertypbv1.PricingUnit {
	if p == nil {
		return propertypbv1.PricingUnit_PRICING_UNIT_UNSPECIFIED
	}
	return propertypbv1.PricingUnit(propertypbv1.PricingUnit_value["PRICING_UNIT_"+string(*p)])
}

func mediaTypeToModel(t sharedpbv1.MediaType) property.MediaType {
	return property.MediaType(strings.TrimPrefix(t.String(), "MEDIA_TYPE_"))
}

func mediaTypeFromModel(t property.MediaType) sharedpbv1.MediaType {
	return sharedpbv1.MediaType(sharedpbv1.MediaType_value["MEDIA_TYPE_"+string(t)])
}
