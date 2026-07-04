// Package db is the booking persistence layer. It defines the provider-agnostic
// BookingRepository contract (spoken in protobuf domain types) and a factory that
// builds the implementation for the configured backend. Shared, provider-neutral
// vocabulary (errors, list params, names, field masks) lives in internal/types.
package db

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/service/booking/db/gorm"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"google.golang.org/genproto/googleapis/type/money"
)

// BookingRepository provides the booking lifecycle: placing holds, confirming,
// cancelling (with refund), rescheduling, and reads. Errors are the sentinels in
// internal/types.
type BookingRepository interface {
	// CreateBooking places a PENDING_HOLD on a unit for a window (server-computed
	// price/state/hold-expiry). Rejects with types.ErrConflict when capacity is
	// exhausted for the window.
	CreateBooking(ctx context.Context, b *bookingpbv1.Booking) (*bookingpbv1.Booking, error)

	// GetBooking returns the booking by resource name.
	GetBooking(ctx context.Context, name string) (*bookingpbv1.Booking, error)

	// ListBookings returns a page of bookings.
	ListBookings(ctx context.Context, params types.ListParams) (items []*bookingpbv1.Booking, nextPageToken string, err error)

	// ConfirmBooking flips a held booking to CONFIRMED.
	ConfirmBooking(ctx context.Context, name string) (*bookingpbv1.Booking, error)

	// CancelBooking cancels a booking, computing the refund from the unit's
	// cancellation policy.
	CancelBooking(ctx context.Context, name string, reason bookingpbv1.CancelReason) (*bookingpbv1.Booking, error)

	// PreviewCancellation computes the refund a cancellation would yield now.
	PreviewCancellation(ctx context.Context, name string) (refundable bool, percent int32, amount, nonRefundable *money.Money, summary string, err error)

	// RescheduleBooking moves a booking to a new window (and optionally unit),
	// re-checking capacity and recomputing the price.
	RescheduleBooking(ctx context.Context, name string, b *bookingpbv1.Booking, newUnit string) (*bookingpbv1.Booking, error)
}

// Assert the provider implementation satisfies the contract here.
var _ BookingRepository = (*gorm.BookingRepository)(nil)

// New returns the BookingRepository for the configured provider. GORM is the
// default; a Hasura implementation is a follow-up increment, so for now every
// provider resolves to the GORM adapter.
func New(conn *database.Connection) BookingRepository {
	return gorm.NewBookingRepository(conn.PgSQLConn)
}
