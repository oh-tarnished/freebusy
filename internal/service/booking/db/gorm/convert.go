package gorm

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf Booking type and its
// GORM storage model. The booking row references its unit / customer / promo code
// by bare id (the FKs target those tables' ids) and stores its window, contact,
// and Money value-objects as belongs-to rows. Instants (window) are UTC; nights
// are counted in the unit's timezone.

const defaultHoldTTL = 15 * time.Minute

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func tsToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

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

// moneyMul multiplies a proto Money by n, normalizing the nanos carry.
func moneyMul(m *money.Money, n int64) *money.Money {
	if m == nil {
		return nil
	}
	total := (m.GetUnits()*1_000_000_000 + int64(m.GetNanos())) * n
	return &money.Money{
		CurrencyCode: m.GetCurrencyCode(),
		Units:        total / 1_000_000_000,
		Nanos:        int32(total % 1_000_000_000),
	}
}

// moneyPct returns pct percent of m (used for refund amounts).
func moneyPct(m *money.Money, pct int32) *money.Money {
	if m == nil {
		return nil
	}
	total := (m.GetUnits()*1_000_000_000 + int64(m.GetNanos())) * int64(pct) / 100
	return &money.Money{
		CurrencyCode: m.GetCurrencyCode(),
		Units:        total / 1_000_000_000,
		Nanos:        int32(total % 1_000_000_000),
	}
}

func contactToModel(c *sharedpbv1.Contact) *shared.Contact {
	if c == nil {
		return nil
	}
	return &shared.Contact{
		ID:          ulid.GenerateString(),
		DisplayName: strOrNil(c.GetDisplayName()),
		Email:       strOrNil(c.GetEmail()),
		PhoneNumber: strOrNil(c.GetPhoneNumber()),
	}
}

func contactFromModel(c *shared.Contact) *sharedpbv1.Contact {
	if c == nil {
		return nil
	}
	return &sharedpbv1.Contact{
		DisplayName: deref(c.DisplayName),
		Email:       deref(c.Email),
		PhoneNumber: deref(c.PhoneNumber),
	}
}

func timeWindowToModel(w *sharedpbv1.TimeWindow) *shared.TimeWindow {
	if w == nil {
		return nil
	}
	return &shared.TimeWindow{
		ID:        ulid.GenerateString(),
		StartTime: w.GetStartTime().AsTime().UTC(),
		EndTime:   w.GetEndTime().AsTime().UTC(),
	}
}

func timeWindowFromModel(w *shared.TimeWindow) *sharedpbv1.TimeWindow {
	if w == nil {
		return nil
	}
	return &sharedpbv1.TimeWindow{
		StartTime: timestamppb.New(w.StartTime),
		EndTime:   timestamppb.New(w.EndTime),
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

// --- enum + name conversions -------------------------------------------------

func stateFromModel(s *booking.BookingState) bookingpbv1.BookingState {
	if s == nil {
		return bookingpbv1.BookingState_BOOKING_STATE_UNSPECIFIED
	}
	return bookingpbv1.BookingState(bookingpbv1.BookingState_value["BOOKING_STATE_"+string(*s)])
}

func cancelReasonToModel(r bookingpbv1.CancelReason) *booking.CancelReason {
	if r == bookingpbv1.CancelReason_CANCEL_REASON_UNSPECIFIED {
		return nil
	}
	v := booking.CancelReason(trimPrefix(r.String(), "CANCEL_REASON_"))
	return &v
}

func cancelReasonFromModel(r *booking.CancelReason) bookingpbv1.CancelReason {
	if r == nil {
		return bookingpbv1.CancelReason_CANCEL_REASON_UNSPECIFIED
	}
	return bookingpbv1.CancelReason(bookingpbv1.CancelReason_value["CANCEL_REASON_"+string(*r)])
}

// userNameOrEmpty rebuilds "users/{id}" from the bare customer FK id.
func userNameOrEmpty(id *string) string {
	if id == nil || *id == "" {
		return ""
	}
	return "users/" + *id
}

// promoCodeNameOrEmpty rebuilds the promo code resource name from the bare id.
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

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

// lastSegment returns the final path component of an AIP resource name, or ""
// for an empty input — used to reduce a full name to the bare id an FK stores.
func lastSegment(name string) string {
	if name == "" {
		return ""
	}
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			return name[i+1:]
		}
	}
	return name
}

// bookingFromModel assembles the protobuf Booking from a stored row, its
// preloaded value-objects, and the unit's full resource name (resolved by the
// repository, since the row stores only the bare unit id).
func bookingFromModel(m *booking.Booking, unitName string) *bookingpbv1.Booking {
	return &bookingpbv1.Booking{
		Name:           m.Name,
		Unit:           unitName,
		Customer:       userNameOrEmpty(m.CustomerID),
		Contact:        contactFromModel(m.Contact),
		Units:          deref(m.Units),
		Window:         timeWindowFromModel(m.Window),
		AssignedUnit:   deref(m.AssignedUnit),
		State:          stateFromModel(m.State),
		HoldExpireTime: timeToTS(m.HoldExpireTime),
		Price:          moneyFromModel(m.Price),
		PromoCode:      promoCodeNameOrEmpty(m.PromoCodeID),
		Discount:       moneyFromModel(m.Discount),
		Total:          moneyFromModel(m.Total),
		Notes:          deref(m.Notes),
		Attributes:     jsonToStruct(m.Attributes),
		CancelReason:   cancelReasonFromModel(m.CancelReason),
		CreateTime:     timeToTS(&m.CreateTime),
		UpdateTime:     timeToTS(&m.UpdateTime),
		ConfirmTime:    timeToTS(m.ConfirmTime),
		CancelTime:     timeToTS(m.CancelTime),
		RefundAmount:   moneyFromModel(m.RefundAmount),
		RefundPercent:  deref(m.RefundPercent),
		HoldTtl:        durationFromStr(m.HoldTtl),
		Etag:           deref(m.Etag),
	}
}
