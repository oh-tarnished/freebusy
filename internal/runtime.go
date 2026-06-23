package internal

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/runtime"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/runtime-go/grpc"
)

// newServiceInstance assembles the freebusy services (opening the configured
// database backend) and wraps them in the registered Service adapter. Called at
// startup and on every Restart.
func newServiceInstance() (*Service, error) {
	promoCode, err := runtime.NewPromoCodeServer()
	if err != nil {
		return nil, err
	}
	return NewService(promoCode), nil
}

// registerGRPCServers returns a server option that registers the PromoCode gRPC
// server with the hybrid server's gRPC multiplexer.
func registerGRPCServers(svc *Service) grpc.Option {
	return grpc.WithGRPCServers(func(s *grpc.GRPCServer) {
		promocodepbv1.RegisterPromoCodeServiceServer(s, svc)
	})
}

// registerHTTPGateways returns a server option that registers the PromoCode REST
// gateway so protobuf RPCs are also reachable over HTTP/JSON.
func registerHTTPGateways() grpc.Option {
	return grpc.WithHTTPGateways(func(mux *grpc.ServeMux, endpoint string, opts []grpc.DialOption) error {
		return promocodepbv1.RegisterPromoCodeServiceHandlerFromEndpoint(context.Background(), mux, endpoint, opts)
	})
}

// registerMCPServices returns a server option that exposes the PromoCode service
// over the Model Context Protocol, making it discoverable by LLM tool routers.
func registerMCPServices(svc *Service) grpc.Option {
	return grpc.WithMCPServices(func(ctx context.Context, cfg *grpc.MCPServerConfig) error {
		return promocodepbv1.ServePromoCodeServiceMCP(ctx, svc, cfg)
	})
}
