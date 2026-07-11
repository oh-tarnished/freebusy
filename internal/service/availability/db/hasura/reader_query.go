// Bulk query surface: bookings, closures, and catalog search over DDN.
package hasura

import (
	"context"
	"time"

	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/resourceql"
	propertiesql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	unitsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/money"
)

// ActiveBookings returns the overlapping held/confirmed reservations on unitID.
func (r *AvailabilityReader) ActiveBookings(ctx context.Context, unitID string, start, end time.Time) ([]engine.Reservation, error) {
	m, err := r.activeBookings(ctx, []string{unitID}, start, end)
	if err != nil {
		return nil, err
	}
	return m[unitID], nil
}

// ActiveBookingsForUnits batches ActiveBookings across many units.
func (r *AvailabilityReader) ActiveBookingsForUnits(ctx context.Context, unitIDs []string, start, end time.Time) (map[string][]engine.Reservation, error) {
	return r.activeBookings(ctx, unitIDs, start, end)
}

func (r *AvailabilityReader) activeBookings(ctx context.Context, unitIDs []string, start, end time.Time) (map[string][]engine.Reservation, error) {
	out := map[string][]engine.Reservation{}
	if len(unitIDs) == 0 {
		return out, nil
	}
	rows, err := r.svc.Query.Booking.Resource.List(ctx, resourceql.List().Where(resourceql.And(
		resourceql.Unit.In(unitIDs...),
		resourceql.State.In("PENDING_HOLD", "CONFIRMED"),
	)))
	if err != nil {
		return nil, mapErr(err)
	}
	for i := range rows {
		if rows[i].WindowId == "" {
			continue
		}
		w, err := r.svc.Query.Shared.TimeWindows.Get(ctx, rows[i].WindowId)
		if err != nil {
			return nil, mapErr(err)
		}
		if w == nil {
			continue
		}
		ws, we := parseTS(w.StartTime), parseTS(w.EndTime)
		if ws.Before(end) && we.After(start) {
			u := int32(1)
			if rows[i].Units != nil && *rows[i].Units > 0 {
				u = *rows[i].Units
			}
			out[rows[i].Unit] = append(out[rows[i].Unit], engine.Reservation{Start: ws, End: we, Units: u})
		}
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
	for i := range rows {
		e := &rows[i]
		loc, lerr := time.LoadLocation(tzByUnit[e.UnitId])
		if lerr != nil {
			loc = time.UTC
		}
		switch {
		case e.WindowId != nil:
			w, err := r.svc.Query.Shared.TimeWindows.Get(ctx, *e.WindowId)
			if err != nil {
				return nil, mapErr(err)
			}
			if w != nil {
				out[e.UnitId] = append(out[e.UnitId], engine.Closure{Start: parseTS(w.StartTime), End: parseTS(w.EndTime)})
			}
		case e.DateRangeId != nil:
			d, err := r.svc.Query.Shared.DateRanges.Get(ctx, *e.DateRangeId)
			if err != nil {
				return nil, mapErr(err)
			}
			if d != nil {
				start := parseDate(d.StartDate, loc)
				end := parseDate(d.EndDate, loc)
				out[e.UnitId] = append(out[e.UnitId], engine.Closure{Start: start.UTC(), End: end.UTC()})
			}
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
