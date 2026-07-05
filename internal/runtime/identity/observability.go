// Package identity is the gRPC/protobuf layer for the IdentityService: it
// implements identitypbv1.IdentityServiceServer, owning request validation, the
// "users/me" caller resolution, observability, and the mapping of repository
// errors to gRPC status codes. Persistence stays behind db.UserRepository.
package identity

import (
	"context"
	"errors"

	"github.com/machanirobotics/pulse/pulse-go"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type rpcMetrics struct {
	Requests int64 `pulse:"metric:counter:identity.rpc.requests"`
	Errors   int64 `pulse:"metric:counter:identity.rpc.errors"`
}

// traced runs fn inside a pulse span named "IdentityService/<method>", records
// request/error counters, and logs the outcome.
func traced(ctx context.Context, method string, fn func(context.Context) error) error {
	return shared.Pulse.Tracing.Trace(ctx, "IdentityService/"+method, nil, func(ctx context.Context, _ *pulse.Span) error {
		err := fn(ctx)
		recordRPC(method, err)
		switch {
		case err == nil:
			shared.Pulse.Logger.Debug(method + " ok")
		case isServerError(err):
			_ = shared.Pulse.Logger.Error(method+" failed", map[string]any{"error": err.Error()})
		default:
			shared.Pulse.Logger.Debug(method+" rejected", map[string]any{"error": err.Error()})
		}
		return err
	})
}

func recordRPC(method string, err error) {
	m := rpcMetrics{Requests: 1}
	if isServerError(err) {
		m.Errors = 1
	}
	_ = shared.Pulse.Metrics.Record(m, pulse.WithAttributes(pulse.StringAttribute("method", method)))
}

func isServerError(err error) bool {
	if err == nil {
		return false
	}
	switch status.Code(err) {
	case codes.Internal, codes.Unknown, codes.DataLoss, codes.Unavailable, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

// toStatusErr maps repository sentinel errors onto gRPC status codes.
func toStatusErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, types.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, types.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, types.ErrConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, types.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	return status.Error(codes.Internal, err.Error())
}
