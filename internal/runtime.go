package internal

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/runtime"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/availability/v1/availabilitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/runtime-go/grpc"
	stdgrpc "google.golang.org/grpc"
)

// newServiceInstance assembles the freebusy services (opening the configured
// database backend) and wraps them in the registered Service adapter. Called at
// startup and on every Restart.
func newServiceInstance() (*Service, error) {
	promoCode, err := runtime.NewPromoCodeServer()
	if err != nil {
		return nil, err
	}
	property, err := runtime.NewPropertyServer()
	if err != nil {
		return nil, err
	}
	organisation, err := runtime.NewOrganisationServer()
	if err != nil {
		return nil, err
	}
	schedule, err := runtime.NewScheduleServer()
	if err != nil {
		return nil, err
	}
	booking, err := runtime.NewBookingServer()
	if err != nil {
		return nil, err
	}
	availability, err := runtime.NewAvailabilityServer()
	if err != nil {
		return nil, err
	}
	identity, err := runtime.NewIdentityServer()
	if err != nil {
		return nil, err
	}
	return NewService(promoCode, property, organisation, schedule, booking, availability, identity), nil
}

// registerGRPCServers returns a server option that registers the freebusy gRPC
// servers with the hybrid server's gRPC multiplexer.
func registerGRPCServers(svc *Service) grpc.Option {
	return grpc.WithGRPCServers(func(s *grpc.GRPCServer) {
		registerServices(s, svc)
	})
}

// registerServices registers every freebusy service on a gRPC registrar — the
// hybrid server's multiplexer in production, a plain server in the e2e suites.
func registerServices(s stdgrpc.ServiceRegistrar, svc *Service) {
	promocodepbv1.RegisterPromoCodeServiceServer(s, svc)
	propertypbv1.RegisterPropertyServiceServer(s, svc)
	propertypbv1.RegisterLicenceServiceServer(s, svc)
	orgpbv1.RegisterOrganisationServiceServer(s, svc)
	schedulepbv1.RegisterScheduleServiceServer(s, svc)
	bookingpbv1.RegisterBookingServiceServer(s, svc)
	availabilitypbv1.RegisterAvailabilityServiceServer(s, svc)
	identitypbv1.RegisterIdentityServiceServer(s, svc)
}

// NewGRPCServer assembles the freebusy service against the configured (or
// test-overridden — see database.SetTestBackend) backend and returns a plain
// gRPC server carrying the production unary interceptor chain with every
// service registered. It is the seam the e2e suites serve over bufconn; the
// caller owns Serve and Stop.
func NewGRPCServer() (*stdgrpc.Server, *Service, error) {
	svc, err := newServiceInstance()
	if err != nil {
		return nil, nil, err
	}
	validate, err := validationInterceptor()
	if err != nil {
		return nil, nil, err
	}
	srv := stdgrpc.NewServer(stdgrpc.ChainUnaryInterceptor(validate))
	registerServices(srv, svc)
	return srv, svc, nil
}

// registerHTTPGateways returns a server option that registers the REST gateways
// so protobuf RPCs are also reachable over HTTP/JSON.
func registerHTTPGateways() grpc.Option {
	return grpc.WithHTTPGateways(func(mux *grpc.ServeMux, endpoint string, opts []grpc.DialOption) error {
		if err := promocodepbv1.RegisterPromoCodeServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		if err := propertypbv1.RegisterPropertyServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		if err := propertypbv1.RegisterLicenceServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		if err := orgpbv1.RegisterOrganisationServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		if err := schedulepbv1.RegisterScheduleServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		if err := bookingpbv1.RegisterBookingServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		if err := availabilitypbv1.RegisterAvailabilityServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts); err != nil {
			return err
		}
		return identitypbv1.RegisterIdentityServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	})
}

// registerMCPServices returns a server option that exposes the freebusy services
// over the Model Context Protocol, making them discoverable by LLM tool routers.
func registerMCPServices(svc *Service) grpc.Option {
	return grpc.WithMCPServices(func(ctx context.Context, cfg *grpc.MCPServerConfig) error {
		if err := promocodepbv1.ServePromoCodeServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		if err := propertypbv1.ServePropertyServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		if err := propertypbv1.ServeLicenceServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		if err := orgpbv1.ServeOrganisationServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		if err := schedulepbv1.ServeScheduleServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		if err := bookingpbv1.ServeBookingServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		if err := availabilitypbv1.ServeAvailabilityServiceMCP(ctx, svc, cfg); err != nil {
			return err
		}
		return identitypbv1.ServeIdentityServiceMCP(ctx, svc, cfg)
	})
}
