// Package runtime assembles the freebusy gRPC services from configuration: it
// opens the configured database backend, builds the provider-agnostic
// repository, and constructs the protobuf service implementations in the sibling
// packages (runtime/promocode). The transport layer (package internal) registers
// what this package builds; the database layer stays agnostic to protobuf.
package runtime

import (
	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/runtime/availability"
	"github.com/oh-tarnished/freebusy/internal/runtime/booking"
	"github.com/oh-tarnished/freebusy/internal/runtime/organisation"
	"github.com/oh-tarnished/freebusy/internal/runtime/promocode"
	"github.com/oh-tarnished/freebusy/internal/runtime/property"
	"github.com/oh-tarnished/freebusy/internal/runtime/schedule"
	availabilitydb "github.com/oh-tarnished/freebusy/internal/service/availability/db"
	bookingdb "github.com/oh-tarnished/freebusy/internal/service/booking/db"
	organisationdb "github.com/oh-tarnished/freebusy/internal/service/organisation/db"
	promocodedb "github.com/oh-tarnished/freebusy/internal/service/promocode/db"
	propertydb "github.com/oh-tarnished/freebusy/internal/service/property/db"
	scheduledb "github.com/oh-tarnished/freebusy/internal/service/schedule/db"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// NewPromoCodeServer opens the configured backend, builds the repository, and
// returns the promocode gRPC service implementation ready to register.
func NewPromoCodeServer() (promocodepbv1.PromoCodeServiceServer, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return promocode.NewServer(promocodedb.New(conn)), nil
}

// NewPropertyServer opens the configured backend, builds the repository, and
// returns the property gRPC service implementation ready to register.
func NewPropertyServer() (propertypbv1.PropertyServiceServer, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return property.NewServer(propertydb.New(conn)), nil
}

// NewOrganisationServer opens the configured backend, builds the repository, and
// returns the organisation gRPC service implementation ready to register.
func NewOrganisationServer() (orgpbv1.OrganisationServiceServer, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return organisation.NewServer(organisationdb.New(conn)), nil
}

// NewScheduleServer opens the configured backend, builds the repository, and
// returns the schedule gRPC service implementation ready to register.
func NewScheduleServer() (schedulepbv1.ScheduleServiceServer, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return schedule.NewServer(scheduledb.New(conn)), nil
}

// NewBookingServer opens the configured backend, builds the repository, and
// returns the booking gRPC service implementation ready to register. It returns
// the concrete *booking.Server (which satisfies bookingpbv1.BookingServiceServer)
// so the caller can also start its background hold sweeper.
func NewBookingServer() (*booking.Server, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return booking.NewServer(bookingdb.New(conn)), nil
}

// NewAvailabilityServer opens the configured backend, builds the read port, and
// returns the availability gRPC service implementation ready to register.
func NewAvailabilityServer() (availabilitypbv1.AvailabilityServiceServer, error) {
	conn, err := database.Open()
	if err != nil {
		return nil, err
	}
	return availability.NewServer(availabilitydb.New(conn)), nil
}

// Other Services can be added here in the future, following the same pattern: open the database connection, build the repository, and return the service implementation.
