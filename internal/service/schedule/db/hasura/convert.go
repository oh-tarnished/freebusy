// Package hasura provides the Hasura/GraphQL-backed implementation of the
// schedule persistence contract (internal/service/schedule/db.ScheduleRepository).
// It adapts the generated freebusyql handlers to that contract, converting between
// the protobuf Schedule/AvailabilityException domain types and the normalized
// GraphQL schema (the schedule's buffers, stay-constraints, cancellation-policy +
// refund-tiers, and recurring-rule children, plus the shared TimeWindow/DateRange
// value-objects an exception's span references).
package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"strings"
	"time"

	exceptionsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	buffersschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/buffersettingsql"
	cancelschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/cancellationpoliciesql"
	recurringschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/recurringrulesql"
	refundschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/refundtiersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/resourceql"
	scheduleschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/schemaql"
	stayschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/stayconstraintsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/timewindowsql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	rfc3339    = time.RFC3339
	dateLayout = "2006-01-02"
)

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- scalar (de)serialization matching the GraphQL string columns --------------

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	t, err := time.Parse(rfc3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

func dateToStr(d *date.Date) string {
	if d == nil {
		return ""
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC).Format(dateLayout)
}

func strToDate(s string) *date.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func durationToStr(d *durationpb.Duration) string {
	if d == nil {
		return ""
	}
	return d.AsDuration().String()
}

func durationFromStr(s string) *durationpb.Duration {
	if s == "" {
		return nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return nil
	}
	return durationpb.New(d)
}

func weekdaysToStr(days []sharedpbv1.Weekday) string {
	if len(days) == 0 {
		return ""
	}
	parts := make([]string, 0, len(days))
	for _, d := range days {
		parts = append(parts, d.String())
	}
	return strings.Join(parts, ",")
}

func weekdaysFromStr(s *string) []sharedpbv1.Weekday {
	if s == nil || *s == "" {
		return nil
	}
	names := strings.Split(*s, ",")
	out := make([]sharedpbv1.Weekday, 0, len(names))
	for _, n := range names {
		out = append(out, sharedpbv1.Weekday(sharedpbv1.Weekday_value[strings.TrimSpace(n)]))
	}
	return out
}

func kindToStr(k schedulepbv1.ExceptionKind) string {
	if k == schedulepbv1.ExceptionKind_EXCEPTION_KIND_UNSPECIFIED {
		return ""
	}
	return strings.TrimPrefix(k.String(), "EXCEPTION_KIND_")
}

func kindFromStr(s string) schedulepbv1.ExceptionKind {
	if s == "" {
		return schedulepbv1.ExceptionKind_EXCEPTION_KIND_UNSPECIFIED
	}
	return schedulepbv1.ExceptionKind(schedulepbv1.ExceptionKind_value["EXCEPTION_KIND_"+s])
}

// --- schedule graph (insert inputs) ------------------------------------------

// scheduleGraph is the set of GraphQL insert inputs a Schedule materializes into:
// the schedule row, its belongs-to buffers / stay-constraints / cancellation
// policy, the policy's refund tiers, and the schedule's recurring rules. The FK
// ids (schedule_id, cancellation_policy_id) are stamped as the inputs are built.
type scheduleGraph struct {
	schedule           resourceql.CreateInput
	buffers            *buffersschema.CreateInput
	stayConstraints    *stayschema.CreateInput
	cancellationPolicy *cancelschema.CreateInput
	refundTiers        []refundschema.CreateInput
	recurringRules     []recurringschema.CreateInput
}

// buildScheduleGraph turns a proto Schedule into its insert graph under
// propertyID. The schedule row's identity (Id/Name/Etag) is stamped by the
// repository; child FKs are wired here.
func buildScheduleGraph(s *schedulepbv1.Schedule, propertyID string) *scheduleGraph {
	g := &scheduleGraph{}
	g.schedule = resourceql.CreateInput{PropertyId: propertyID}

	if b := s.GetBuffers(); b != nil {
		id := ulid.GenerateString()
		g.buffers = &buffersschema.CreateInput{
			Id:         id,
			StartDelta: durationToStr(b.GetStartDelta()),
			EndDelta:   durationToStr(b.GetEndDelta()),
			MinNotice:  durationToStr(b.GetMinNotice()),
			MaxAdvance: durationToStr(b.GetMaxAdvance()),
			Gap:        durationToStr(b.GetGap()),
		}
		g.schedule.BuffersId = id
	}
	if sc := s.GetStayConstraints(); sc != nil {
		id := ulid.GenerateString()
		g.stayConstraints = &stayschema.CreateInput{
			Id:               id,
			MinNights:        sc.GetMinNights(),
			MaxNights:        sc.GetMaxNights(),
			CheckinWeekdays:  weekdaysToStr(sc.GetCheckinWeekdays()),
			CheckoutWeekdays: weekdaysToStr(sc.GetCheckoutWeekdays()),
			AdvanceMinDays:   sc.GetAdvanceMinDays(),
			AdvanceMaxDays:   sc.GetAdvanceMaxDays(),
		}
		g.schedule.StayConstraintsId = id
	}
	if cp := s.GetCancellationPolicy(); cp != nil {
		id := ulid.GenerateString()
		g.cancellationPolicy = &cancelschema.CreateInput{Id: id}
		g.schedule.CancellationPolicyId = id
		for _, t := range cp.GetTiers() {
			g.refundTiers = append(g.refundTiers, refundschema.CreateInput{
				Id:                   ulid.GenerateString(),
				CancellationPolicyId: id,
				Cutoff:               durationToStr(t.GetCutoff()),
				RefundPercent:        t.GetRefundPercent(),
			})
		}
	}
	for _, r := range s.GetRecurringRules() {
		g.recurringRules = append(g.recurringRules, recurringschema.CreateInput{
			Id:     ulid.GenerateString(),
			Rrule:  r.GetRrule(),
			Opens:  r.GetOpens(),
			Closes: r.GetCloses(),
		})
	}
	return g
}

// scheduleParts is a schedule row plus its hydrated child rows.
type scheduleParts struct {
	res         *scheduleschema.ScheduleResource
	buffers     *scheduleschema.ScheduleBufferSettings
	stay        *scheduleschema.ScheduleStayConstraints
	refundTiers []scheduleschema.ScheduleRefundTiers
	recurring   []scheduleschema.ScheduleRecurringRules
	hasPolicy   bool
}

// scheduleRefs captures the child row ids to delete when a schedule is replaced.
type scheduleRefs struct {
	buffersID    *string
	stayID       *string
	cancelID     *string
	recurringIDs []string
}

func scheduleFromParts(p scheduleParts) *schedulepbv1.Schedule {
	out := &schedulepbv1.Schedule{
		Name: p.res.Name,
		Etag: repox.Deref(p.res.Etag),
	}
	if p.buffers != nil {
		out.Buffers = &schedulepbv1.BufferSettings{
			StartDelta: durationFromStr(repox.Deref(p.buffers.StartDelta)),
			EndDelta:   durationFromStr(repox.Deref(p.buffers.EndDelta)),
			MinNotice:  durationFromStr(repox.Deref(p.buffers.MinNotice)),
			MaxAdvance: durationFromStr(repox.Deref(p.buffers.MaxAdvance)),
			Gap:        durationFromStr(repox.Deref(p.buffers.Gap)),
		}
	}
	if p.stay != nil {
		out.StayConstraints = &schedulepbv1.StayConstraints{
			MinNights:        repox.Deref(p.stay.MinNights),
			MaxNights:        repox.Deref(p.stay.MaxNights),
			CheckinWeekdays:  weekdaysFromStr(p.stay.CheckinWeekdays),
			CheckoutWeekdays: weekdaysFromStr(p.stay.CheckoutWeekdays),
			AdvanceMinDays:   repox.Deref(p.stay.AdvanceMinDays),
			AdvanceMaxDays:   repox.Deref(p.stay.AdvanceMaxDays),
		}
	}
	if p.hasPolicy {
		policy := &schedulepbv1.CancellationPolicy{}
		for i := range p.refundTiers {
			policy.Tiers = append(policy.Tiers, &schedulepbv1.RefundTier{
				Cutoff:        durationFromStr(p.refundTiers[i].Cutoff),
				RefundPercent: p.refundTiers[i].RefundPercent,
			})
		}
		out.CancellationPolicy = policy
	}
	for i := range p.recurring {
		out.RecurringRules = append(out.RecurringRules, &schedulepbv1.RecurringRule{
			Rrule:  p.recurring[i].Rrule,
			Opens:  repox.Deref(p.recurring[i].Opens),
			Closes: repox.Deref(p.recurring[i].Closes),
		})
	}
	return out
}

// --- exception graph ---------------------------------------------------------

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
