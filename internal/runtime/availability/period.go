// Resolving the window/date_range oneof into engine time.
package availability

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// period is a request's resolved period, in both instant and calendar forms.
type period struct {
	start, end time.Time // UTC instants
	dateStart  *date.Date
	dateEnd    *date.Date
	nights     int32
}

// resolvePeriod turns a request's window/date_range oneof into a period evaluated
// in the unit's timezone. Exactly one of window/dateRange must be set.
func resolvePeriod(u *engine.UnitInfo, window *sharedpbv1.TimeWindow, dr *sharedpbv1.DateRange) (period, error) {
	loc, err := time.LoadLocation(u.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	switch {
	case dr != nil && dr.GetStartDate() != nil && dr.GetEndDate() != nil:
		ds, de := dr.GetStartDate(), dr.GetEndDate()
		return period{
			start:     startOfDate(ds, loc),
			end:       startOfDate(de, loc),
			dateStart: ds,
			dateEnd:   de,
			nights:    engine.NightsBetween(ds, de, loc),
		}, nil
	case window != nil && window.GetStartTime() != nil && window.GetEndTime() != nil:
		s := window.GetStartTime().AsTime()
		e := window.GetEndTime().AsTime()
		ds := dateOf(s.In(loc))
		de := dateOf(e.In(loc))
		return period{
			start:     s.UTC(),
			end:       e.UTC(),
			dateStart: ds,
			dateEnd:   de,
			nights:    engine.NightsBetween(ds, de, loc),
		}, nil
	default:
		return period{}, status.Error(codes.InvalidArgument, "a window or date_range period is required")
	}
}

func startOfDate(d *date.Date, loc *time.Location) time.Time {
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, loc).UTC()
}

func dateOf(t time.Time) *date.Date {
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func modeProto(m string) sharedpbv1.BookingMode {
	switch m {
	case engine.ModeNightly:
		return sharedpbv1.BookingMode_BOOKING_MODE_NIGHTLY
	case engine.ModeTimeSlot:
		return sharedpbv1.BookingMode_BOOKING_MODE_TIME_SLOT
	default:
		return sharedpbv1.BookingMode_BOOKING_MODE_UNSPECIFIED
	}
}
