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

	buffersschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/buffersettingsql"
	cancelschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/cancellationpoliciesql"
	recurringschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/recurringrulesql"
	refundschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/refundtiersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/resourceql"
	scheduleschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/schemaql"
	stayschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/stayconstraintsql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// --- scalar (de)serialization matching the GraphQL string columns --------------

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
