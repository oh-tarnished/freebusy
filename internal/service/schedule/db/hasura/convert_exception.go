// Exception write graph and read assembly.
package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	exceptionsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	scheduleschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/timewindowsql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// exceptionGraph is the insert graph an AvailabilityException materializes into:
// the exception row plus its belongs-to TimeWindow or DateRange span.
type exceptionGraph struct {
	exc    exceptionsql.CreateInput
	window *timewindowsql.CreateInput
	dates  *daterangesql.CreateInput
}

func buildExceptionGraph(e *schedulepbv1.AvailabilityException, propertyID, unitID string, now time.Time) *exceptionGraph {
	g := &exceptionGraph{}
	g.exc = exceptionsql.CreateInput{
		Kind:       kindToStr(e.GetKind()),
		Reason:     e.GetReason(),
		PropertyId: propertyID,
		UnitId:     unitID,
		CreateTime: dbutil.TsToStr(timestamppb.New(now)),
	}
	if w := e.GetWindow(); w != nil {
		id := ulid.GenerateString()
		g.window = &timewindowsql.CreateInput{
			Id:        id,
			StartTime: dbutil.TsToStr(w.GetStartTime()),
			EndTime:   dbutil.TsToStr(w.GetEndTime()),
		}
		g.exc.WindowId = id
		g.exc.SpanCase = "WINDOW"
	} else if dr := e.GetDateRange(); dr != nil {
		id := ulid.GenerateString()
		g.dates = &daterangesql.CreateInput{
			Id:        id,
			StartDate: dateToStr(dr.GetStartDate()),
			EndDate:   dateToStr(dr.GetEndDate()),
		}
		g.exc.DateRangeId = id
		g.exc.SpanCase = "DATE_RANGE"
	}
	return g
}

func exceptionFromParts(res *scheduleschema.ScheduleAvailabilityExceptions, window *sharedschema.SharedTimeWindows, dates *sharedschema.SharedDateRanges) *schedulepbv1.AvailabilityException {
	out := &schedulepbv1.AvailabilityException{
		Name:       res.Name,
		Kind:       kindFromStr(res.Kind),
		Reason:     repox.Deref(res.Reason),
		CreateTime: strToTS(res.CreateTime),
	}
	switch {
	case window != nil:
		out.Span = &schedulepbv1.AvailabilityException_Window{Window: &sharedpbv1.TimeWindow{
			StartTime: strToTS(window.StartTime),
			EndTime:   strToTS(window.EndTime),
		}}
	case dates != nil:
		out.Span = &schedulepbv1.AvailabilityException_DateRange{DateRange: &sharedpbv1.DateRange{
			StartDate: strToDate(dates.StartDate),
			EndDate:   strToDate(dates.EndDate),
		}}
	}
	return out
}
