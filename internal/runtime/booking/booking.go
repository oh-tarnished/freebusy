// Package booking is the gRPC/protobuf layer for the BookingService: it
// implements bookingpbv1.BookingServiceServer, owning request validation,
// observability, and the mapping of repository errors to gRPC status codes.
// Persistence and the hold lifecycle stay behind db.BookingRepository.
package booking

import (
	"context"
	"errors"
	"strings"

	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/runtime/rpc"
	bookingdb "github.com/oh-tarnished/freebusy/internal/service/booking/db"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Server implements bookingpbv1.BookingServiceServer on top of a
// provider-agnostic db.BookingRepository.
type Server struct {
	bookingpbv1.UnimplementedBookingServiceServer
	repo bookingdb.BookingRepository
}

// NewServer returns a Server backed by repo.
func NewServer(repo bookingdb.BookingRepository) *Server {
	return &Server{repo: repo}
}

// CreateBooking places a hold on a unit for a window. validate_only checks the
// request without persisting a hold.
func (s *Server) CreateBooking(ctx context.Context, req *bookingpbv1.CreateBookingRequest) (*bookingpbv1.Booking, error) {
	b := proto.Clone(req.GetBooking()).(*bookingpbv1.Booking)
	if id := req.GetBookingId(); id != "" {
		name, err := types.BookingName(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid booking_id")
		}
		b.Name = name
	}
	if req.GetValidateOnly() {
		// Dry run: the request passed validation; report it without placing a hold.
		return b, nil
	}
	var out *bookingpbv1.Booking
	err := rpc.Traced(ctx, "BookingService", "CreateBooking", func(ctx context.Context) error {
		created, err := s.repo.CreateBooking(ctx, b)
		if err != nil {
			return toStatusErr(err)
		}
		out = created
		return nil
	})
	return out, err
}

// GetBooking returns a single booking by resource name.
func (s *Server) GetBooking(ctx context.Context, req *bookingpbv1.GetBookingRequest) (*bookingpbv1.Booking, error) {
	var out *bookingpbv1.Booking
	err := rpc.Traced(ctx, "BookingService", "GetBooking", func(ctx context.Context) error {
		b, err := s.repo.GetBooking(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = b
		return nil
	})
	return out, err
}

// ListBookings returns a page of bookings.
func (s *Server) ListBookings(ctx context.Context, req *bookingpbv1.ListBookingsRequest) (*bookingpbv1.ListBookingsResponse, error) {
	var out *bookingpbv1.ListBookingsResponse
	err := rpc.Traced(ctx, "BookingService", "ListBookings", func(ctx context.Context) error {
		items, next, err := s.repo.ListBookings(ctx, repox.ListInput{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    req.GetFilter(),
		})
		if err != nil {
			return toStatusErr(err)
		}
		out = &bookingpbv1.ListBookingsResponse{Bookings: items, NextPageToken: next}
		return nil
	})
	return out, err
}

// ConfirmBooking confirms a held booking.
func (s *Server) ConfirmBooking(ctx context.Context, req *bookingpbv1.ConfirmBookingRequest) (*bookingpbv1.Booking, error) {
	var out *bookingpbv1.Booking
	err := rpc.Traced(ctx, "BookingService", "ConfirmBooking", func(ctx context.Context) error {
		b, err := s.repo.ConfirmBooking(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = b
		return nil
	})
	return out, err
}

// CancelBooking cancels a booking, computing the refund from the cancellation policy.
func (s *Server) CancelBooking(ctx context.Context, req *bookingpbv1.CancelBookingRequest) (*bookingpbv1.Booking, error) {
	var out *bookingpbv1.Booking
	err := rpc.Traced(ctx, "BookingService", "CancelBooking", func(ctx context.Context) error {
		b, err := s.repo.CancelBooking(ctx, req.GetName(), req.GetReason())
		if err != nil {
			return toStatusErr(err)
		}
		out = b
		return nil
	})
	return out, err
}

// PreviewCancellation reports the refund a cancellation would yield now.
func (s *Server) PreviewCancellation(ctx context.Context, req *bookingpbv1.PreviewCancellationRequest) (*bookingpbv1.PreviewCancellationResponse, error) {
	var out *bookingpbv1.PreviewCancellationResponse
	err := rpc.Traced(ctx, "BookingService", "PreviewCancellation", func(ctx context.Context) error {
		refundable, pct, amount, nonRefundable, summary, err := s.repo.PreviewCancellation(ctx, req.GetName())
		if err != nil {
			return toStatusErr(err)
		}
		out = &bookingpbv1.PreviewCancellationResponse{
			Refundable:          refundable,
			RefundPercent:       pct,
			RefundAmount:        amount,
			NonRefundableAmount: nonRefundable,
			PolicySummary:       summary,
		}
		return nil
	})
	return out, err
}

// GetBookingGuests returns the staying party (guests + occupancy) on a booking.
func (s *Server) GetBookingGuests(ctx context.Context, req *bookingpbv1.GetBookingGuestsRequest) (*bookingpbv1.BookingGuests, error) {
	bookingName, ok := bookingNameFromGuestsName(req.GetName())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "name must be of the form bookings/{booking}/guests")
	}
	var out *bookingpbv1.BookingGuests
	err := rpc.Traced(ctx, "BookingService", "GetBookingGuests", func(ctx context.Context) error {
		b, err := s.repo.GetBooking(ctx, bookingName)
		if err != nil {
			return toStatusErr(err)
		}
		out = &bookingpbv1.BookingGuests{Name: req.GetName(), Guests: b.GetGuests(), Occupancy: b.GetOccupancy()}
		return nil
	})
	return out, err
}

// UpdateBookingGuests replaces the staying party (guests + occupancy) on a
// booking. Allowed only while PENDING_HOLD or CONFIRMED.
func (s *Server) UpdateBookingGuests(ctx context.Context, req *bookingpbv1.UpdateBookingGuestsRequest) (*bookingpbv1.BookingGuests, error) {
	bg := req.GetBookingGuests()
	bookingName, ok := bookingNameFromGuestsName(bg.GetName())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "booking_guests.name must be of the form bookings/{booking}/guests")
	}
	var out *bookingpbv1.BookingGuests
	err := rpc.Traced(ctx, "BookingService", "UpdateBookingGuests", func(ctx context.Context) error {
		b, err := s.repo.UpdateBookingGuests(ctx, bookingName, bg.GetGuests(), bg.GetOccupancy())
		if err != nil {
			return toStatusErr(err)
		}
		out = &bookingpbv1.BookingGuests{Name: bg.GetName(), Guests: b.GetGuests(), Occupancy: b.GetOccupancy()}
		return nil
	})
	return out, err
}

