// Package booking is the gRPC/protobuf layer for the BookingService: it
// implements bookingpbv1.BookingServiceServer, owning request validation,
// observability, and the mapping of repository errors to gRPC status codes.
// Persistence and the hold lifecycle stay behind db.BookingRepository.
package booking

import (
	"context"

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
	b := req.GetBooking()
	switch {
	case b == nil:
		return nil, status.Error(codes.InvalidArgument, "booking is required")
	case b.GetUnit() == "":
		return nil, status.Error(codes.InvalidArgument, "booking.unit is required")
	case b.GetWindow() == nil:
		return nil, status.Error(codes.InvalidArgument, "booking.window is required")
	}
	b = proto.Clone(b).(*bookingpbv1.Booking)
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
	err := traced(ctx, "CreateBooking", func(ctx context.Context) error {
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
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *bookingpbv1.Booking
	err := traced(ctx, "GetBooking", func(ctx context.Context) error {
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
	filter, err := types.ParseFilter(req.GetFilter())
	if err != nil {
		return nil, toStatusErr(err)
	}
	var out *bookingpbv1.ListBookingsResponse
	err = traced(ctx, "ListBookings", func(ctx context.Context) error {
		items, next, err := s.repo.ListBookings(ctx, types.ListParams{
			PageSize:  req.GetPageSize(),
			PageToken: req.GetPageToken(),
			OrderBy:   req.GetOrderBy(),
			Filter:    filter,
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
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *bookingpbv1.Booking
	err := traced(ctx, "ConfirmBooking", func(ctx context.Context) error {
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
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *bookingpbv1.Booking
	err := traced(ctx, "CancelBooking", func(ctx context.Context) error {
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
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	var out *bookingpbv1.PreviewCancellationResponse
	err := traced(ctx, "PreviewCancellation", func(ctx context.Context) error {
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

// RescheduleBooking moves a booking to a new span (and optionally unit).
func (s *Server) RescheduleBooking(ctx context.Context, req *bookingpbv1.RescheduleBookingRequest) (*bookingpbv1.Booking, error) {
	switch {
	case req.GetName() == "":
		return nil, status.Error(codes.InvalidArgument, "name is required")
	case req.GetWindow() == nil:
		return nil, status.Error(codes.InvalidArgument, "window is required")
	}
	var out *bookingpbv1.Booking
	err := traced(ctx, "RescheduleBooking", func(ctx context.Context) error {
		b, err := s.repo.RescheduleBooking(ctx, req.GetName(), &bookingpbv1.Booking{Window: req.GetWindow()}, req.GetUnit())
		if err != nil {
			return toStatusErr(err)
		}
		out = b
		return nil
	})
	return out, err
}
