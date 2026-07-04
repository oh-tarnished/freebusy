// Package internal is the transport/bootstrap layer: it builds the hybrid
// gRPC/HTTP/MCP server and registers the freebusy services assembled by
// internal/runtime. The protobuf/gRPC translation lives under internal/runtime;
// the database layer stays agnostic to it.
package internal

import (
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// Service is the registered gRPC adapter. It embeds the assembled service
// implementations, so it satisfies each of their gRPC server interfaces
// (promocodepbv1.PromoCodeServiceServer, propertypbv1.PropertyServiceServer,
// orgpbv1.OrganisationServiceServer, schedulepbv1.ScheduleServiceServer, and any
// future service interfaces composed in here).
type Service struct {
	promocodepbv1.PromoCodeServiceServer
	propertypbv1.PropertyServiceServer
	orgpbv1.OrganisationServiceServer
	schedulepbv1.ScheduleServiceServer
}

// NewService wraps the assembled service servers as the registered Service.
func NewService(
	promoCode promocodepbv1.PromoCodeServiceServer,
	property propertypbv1.PropertyServiceServer,
	organisation orgpbv1.OrganisationServiceServer,
	schedule schedulepbv1.ScheduleServiceServer,
) *Service {
	return &Service{
		PromoCodeServiceServer:    promoCode,
		PropertyServiceServer:     property,
		OrganisationServiceServer: organisation,
		ScheduleServiceServer:     schedule,
	}
}
