// Package service implements the freebusy gRPC service servers on top of the
// provider-agnostic repositories in internal/database/repository. Each server
// owns request validation, observability (pulse logging, tracing, and metrics),
// and the mapping of repository errors to gRPC status codes; persistence and
// protobuf conversions live in the repository layer.
package service

import (
	"context"
	"errors"

	"github.com/machanirobotics/pulse/pulse-go"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// rpcMetrics is recorded once per RPC. The method name is attached as a metric
// attribute (see recordRPC) so the counters can be sliced per endpoint.
type rpcMetrics struct {
	Requests int64 `pulse:"metric:counter:promocode.rpc.requests"`
	Errors   int64 `pulse:"metric:counter:promocode.rpc.errors"`
}

// traced runs fn inside a pulse span named "PromoCodeService/<method>", records
// request/error counters, and logs the outcome. The RPC result is returned to the
// caller through a closure variable; traced only carries the error so the tracing
// layer can set the span status.
func traced(ctx context.Context, method string, fn func(context.Context) error) error {
	return shared.Pulse.Tracing.Trace(ctx, "PromoCodeService/"+method, nil, func(ctx context.Context, _ *pulse.Span) error {
		err := fn(ctx)
		recordRPC(method, err)
		if err != nil {
			_ = shared.Pulse.Logger.Error(method+" failed", map[string]any{"error": err.Error()})
		} else {
			shared.Pulse.Logger.Debug(method + " ok")
		}
		return err
	})
}

// recordRPC emits the per-call counters tagged with the method name.
func recordRPC(method string, err error) {
	m := rpcMetrics{Requests: 1}
	if err != nil {
		m.Errors = 1
	}
	_ = shared.Pulse.Metrics.Record(m, pulse.WithAttributes(pulse.StringAttribute("method", method)))
}

// toStatusErr maps repository sentinel errors onto gRPC status codes. Errors that
// are already gRPC statuses (e.g. InvalidArgument from request validation) pass
// through unchanged; anything else becomes Internal.
func toStatusErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, repository.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, repository.ErrConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, repository.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	return status.Error(codes.Internal, err.Error())
}
