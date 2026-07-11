// Package hasura provides the Hasura/GraphQL-backed implementation of the booking
// persistence contract (internal/service/booking/db.BookingRepository). It adapts
// the generated freebusyql handlers to that contract, owning the hold lifecycle,
// the capacity/overlap check that prevents overbooking (computed by querying the
// unit's active bookings, since Hasura has no raw SQL), the shared pricing engine,
// and the refund computation from the unit's schedule cancellation policy.
package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"strings"
	"time"

	bookingschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/schemaql"
	moneysql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	contactsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/contactsql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	timewindowsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/timewindowsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const rfc3339 = time.RFC3339

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	t, err := time.Parse(rfc3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

func durationToStr(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func durationFromStr(s string) *durationpb.Duration {
	if s == "" {
		return nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
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

// --- value-object conversions ------------------------------------------------

// moneyInput builds an insert input for a Money value-object row with a fresh id.
func moneyInput(m *money.Money) moneysql.CreateInput {
	return moneysql.CreateInput{
		Id:           ulid.GenerateString(),
		CurrencyCode: m.GetCurrencyCode(),
		Units:        graphql.Int64(m.GetUnits()),
		Nanos:        m.GetNanos(),
	}
}

func moneyFromSchema(m *commonschema.CommonMoneys) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: repox.Deref(m.CurrencyCode),
		Units:        int64(repox.Deref(m.Units)),
		Nanos:        repox.Deref(m.Nanos),
	}
}

func contactInput(c *sharedpbv1.Contact) *contactsql.CreateInput {
	if c == nil {
		return nil
	}
	return &contactsql.CreateInput{
		Id:          ulid.GenerateString(),
		DisplayName: c.GetDisplayName(),
		Email:       c.GetEmail(),
		PhoneNumber: c.GetPhoneNumber(),
	}
}

func contactFromSchema(c *sharedschema.SharedContacts) *sharedpbv1.Contact {
	if c == nil {
		return nil
	}
	return &sharedpbv1.Contact{
		DisplayName: repox.Deref(c.DisplayName),
		Email:       repox.Deref(c.Email),
		PhoneNumber: repox.Deref(c.PhoneNumber),
	}
}

func windowInput(w *sharedpbv1.TimeWindow) timewindowsql.CreateInput {
	return timewindowsql.CreateInput{
		Id:        ulid.GenerateString(),
		StartTime: dbutil.TsToStr(w.GetStartTime()),
		EndTime:   dbutil.TsToStr(w.GetEndTime()),
	}
}

func windowFromSchema(w *sharedschema.SharedTimeWindows) *sharedpbv1.TimeWindow {
	if w == nil {
		return nil
	}
	return &sharedpbv1.TimeWindow{
		StartTime: strToTS(w.StartTime),
		EndTime:   strToTS(w.EndTime),
	}
}

// --- enum + name conversions -------------------------------------------------

func stateFromStr(s *string) bookingpbv1.BookingState {
	if s == nil || *s == "" {
		return bookingpbv1.BookingState_BOOKING_STATE_UNSPECIFIED
	}
	return bookingpbv1.BookingState(bookingpbv1.BookingState_value["BOOKING_STATE_"+*s])
}

func cancelReasonToStr(r bookingpbv1.CancelReason) string {
	if r == bookingpbv1.CancelReason_CANCEL_REASON_UNSPECIFIED {
		return ""
	}
	return strings.TrimPrefix(r.String(), "CANCEL_REASON_")
}

func cancelReasonFromStr(s *string) bookingpbv1.CancelReason {
	if s == nil || *s == "" {
		return bookingpbv1.CancelReason_CANCEL_REASON_UNSPECIFIED
	}
	return bookingpbv1.CancelReason(bookingpbv1.CancelReason_value["CANCEL_REASON_"+*s])
}

func userNameOrEmpty(id *string) string {
	if id == nil || *id == "" {
		return ""
	}
	return "users/" + *id
}

func promoCodeNameOrEmpty(id *string) string {
	if id == nil || *id == "" {
		return ""
	}
	name, err := types.PromoCodeName(*id)
	if err != nil {
		return ""
	}
	return name
}

// bookingParts is a booking row plus its hydrated value-objects and the unit's
// full resource name.
type bookingParts struct {
	res      *bookingschema.BookingResource
	unitName string
	contact  *sharedschema.SharedContacts
	window   *sharedschema.SharedTimeWindows
	price    *money.Money
	discount *money.Money
	total    *money.Money
	refund   *money.Money
}

func bookingFromParts(p bookingParts) *bookingpbv1.Booking {
	r := p.res
	return &bookingpbv1.Booking{
		Name:           r.Name,
		Unit:           p.unitName,
		Customer:       userNameOrEmpty(r.Customer),
		Contact:        contactFromSchema(p.contact),
		Units:          repox.Deref(r.Units),
		Window:         windowFromSchema(p.window),
		AssignedUnit:   repox.Deref(r.AssignedUnit),
		State:          stateFromStr(r.State),
		HoldExpireTime: strToTS(repox.Deref(r.HoldExpireTime)),
		Price:          p.price,
		PromoCode:      promoCodeNameOrEmpty(r.PromoCode),
		Discount:       p.discount,
		Total:          p.total,
		Notes:          repox.Deref(r.Notes),
		Attributes:     jsonToStruct(repox.Deref(r.Attributes)),
		CancelReason:   cancelReasonFromStr(r.CancelReason),
		CreateTime:     strToTS(r.CreateTime),
		UpdateTime:     strToTS(r.UpdateTime),
		ConfirmTime:    strToTS(repox.Deref(r.ConfirmTime)),
		CancelTime:     strToTS(repox.Deref(r.CancelTime)),
		RefundAmount:   p.refund,
		RefundPercent:  repox.Deref(r.RefundPercent),
		HoldTtl:        durationFromStr(repox.Deref(r.HoldTtl)),
		Etag:           repox.Deref(r.Etag),
	}
}

// nightsBetween counts calendar nights of a window evaluated in tz (an IANA name
// on the unit). It never returns less than one night.
func nightsBetween(w *sharedpbv1.TimeWindow, tz string) int64 {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	start := w.GetStartTime().AsTime().In(loc)
	end := w.GetEndTime().AsTime().In(loc)
	sd := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	ed := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)
	nights := int64(ed.Sub(sd).Hours() / 24)
	if nights < 1 {
		return 1
	}
	return nights
}
