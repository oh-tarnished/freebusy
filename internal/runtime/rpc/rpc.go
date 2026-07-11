// Package rpc is the boundary between freebusy's gRPC handlers and the
// platform runtime: Traced hands each handler body to the shared runtime-go
// Observer (spans, rpc.requests/rpc.errors counters, outcome logs, all emitted
// through freebusy's pulse identity), and ToStatusErr maps freebusy's
// repository sentinel errors onto gRPC status codes — the one piece that is
// domain-specific and stays here.
package rpc

import (
	"context"
	"errors"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/shared"
	"github.com/oh-tarnished/runtime-go/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// observer wraps handler bodies with tracing, metrics, and outcome logging;
// built once for the process against freebusy's pulse client.
var observer = grpc.NewObserver(shared.Pulse)

// Traced runs fn inside a span named "<service>/<method>" with request/error
// counters and outcome logging. See grpc.Observer.Traced.
func Traced(ctx context.Context, service, method string, fn func(context.Context) error) error {
	return observer.Traced(ctx, service, method, fn)
}

// ToStatusErr maps repository sentinel errors onto gRPC status codes. Errors
// that are already gRPC statuses (e.g. InvalidArgument from request
// validation) pass through unchanged; anything else becomes Internal. A
// conflict maps to Aborted (the optimistic-concurrency retryable); booking
// overrides that locally to FailedPrecondition for capacity conflicts.
func ToStatusErr(err error) error {
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
	case errors.Is(err, types.ErrUnimplemented):
		return status.Error(codes.Unimplemented, err.Error())
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	return status.Error(codes.Internal, err.Error())
}
