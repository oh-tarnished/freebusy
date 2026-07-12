// Bulk query surface: closures, catalog search, and the shared value-object
// batch reads over DDN. (Booking reads live in reader_bookings.go.)
package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/daterangesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/timewindowsql"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/money"
)

// timeWindowsByID fetches the given time windows in one query, keyed by id.
func (r *AvailabilityReader) timeWindowsByID(ctx context.Context, ids []string) (map[string]timewindowsql.SharedTimeWindows, error) {
	out := make(map[string]timewindowsql.SharedTimeWindows, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	ws, err := r.svc.Query.Shared.TimeWindows.List(ctx, timewindowsql.List().Where(timewindowsql.Id.In(ids...)))
	if err != nil {
		return nil, mapErr(err)
	}
	for i := range ws {
		out[ws[i].Id] = ws[i]
	}
	return out, nil
}

// dateRangesByID fetches the given date ranges in one query, keyed by id.
func (r *AvailabilityReader) dateRangesByID(ctx context.Context, ids []string) (map[string]daterangesql.SharedDateRanges, error) {
	out := make(map[string]daterangesql.SharedDateRanges, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	ds, err := r.svc.Query.Shared.DateRanges.List(ctx, daterangesql.List().Where(daterangesql.Id.In(ids...)))
	if err != nil {
		return nil, mapErr(err)
	}
	for i := range ds {
		out[ds[i].Id] = ds[i]
	}
	return out, nil
}

// Closures returns the unit's CLOSURE exceptions as UTC spans.
func (r *AvailabilityReader) Closures(ctx context.Context, unitID, tz string) ([]engine.Closure, error) {
	m, err := r.ClosuresForUnits(ctx, []string{unitID}, map[string]string{unitID: tz})
	if err != nil {
		return nil, err
	}
	return m[unitID], nil
}

// ClosuresForUnits batches Closures across many units, each expanded in its tz.
func (r *AvailabilityReader) ClosuresForUnits(ctx context.Context, unitIDs []string, tzByUnit map[string]string) (map[string][]engine.Closure, error) {
	out := map[string][]engine.Closure{}
	if len(unitIDs) == 0 {
		return out, nil
	}
	rows, err := r.svc.Query.Schedule.AvailabilityExceptions.List(ctx, availabilityexceptionsql.List().Where(availabilityexceptionsql.And(
		availabilityexceptionsql.UnitId.In(unitIDs...),
		availabilityexceptionsql.Kind.Eq("CLOSURE"),
	)))
	if err != nil {
		return nil, mapErr(err)
	}
	wids := make([]string, 0, len(rows))
	dids := make([]string, 0, len(rows))
	for i := range rows {
		if rows[i].WindowId != nil {
			wids = append(wids, *rows[i].WindowId)
		}
		if rows[i].DateRangeId != nil {
			dids = append(dids, *rows[i].DateRangeId)
		}
	}
	windows, err := r.timeWindowsByID(ctx, wids)
	if err != nil {
		return nil, err
	}
	ranges, err := r.dateRangesByID(ctx, dids)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		e := &rows[i]
		switch {
		case e.WindowId != nil:
			if w, ok := windows[*e.WindowId]; ok {
				out[e.UnitId] = append(out[e.UnitId], engine.Closure{Start: parseTS(w.StartTime), End: parseTS(w.EndTime)})
			}
		case e.DateRangeId != nil:
			d, ok := ranges[*e.DateRangeId]
			if !ok {
				continue
			}
			loc, lerr := time.LoadLocation(tzByUnit[e.UnitId])
			if lerr != nil {
				loc = time.UTC
			}
			start := parseDate(d.StartDate, loc)
			end := parseDate(d.EndDate, loc)
			out[e.UnitId] = append(out[e.UnitId], engine.Closure{Start: start.UTC(), End: end.UTC()})
		}
	}
	return out, nil
}

// SearchUnits returns active units for the storefront search, scoped and filtered.
func (r *AvailabilityReader) SearchUnits(ctx context.Context, propertyRef, organisation, filter string) ([]*engine.UnitInfo, error) {
	preds := []graphql.Predicate{unitsql.State.Neq("ARCHIVED")}
	if propertyRef != "" {
		propertyID, err := types.PropertyID(propertyRef)
		if err != nil {
			return nil, err
		}
		preds = append(preds, unitsql.PropertyId.Eq(propertyID))
	}
	if organisation != "" {
		orgID, err := types.OrganisationID(organisation)
		if err != nil {
			return nil, err
		}
		props, err := r.svc.Query.Property.Properties.List(ctx, propertiesql.List().Where(propertiesql.Organisation.Eq(orgID)))
		if err != nil {
			return nil, mapErr(err)
		}
		ids := make([]string, 0, len(props))
		for i := range props {
			ids = append(ids, props[i].Id)
		}
		if len(ids) == 0 {
			return nil, nil
		}
		preds = append(preds, unitsql.PropertyId.In(ids...))
	}
	fp, err := unitFilterPredicate(filter)
	if err != nil {
		return nil, err
	}
	if fp != nil {
		preds = append(preds, *fp)
	}

	rows, err := r.svc.Query.Property.Units.List(ctx, unitsql.List().Where(unitsql.And(preds...)))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*engine.UnitInfo, 0, len(rows))
	for i := range rows {
		info, err := r.buildUnitInfo(ctx, &rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, nil
}

func (r *AvailabilityReader) money(ctx context.Context, id string) (*money.Money, error) {
	m, err := r.svc.Query.Common.Moneys.Get(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return moneyFromSchema(m), nil
}
