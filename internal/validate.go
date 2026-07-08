package internal

import (
	"context"
	"fmt"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// validationInterceptor returns a unary interceptor that checks every request
// message against its buf.validate rules (annotated in the protos and carried
// by the generated descriptors) before the handler runs. A violation maps to
// InvalidArgument with the rule's message, mirroring AIP error semantics.
//
// This covers the message-scoped rules only — resource-name shapes, ranges,
// CEL field relations. Rules that need state (party size vs a unit's max
// occupancy, capacity, booking state) stay in the repository layer.
func validationInterceptor() (grpc.UnaryServerInterceptor, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("build protovalidate validator: %w", err)
	}
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := validator.Validate(msg); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}, nil
}
