package promocode

import (
	"context"
	"errors"

	"github.com/machanirobotics/pulse/pulse-go"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// rpcMetrics is recorded once per RPC. The method name is attached as a metric
// attribute (see recordRPC) so the counters can be sliced per endpoint. Errors
// counts only server faults (see isServerError), not expected client/business
// rejections like NotFound or FailedPrecondition.
type rpcMetrics struct {
	Requests int64 `pulse:"metric:counter:promocode.rpc.requests"`
	Errors   int64 `pulse:"metric:counter:promocode.rpc.errors"`
}

// traced runs fn inside a pulse span named "PromoCodeService/<method>", records
// request/error counters, and logs the outcome. The RPC result is returned to the
// caller through a closure variable; traced only carries the error so the tracing
// layer can set the span status. Expected client/business rejections are logged
// at debug, not error, and excluded from the error counter.
func traced(ctx context.Context, method string, fn func(context.Context) error) error {
	return shared.Pulse.Tracing.Trace(ctx, "PromoCodeService/"+method, nil, func(ctx context.Context, _ *pulse.Span) error {
		err := fn(ctx)
		recordRPC(method, err)
		switch {
		case err == nil:
			shared.Pulse.Logger.Debug(method + " ok")
		case isServerError(err):
			_ = shared.Pulse.Logger.Error(method+" failed", map[string]any{"error": err.Error()})
		default:
			// Expected client/business outcome (NotFound, FailedPrecondition, etc.) —
			// returned to the caller but not a service fault.
			shared.Pulse.Logger.Debug(method+" rejected", map[string]any{"error": err.Error()})
		}
		return err
	})
}

// recordRPC emits the per-call counters tagged with the method name. Only server
// faults increment Errors; expected client/business rejections do not.
func recordRPC(method string, err error) {
	m := rpcMetrics{Requests: 1}
	if isServerError(err) {
		m.Errors = 1
	}
	_ = shared.Pulse.Metrics.Record(m, pulse.WithAttributes(pulse.StringAttribute("method", method)))
}

// isServerError reports whether err is a server-side fault — the codes that mean
// the service itself misbehaved — as opposed to an expected client/business
// outcome (NotFound, InvalidArgument, FailedPrecondition, ResourceExhausted,
// Aborted, AlreadyExists, ...). A non-status error reads as Unknown, which counts.
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

// toStatusErr maps repository sentinel errors onto gRPC status codes. Errors that
// are already gRPC statuses (e.g. InvalidArgument from request validation) pass
// through unchanged; anything else becomes Internal.
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
