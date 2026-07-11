package gorm

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
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

// strOrNil maps an empty proto string (which cannot represent NULL) to a nil
// column pointer, so unset optional strings stay NULL in the database.
func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

func dateToTime(d *date.Date) time.Time {
	if d == nil {
		return time.Time{}
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC)
}
