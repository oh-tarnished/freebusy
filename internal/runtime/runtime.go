// Package runtime wires the freebusy services onto a gRPC server and its HTTP/JSON
// gateway. It is the single place that knows which services exist; the bootstrap
// (main) only supplies a database.Connection and the transport, and every new
// service is added here once.
package runtime

import (
	"context"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/service"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/grpc"
)

// Register builds every service from conn (through the database factory) and
// registers it on the gRPC server.
func Register(s grpc.ServiceRegistrar, conn *database.Connection) {
	f := database.NewFactory(conn)
	promocodepbv1.RegisterPromoCodeServiceServer(s, service.NewPromoCodeServer(f.PromoCodes()))
}

// RegisterGateway registers every service's HTTP/JSON gateway handler against the
// running gRPC endpoint.
func RegisterGateway(ctx context.Context, mux *gwruntime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return promocodepbv1.RegisterPromoCodeServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
