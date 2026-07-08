package internal

import (
	"context"
	"testing"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The interceptor enforces the buf.validate rules annotated on the protos:
// the BookingGuests singleton name must match ^bookings/[^/]+/guests$ and
// occupancy counts are >= 0. Valid requests pass through to the handler
// untouched.
func TestValidationInterceptor(t *testing.T) {
	intercept, err := validationInterceptor()
	if err != nil {
		t.Fatalf("build interceptor: %v", err)
	}

	handlerCalled := false
	handler := func(ctx context.Context, req any) (any, error) {
		handlerCalled = true
		return "ok", nil
	}

	// Bad resource name → InvalidArgument before the handler runs.
	bad := &bookingpbv1.UpdateBookingGuestsRequest{
		BookingGuests: &bookingpbv1.BookingGuests{Name: "rooms/nope"},
	}
	if _, err := intercept(context.Background(), bad, nil, handler); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("bad name: got err %v, want InvalidArgument", err)
	}
	if handlerCalled {
		t.Fatal("handler must not run on a validation failure")
	}

	// Negative occupancy → InvalidArgument.
	neg := &bookingpbv1.UpdateBookingGuestsRequest{
		BookingGuests: &bookingpbv1.BookingGuests{
			Name:      "bookings/b1/guests",
			Occupancy: &bookingpbv1.Occupancy{Adults: -1},
		},
	}
	if _, err := intercept(context.Background(), neg, nil, handler); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("negative occupancy: got err %v, want InvalidArgument", err)
	}

	// Valid request → handler runs.
	good := &bookingpbv1.UpdateBookingGuestsRequest{
		BookingGuests: &bookingpbv1.BookingGuests{
			Name:      "bookings/b1/guests",
			Occupancy: &bookingpbv1.Occupancy{Adults: 2, Children: 1},
		},
	}
	out, err := intercept(context.Background(), good, nil, handler)
	if err != nil || out != "ok" || !handlerCalled {
		t.Fatalf("valid request: out=%v err=%v handlerCalled=%t", out, err, handlerCalled)
	}
}
