package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	exceptionsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	recurringschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/recurringrulesql"
	refundschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/refundtiersql"
	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/resourceql"
	scheduleschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/schemaql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// ScheduleRepository is the Hasura-backed schedule + availability-exception
// repository. A Schedule is a per-unit singleton whose child graph is read with
// follow-up queries and written as one atomic mutation batch (svc.Mutation.Tx());
// an AvailabilityException is a separate resource under the unit.
type ScheduleRepository struct {
	svc *freebusyql.Service
}

// NewScheduleRepository returns a Hasura-backed ScheduleRepository bound to svc.
func NewScheduleRepository(svc *freebusyql.Service) *ScheduleRepository {
	return &ScheduleRepository{svc: svc}
}

// --- Schedule ----------------------------------------------------------------

// GetSchedule returns a unit's schedule configuration. When none is stored yet the
// singleton is reported as an empty Schedule (name only). The exceptions list is
// always derived from the unit's AvailabilityException rows.
func (r *ScheduleRepository) GetSchedule(ctx context.Context, name string) (*schedulepbv1.Schedule, error) {
	_, unitID, err := types.ParseScheduleName(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Schedule.Resource.Find(ctx, resourceql.List().Where(resourceql.Name.Eq(name)))
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	var out *schedulepbv1.Schedule
	if res == nil {
		out = &schedulepbv1.Schedule{Name: name}
	} else {
		parts, _, err := r.hydrateSchedule(ctx, res)
		if err != nil {
			return nil, err
		}
		out = scheduleFromParts(parts)
	}
	names, err := r.exceptionNames(ctx, unitID)
	if err != nil {
		return nil, err
	}
	out.Exceptions = names
	return out, nil
}

// hydrateSchedule loads a schedule row's belongs-to children (buffers, stay
// constraints, cancellation policy + refund tiers) and its recurring rules, and
// returns the ids of the replaceable child rows.
func (r *ScheduleRepository) hydrateSchedule(ctx context.Context, res *scheduleschema.ScheduleResource) (scheduleParts, scheduleRefs, error) {
	p := scheduleParts{res: res}
	refs := scheduleRefs{buffersID: res.BuffersId, stayID: res.StayConstraintsId, cancelID: res.CancellationPolicyId}

	if res.BuffersId != nil {
		b, err := r.svc.Query.Schedule.BufferSettings.Get(ctx, *res.BuffersId)
		if err != nil {
			return scheduleParts{}, scheduleRefs{}, mapHasuraErr(err)
		}
		p.buffers = b
	}
	if res.StayConstraintsId != nil {
		s, err := r.svc.Query.Schedule.StayConstraints.Get(ctx, *res.StayConstraintsId)
		if err != nil {
			return scheduleParts{}, scheduleRefs{}, mapHasuraErr(err)
		}
		p.stay = s
	}
	if res.CancellationPolicyId != nil {
		p.hasPolicy = true
		tiers, err := r.svc.Query.Schedule.RefundTiers.List(ctx,
			refundschema.List().Where(refundschema.CancellationPolicyId.Eq(*res.CancellationPolicyId)))
		if err != nil {
			return scheduleParts{}, scheduleRefs{}, mapHasuraErr(err)
		}
		p.refundTiers = tiers
	}
	rules, err := r.svc.Query.Schedule.RecurringRules.List(ctx,
		recurringschema.List().Where(recurringschema.ScheduleId.Eq(res.Id)))
	if err != nil {
		return scheduleParts{}, scheduleRefs{}, mapHasuraErr(err)
	}
	p.recurring = rules
	for i := range rules {
		refs.recurringIDs = append(refs.recurringIDs, rules[i].Id)
	}
	return p, refs, nil
}

// exceptionNames returns the resource names of a unit's availability exceptions.
func (r *ScheduleRepository) exceptionNames(ctx context.Context, unitID string) ([]string, error) {
	rows, err := r.svc.Query.Schedule.AvailabilityExceptions.List(ctx,
		exceptionsql.List().Where(exceptionsql.UnitId.Eq(unitID)).OrderBy(exceptionsql.CreateTime.Asc()))
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	names := make([]string, 0, len(rows))
	for i := range rows {
		names = append(names, rows[i].Name)
	}
	return names, nil
}

// --- AvailabilityException ---------------------------------------------------

func (r *ScheduleRepository) CreateAvailabilityException(ctx context.Context, parent string, e *schedulepbv1.AvailabilityException) (*schedulepbv1.AvailabilityException, error) {
	propertyID, unitID, id, name, err := types.ResolveAvailabilityExceptionName(parent, e.GetName())
	if err != nil {
		return nil, err
	}
	g := buildExceptionGraph(e, propertyID, unitID, time.Now().UTC())
	g.exc.Id = id
	g.exc.Name = name

	tx := r.svc.Mutation.Tx()
	if g.window != nil {
		var res sharedschema.InsertSharedTimeWindowsResponse
		tx.Add(r.svc.Mutation.Shared.TimeWindows.CreateOp(*g.window, &res))
	}
	if g.dates != nil {
		var res sharedschema.InsertSharedDateRangesResponse
		tx.Add(r.svc.Mutation.Shared.DateRanges.CreateOp(*g.dates, &res))
	}
	var eRes scheduleschema.InsertScheduleAvailabilityExceptionsResponse
	tx.Add(r.svc.Mutation.Schedule.AvailabilityExceptions.CreateOp(g.exc, &eRes))
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetAvailabilityException(ctx, name)
}

func (r *ScheduleRepository) GetAvailabilityException(ctx context.Context, name string) (*schedulepbv1.AvailabilityException, error) {
	id, err := types.AvailabilityExceptionID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Schedule.AvailabilityExceptions.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return r.hydrateException(ctx, res)
}

func (r *ScheduleRepository) ListAvailabilityExceptions(ctx context.Context, parent string, params types.ListParams) ([]*schedulepbv1.AvailabilityException, string, error) {
	unitID, err := types.UnitID(parent)
	if err != nil {
		return nil, "", err
	}
	order, err := exceptionOrderTerms(params.OrderBy)
	if err != nil {
		return nil, "", err
	}
	where, err := exceptionFilterPredicate(params.Filter, unitID)
	if err != nil {
		return nil, "", err
	}
	limit, offset := types.PageBounds(params)
	req := exceptionsql.List().Limit(limit + 1).Offset(offset).Where(where)
	if len(order) > 0 {
		req = req.OrderBy(order...)
	}
	rows, err := r.svc.Query.Schedule.AvailabilityExceptions.List(ctx, req)
	if err != nil {
		return nil, "", mapHasuraErr(err)
	}
	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		next = types.EncodeOffset(offset + limit)
	}
	items := make([]*schedulepbv1.AvailabilityException, 0, len(rows))
	for i := range rows {
		exc, err := r.hydrateException(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, exc)
	}
	return items, next, nil
}

// hydrateException loads an exception's TimeWindow or DateRange span and builds
// the protobuf message.
func (r *ScheduleRepository) hydrateException(ctx context.Context, res *scheduleschema.ScheduleAvailabilityExceptions) (*schedulepbv1.AvailabilityException, error) {
	var window *sharedschema.SharedTimeWindows
	var dates *sharedschema.SharedDateRanges
	if res.WindowId != nil {
		w, err := r.svc.Query.Shared.TimeWindows.Get(ctx, *res.WindowId)
		if err != nil {
			return nil, mapHasuraErr(err)
		}
		window = w
	}
	if res.DateRangeId != nil {
		d, err := r.svc.Query.Shared.DateRanges.Get(ctx, *res.DateRangeId)
		if err != nil {
			return nil, mapHasuraErr(err)
		}
		dates = d
	}
	return exceptionFromParts(res, window, dates), nil
}