// bookingNameFromGuestsName strips the singleton "/guests" suffix from a
// BookingGuests resource name, returning the parent Booking's name.
func bookingNameFromGuestsName(name string) (string, bool) {
	const suffix = "/guests"
	if !strings.HasSuffix(name, suffix) {
		return "", false
	}
	booking := strings.TrimSuffix(name, suffix)
	if booking == "" {
		return "", false
	}
	return booking, true
}

// toStatusErr maps repository sentinel errors onto gRPC status codes. Booking
// diverges from the shared rpc.ToStatusErr on conflicts: a capacity/overlap
// conflict surfaces as FailedPrecondition (the request can't be satisfied in
// the current state), distinct from an etag Aborted conflict.
func toStatusErr(err error) error {
	if errors.Is(err, types.ErrConflict) {
		return status.Error(codes.FailedPrecondition, err.Error())
	}
	return rpc.ToStatusErr(err)
}

// RescheduleBooking moves a booking to a new span (and optionally unit).
func (s *Server) RescheduleBooking(ctx context.Context, req *bookingpbv1.RescheduleBookingRequest) (*bookingpbv1.Booking, error) {
	var out *bookingpbv1.Booking
	err := rpc.Traced(ctx, "BookingService", "RescheduleBooking", func(ctx context.Context) error {
		b, err := s.repo.RescheduleBooking(ctx, req.GetName(), &bookingpbv1.Booking{Window: req.GetWindow()}, req.GetUnit())
		if err != nil {
			return toStatusErr(err)
		}
		out = b
		return nil
	})
	return out, err
}
