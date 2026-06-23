// Package internal is the transport/bootstrap layer: it builds the hybrid
// gRPC/HTTP/MCP server and registers the freebusy services assembled by
// internal/runtime. The protobuf/gRPC translation lives under internal/runtime;
// the database layer stays agnostic to it.
package internal

import (
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
)

// Service is the registered gRPC adapter. It embeds the assembled promocode
// service implementation, so it satisfies promocodepbv1.PromoCodeServiceServer
// (and any future service interfaces composed in here).
type Service struct {
	promocodepbv1.PromoCodeServiceServer
}

// NewService wraps the assembled promocode server as the registered Service.
func NewService(promoCode promocodepbv1.PromoCodeServiceServer) *Service {
	return &Service{PromoCodeServiceServer: promoCode}
}
