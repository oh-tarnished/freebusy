// Package internal is the transport/bootstrap layer: it builds the hybrid
// gRPC/HTTP/MCP server and registers the freebusy services assembled by
// internal/runtime. The protobuf/gRPC translation lives under internal/runtime;
// the database layer stays agnostic to it.
package internal

import (
	"context"

	bookingruntime "github.com/oh-tarnished/freebusy/internal/runtime/booking"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// Service is the registered gRPC adapter. It embeds the assembled service
// implementations, so it satisfies each of their gRPC server interfaces
// (promocode, property, organisation, schedule, booking, and any future service
// interfaces composed in here).
type Service struct {
	promocodepbv1.PromoCodeServiceServer
	propertypbv1.PropertyServiceServer
	orgpbv1.OrganisationServiceServer
	schedulepbv1.ScheduleServiceServer
	bookingpbv1.BookingServiceServer

	// booking is the concrete booking server, retained so background tasks (the
	// hold sweeper) can be started against it in StartBackground.
	booking *bookingruntime.Server
}

// NewService wraps the assembled service servers as the registered Service. The
// booking server is passed as its concrete type so its background hold sweeper can
// be started; it still satisfies bookingpbv1.BookingServiceServer for embedding.
func NewService(
	promoCode promocodepbv1.PromoCodeServiceServer,
	property propertypbv1.PropertyServiceServer,
	organisation orgpbv1.OrganisationServiceServer,
	schedule schedulepbv1.ScheduleServiceServer,
	booking *bookingruntime.Server,
) *Service {
	return &Service{
		PromoCodeServiceServer:    promoCode,
		PropertyServiceServer:     property,
		OrganisationServiceServer: organisation,
		ScheduleServiceServer:     schedule,
		BookingServiceServer:      booking,
		booking:                   booking,
	}
}

// StartBackground launches the service's background tasks, tied to ctx: the
// booking hold sweeper, which periodically expires lapsed holds. The goroutines
// exit when ctx is cancelled (on server Stop/Restart).
func (s *Service) StartBackground(ctx context.Context) {
	if s.booking != nil {
		s.booking.StartHoldSweeper(ctx, 0)
	}
}
