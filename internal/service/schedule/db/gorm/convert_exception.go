package gorm

import (
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// exceptionGraph is the set of rows an AvailabilityException materializes into:
// the exception row plus its belongs-to TimeWindow or DateRange span (created
// before it, since the exception references them).
type exceptionGraph struct {
	exc    *schedule.AvailabilityException
	window *shared.TimeWindow
	dates  *shared.DateRange
}

func buildExceptionGraph(e *schedulepbv1.AvailabilityException, propertyID, unitID string) *exceptionGraph {
	g := &exceptionGraph{}
	g.exc = &schedule.AvailabilityException{
		Kind:       kindToModel(e.GetKind()),
		Reason:     strOrNil(e.GetReason()),
		PropertyID: propertyID,
		UnitID:     unitID,
	}
	if w := e.GetWindow(); w != nil {
		g.window = timeWindowToModel(w)
		g.exc.WindowID = &g.window.ID
		span := schedule.AvailabilityExceptionSpanCaseWindow
		g.exc.SpanCase = &span
	} else if dr := e.GetDateRange(); dr != nil {
		g.dates = dateRangeToModel(dr)
		g.exc.DateRangeID = &g.dates.ID
		span := schedule.AvailabilityExceptionSpanCaseDateRange
		g.exc.SpanCase = &span
	}
	return g
}

func exceptionFromModel(m *schedule.AvailabilityException) *schedulepbv1.AvailabilityException {
	out := &schedulepbv1.AvailabilityException{
		Name:       m.Name,
		Kind:       kindFromModel(m.Kind),
		Reason:     deref(m.Reason),
		CreateTime: timeToTS(&m.CreateTime),
	}
	switch {
	case m.Window != nil:
		out.Span = &schedulepbv1.AvailabilityException_Window{Window: timeWindowFromModel(m.Window)}
	case m.DateRange != nil:
		out.Span = &schedulepbv1.AvailabilityException_DateRange{DateRange: dateRangeFromModel(m.DateRange)}
	}
	return out
}
