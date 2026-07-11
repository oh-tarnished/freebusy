// Value-object input/read pairs and enum codecs.
package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/postaladdressql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/genproto/googleapis/type/postaladdress"
)

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
		CurrencyCode: repox.Deref(m.CurrencyCode),
		Units:        int64(repox.Deref(m.Units)),
		Nanos:        repox.Deref(m.Nanos),
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
		Revision:           repox.Deref(a.Revision),
		RegionCode:         repox.Deref(a.RegionCode),
		LanguageCode:       repox.Deref(a.LanguageCode),
		PostalCode:         repox.Deref(a.PostalCode),
		SortingCode:        repox.Deref(a.SortingCode),
		AdministrativeArea: repox.Deref(a.AdministrativeArea),
		Locality:           repox.Deref(a.Locality),
		Sublocality:        repox.Deref(a.Sublocality),
		AddressLines:       strPtrsToSlice(a.AddressLines),
		Recipients:         strPtrsToSlice(a.Recipients),
		Organization:       repox.Deref(a.Organization),
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
