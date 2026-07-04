package hasura

import (
	"context"
	"errors"
	"strings"

	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/resourceql"
	scheduleschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/schemaql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"github.com/oh-tarnished/generateql/runtime/go/runtime"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// UpdateSchedule upserts a unit's schedule configuration as one atomic mutation
// batch. The schedule is a singleton (created on first update); an empty paths
// slice replaces every mutable section, and s.Etag guards against concurrent
// writes on an existing schedule. The merged proto is re-materialized into a fresh
// child graph and the superseded rows are deleted in the same batch.
func (r *ScheduleRepository) UpdateSchedule(ctx context.Context, s *schedulepbv1.Schedule, paths []string) (*schedulepbv1.Schedule, error) {
	propertyID, _, err := types.ParseScheduleName(s.GetName())
	if err != nil {
		return nil, err
	}
	existing, err := r.svc.Query.Schedule.Resource.Find(ctx, resourceql.List().Where(resourceql.Name.Eq(s.GetName())))
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	exists := existing != nil
	if exists && s.GetEtag() != "" && existing.Etag != nil && s.GetEtag() != *existing.Etag {
		return nil, types.ErrConflict
	}

	merged := &schedulepbv1.Schedule{Name: s.GetName()}
	var oldRefs scheduleRefs
	if exists {
		parts, refs, err := r.hydrateSchedule(ctx, existing)
		if err != nil {
			return nil, err
		}
		merged = scheduleFromParts(parts)
		oldRefs = refs
	}
	applyScheduleMask(merged, s, paths)
	g := buildScheduleGraph(merged, propertyID)

	tx := r.svc.Mutation.Tx()
	if g.buffers != nil {
		var res scheduleschema.InsertScheduleBufferSettingsResponse
		tx.Add(r.svc.Mutation.Schedule.BufferSettings.CreateOp(*g.buffers, &res))
	}
	if g.stayConstraints != nil {
		var res scheduleschema.InsertScheduleStayConstraintsResponse
		tx.Add(r.svc.Mutation.Schedule.StayConstraints.CreateOp(*g.stayConstraints, &res))
	}
	if g.cancellationPolicy != nil {
		var res scheduleschema.InsertScheduleCancellationPoliciesResponse
		tx.Add(r.svc.Mutation.Schedule.CancellationPolicies.CreateOp(*g.cancellationPolicy, &res))
	}

	if exists {
		patch := resourceql.UpdateInput{
			BuffersId:            nullableStr(g.schedule.BuffersId),
			StayConstraintsId:    nullableStr(g.schedule.StayConstraintsId),
			CancellationPolicyId: nullableStr(g.schedule.CancellationPolicyId),
			Etag:                 graphql.Value(ulid.GenerateString()),
		}
		var updRes scheduleschema.UpdateScheduleResourceByIdResponse
		tx.Add(r.svc.Mutation.Schedule.Resource.UpdateOp(existing.Id, patch, &updRes))
		r.queueScheduleChildren(tx, g, existing.Id)
		queueScheduleChildDeletes(tx, r, oldRefs)
	} else {
		g.schedule.Id = ulid.GenerateString()
		g.schedule.Name = s.GetName()
		g.schedule.Etag = ulid.GenerateString()
		var res scheduleschema.InsertScheduleResourceResponse
		tx.Add(r.svc.Mutation.Schedule.Resource.CreateOp(g.schedule, &res))
		r.queueScheduleChildren(tx, g, g.schedule.Id)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetSchedule(ctx, s.GetName())
}

// queueScheduleChildren appends the has-many inserts (refund tiers referencing the
// new cancellation policy, recurring rules referencing scheduleID).
func (r *ScheduleRepository) queueScheduleChildren(tx *runtime.Tx, g *scheduleGraph, scheduleID string) {
	for i := range g.refundTiers {
		var res scheduleschema.InsertScheduleRefundTiersResponse
		tx.Add(r.svc.Mutation.Schedule.RefundTiers.CreateOp(g.refundTiers[i], &res))
	}
	for i := range g.recurringRules {
		g.recurringRules[i].ScheduleId = scheduleID
		var res scheduleschema.InsertScheduleRecurringRulesResponse
		tx.Add(r.svc.Mutation.Schedule.RecurringRules.CreateOp(g.recurringRules[i], &res))
	}
}

// queueScheduleChildDeletes appends deletes for a schedule's superseded recurring
// rules and its now-unreferenced belongs-to rows. The old policy's refund tiers
// cascade in the DB when the policy is deleted.
func queueScheduleChildDeletes(tx *runtime.Tx, r *ScheduleRepository, refs scheduleRefs) {
	for _, id := range refs.recurringIDs {
		var res scheduleschema.DeleteScheduleRecurringRulesByIdResponse
		tx.Add(r.svc.Mutation.Schedule.RecurringRules.DeleteOp(id, &res))
	}
	if refs.cancelID != nil {
		var res scheduleschema.DeleteScheduleCancellationPoliciesByIdResponse
		tx.Add(r.svc.Mutation.Schedule.CancellationPolicies.DeleteOp(*refs.cancelID, &res))
	}
	if refs.buffersID != nil {
		var res scheduleschema.DeleteScheduleBufferSettingsByIdResponse
		tx.Add(r.svc.Mutation.Schedule.BufferSettings.DeleteOp(*refs.buffersID, &res))
	}
	if refs.stayID != nil {
		var res scheduleschema.DeleteScheduleStayConstraintsByIdResponse
		tx.Add(r.svc.Mutation.Schedule.StayConstraints.DeleteOp(*refs.stayID, &res))
	}
}

// DeleteAvailabilityException removes an exception and the TimeWindow / DateRange
// value-object its span referenced, in one batch.
func (r *ScheduleRepository) DeleteAvailabilityException(ctx context.Context, name string) error {
	id, err := types.AvailabilityExceptionID(name)
	if err != nil {
		return err
	}
	res, err := r.svc.Query.Schedule.AvailabilityExceptions.Get(ctx, id)
	if err != nil {
		return mapHasuraErr(err)
	}
	if res == nil {
		return types.ErrNotFound
	}
	tx := r.svc.Mutation.Tx()
	var delRes scheduleschema.DeleteScheduleAvailabilityExceptionsByIdResponse
	tx.Add(r.svc.Mutation.Schedule.AvailabilityExceptions.DeleteOp(id, &delRes))
	if res.WindowId != nil {
		var out sharedschema.DeleteSharedTimeWindowsByIdResponse
		tx.Add(r.svc.Mutation.Shared.TimeWindows.DeleteOp(*res.WindowId, &out))
	}
	if res.DateRangeId != nil {
		var out sharedschema.DeleteSharedDateRangesByIdResponse
		tx.Add(r.svc.Mutation.Shared.DateRanges.DeleteOp(*res.DateRangeId, &out))
	}
	return mapHasuraErr(tx.Commit(ctx))
}

// nullableStr maps an empty optional string to a SQL NULL update and a non-empty
// one to a value update, so clearing a section clears its FK column.
func nullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}

// mapHasuraErr translates GraphQL/runtime errors into the repository sentinels.
func mapHasuraErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, graphql.ErrConflict):
		return types.ErrConflict
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") {
		return types.ErrAlreadyExists
	}
	return err
}
