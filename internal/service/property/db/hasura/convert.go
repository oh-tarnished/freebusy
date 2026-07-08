package hasura

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/postaladdressql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/genproto/googleapis/type/postaladdress"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf Property/Unit domain
// types and the normalized Hasura/GraphQL schema. Timestamps cross the boundary
// as RFC 3339 strings, dates as "2006-01-02", times as "15:04:05"; enums as their
// bare value name (no proto prefix); FK ids as strings (empty means unset).

const dateLayout = "2006-01-02"

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// toBigdec / fromBigdec cross the numeric(precision,scale) boundary: the
// GraphQL schema carries such columns as arbitrary-precision decimal strings.
func toBigdec(f float64) graphql.Bigdecimal {
	return graphql.Bigdecimal(strconv.FormatFloat(f, 'f', -1, 64))
}

func fromBigdec(b graphql.Bigdecimal) float64 {
	f, _ := strconv.ParseFloat(string(b), 64)
	return f
}

// lastSegment returns the final path component of an AIP resource name.
func lastSegment(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// orgName rebuilds the "organisations/{id}" resource name from the bare id the
// property row's organisation FK column stores (the FK references
// organisations.id).
func orgName(id string) string {
	if id == "" {
		return ""
	}
	return "organisations/" + id
}

func tsToStr(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t)
		}
	}
	return nil
}

func dateToStr(d *date.Date) string {
	if d == nil {
		return ""
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC).Format(dateLayout)
}

func strToDate(s string) *date.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s[:min(len(s), 10)])
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func todToStr(t *timeofday.TimeOfDay) string {
	if t == nil {
		return ""
	}
	return time.Date(0, 1, 1, int(t.GetHours()), int(t.GetMinutes()), int(t.GetSeconds()), 0, time.UTC).Format("15:04:05")
}

func strToTOD(s *string) *timeofday.TimeOfDay {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("15:04:05", (*s)[:min(len(*s), 8)])
	if err != nil {
		return nil
	}
	return &timeofday.TimeOfDay{Hours: int32(t.Hour()), Minutes: int32(t.Minute()), Seconds: int32(t.Second())}
}

func durationToStr(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func strToDuration(s *string) *durationpb.Duration {
	if s == nil || *s == "" {
		return nil
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
}

func weekdaysToStr(days []sharedpbv1.Weekday) string {
	if len(days) == 0 {
		return ""
	}
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, d.String())
	}
	return strings.Join(parts, ",")
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

// strSliceToPtrs adapts a proto repeated string to the []*string GraphQL arrays
// use; strPtrsToSlice reverses it (dropping nils).
func strSliceToPtrs(in []string) []*string {
	if len(in) == 0 {
		return nil
	}
	out := make([]*string, len(in))
	for i := range in {
		v := in[i]
		out[i] = &v
	}
	return out
}

func strPtrsToSlice(in []*string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, p := range in {
		if p != nil {
			out = append(out, *p)
		}
	}
	return out
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

// jsonBytes unwraps the *json.RawMessage the GraphQL schema uses for jsonb
// columns into the []byte the struct conversions expect.
func jsonBytes(r *json.RawMessage) []byte {
	if r == nil {
		return nil
	}
	return []byte(*r)
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

// --- value-object inputs/reads ----------------------------------------------

func moneyInput(id string, m *money.Money) moneysql.CreateInput {
	return moneysql.CreateInput{
		Id:           id,
		CurrencyCode: m.GetCurrencyCode(),
		Units:        graphql.Int64(m.GetUnits()),
		Nanos:        m.GetNanos(),
	}
}

func moneyFromModel(m *commonschema.CommonMoneys) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: deref(m.CurrencyCode),
		Units:        int64(deref(m.Units)),
		Nanos:        deref(m.Nanos),
	}
}

func addressInput(id string, a *postaladdress.PostalAddress) postaladdressql.CreateInput {
	return postaladdressql.CreateInput{
		Id:                 id,
		Revision:           a.GetRevision(),
		RegionCode:         a.GetRegionCode(),
		LanguageCode:       a.GetLanguageCode(),
		PostalCode:         a.GetPostalCode(),
		SortingCode:        a.GetSortingCode(),
		AdministrativeArea: a.GetAdministrativeArea(),
		Locality:           a.GetLocality(),
		Sublocality:        a.GetSublocality(),
		AddressLines:       strSliceToPtrs(a.GetAddressLines()),
		Recipients:         strSliceToPtrs(a.GetRecipients()),
		Organization:       a.GetOrganization(),
	}
}

func addressFromModel(a *commonschema.CommonPostalAddress) *postaladdress.PostalAddress {
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
		AddressLines:       strPtrsToSlice(a.AddressLines),
		Recipients:         strPtrsToSlice(a.Recipients),
		Organization:       deref(a.Organization),
	}
}

func dateRangeInput(id string, d *sharedpbv1.DateRange) daterangesql.CreateInput {
	return daterangesql.CreateInput{
		Id:        id,
		StartDate: dateToStr(d.GetStartDate()),
		EndDate:   dateToStr(d.GetEndDate()),
	}
}

func dateRangeFromModel(d *sharedschema.SharedDateRanges) *sharedpbv1.DateRange {
	if d == nil {
		return nil
	}
	return &sharedpbv1.DateRange{
		StartDate: strToDate(d.StartDate),
		EndDate:   strToDate(d.EndDate),
	}
}

// --- enum <-> bare value-name conversions -----------------------------------

func unitTypeToStr(t propertypbv1.UnitType) string {
	return strings.TrimPrefix(t.String(), "UNIT_TYPE_")
}

func unitTypeFromStr(s string) propertypbv1.UnitType {
	return propertypbv1.UnitType(propertypbv1.UnitType_value["UNIT_TYPE_"+s])
}

func bookingModeToStr(m sharedpbv1.BookingMode) string {
	return strings.TrimPrefix(m.String(), "BOOKING_MODE_")
}

func bookingModeFromStr(s string) sharedpbv1.BookingMode {
	return sharedpbv1.BookingMode(sharedpbv1.BookingMode_value["BOOKING_MODE_"+s])
}

func pricingUnitToStr(p propertypbv1.PricingUnit) string {
	if p == propertypbv1.PricingUnit_PRICING_UNIT_UNSPECIFIED {
		return ""
	}
	return strings.TrimPrefix(p.String(), "PRICING_UNIT_")
}

func pricingUnitFromStr(s *string) propertypbv1.PricingUnit {
	if s == nil || *s == "" {
		return propertypbv1.PricingUnit_PRICING_UNIT_UNSPECIFIED
	}
	return propertypbv1.PricingUnit(propertypbv1.PricingUnit_value["PRICING_UNIT_"+*s])
}

func mediaTypeToStr(t sharedpbv1.MediaType) string {
	return strings.TrimPrefix(t.String(), "MEDIA_TYPE_")
}

func mediaTypeFromStr(s string) sharedpbv1.MediaType {
	return sharedpbv1.MediaType(sharedpbv1.MediaType_value["MEDIA_TYPE_"+s])
}
