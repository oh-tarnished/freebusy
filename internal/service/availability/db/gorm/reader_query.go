// Bulk query surface: closures, catalog search, and batch lookups.
package gorm

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
)

// Closures returns the unit's CLOSURE exceptions as UTC spans, expanding
// date-range closures to [startDate 00:00, endDate 00:00) in tz.
func (r *AvailabilityReader) Closures(ctx context.Context, unitID, tz string) ([]engine.Closure, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	var rows []schedule.AvailabilityException
	if err := r.db.WithContext(ctx).
		Preload("Window").Preload("DateRange").
		Where("unit_id = ? AND kind = ?", unitID, schedule.ExceptionKindClosure).
		Find(&rows).Error; err != nil {
		return nil, mapErr(err)
	}
	out := make([]engine.Closure, 0, len(rows))
	for i := range rows {
		switch {
		case rows[i].Window != nil:
			out = append(out, engine.Closure{Start: rows[i].Window.StartTime.UTC(), End: rows[i].Window.EndTime.UTC()})
		case rows[i].DateRange != nil:
			s := rows[i].DateRange.StartDate
			e := rows[i].DateRange.EndDate
			start := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, loc)
			end := time.Date(e.Year(), e.Month(), e.Day(), 0, 0, 0, 0, loc)
			out = append(out, engine.Closure{Start: start.UTC(), End: end.UTC()})
		}
	}
	return out, nil
}

// SearchUnits returns active units for the storefront search, scoped and filtered,
// each fully enriched (pricing children + schedule policy) so the handler can
// judge bookability and lead price without a per-unit round trip. Schedules are
// batch-loaded by name in one query.
func (r *AvailabilityReader) SearchUnits(ctx context.Context, propertyRef, organisation, filter string) ([]*engine.UnitInfo, error) {
	q := r.unitPreloads(ctx).
		Where("state IS NULL OR state <> ?", property.UnitStateArchived)

	if propertyRef != "" {
		propertyID, err := types.PropertyID(propertyRef)
		if err != nil {
			return nil, err
		}
		q = q.Where("property_id = ?", propertyID)
	}
	if organisation != "" {
		orgID, err := types.OrganisationID(organisation)
		if err != nil {
			return nil, err
		}
		sub := r.db.Model(&property.Property{}).Select("id").Where("organisation = ?", orgID)
		q = q.Where("property_id IN (?)", sub)
	}
	conds, err := types.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	where, args, err := filterx.Gorm[property.Unit](property.UnitFilterSpec).Where(types.Filterx(conds))
	if err != nil {
		return nil, repox.MapFilterxErr(err)
	}
	if where != "" {
		q = q.Where(where, args...)
	}

	var units []property.Unit
	if err := q.Find(&units).Error; err != nil {
		return nil, mapErr(err)
	}

	// Batch-load the units' schedules by name (one query), keyed for the join.
	names := make([]string, 0, len(units))
	for i := range units {
		if n, err := types.ScheduleName(units[i].PropertyID, units[i].ID); err == nil {
			names = append(names, n)
		}
	}
	schedByName := map[string]*schedule.Schedule{}
	if len(names) > 0 {
		var scheds []schedule.Schedule
		if err := schedulePreloads(r.db.WithContext(ctx)).Where("name IN ?", names).Find(&scheds).Error; err != nil {
			return nil, mapErr(err)
		}
		for i := range scheds {
			schedByName[scheds[i].Name] = &scheds[i]
		}
	}

	out := make([]*engine.UnitInfo, 0, len(units))
	for i := range units {
		name, _ := types.ScheduleName(units[i].PropertyID, units[i].ID)
		out = append(out, buildUnitInfo(&units[i], schedByName[name]))
	}
	return out, nil
}

// ActiveBookingsForUnits batches ActiveBookings across many units: it returns the
// overlapping held/confirmed reservations for each unit id, keyed by unit id.
func (r *AvailabilityReader) ActiveBookingsForUnits(ctx context.Context, unitIDs []string, start, end time.Time) (map[string][]engine.Reservation, error) {
	out := map[string][]engine.Reservation{}
	if len(unitIDs) == 0 {
		return out, nil
	}
	type row struct {
		Unit  string
		Start time.Time
		End   time.Time
		Units int32
	}
	var rows []row
	if err := r.db.WithContext(ctx).Raw(activeBookingBatchSQL, unitIDs, end.UTC(), start.UTC()).Scan(&rows).Error; err != nil {
		return nil, mapErr(err)
	}
	for i := range rows {
		out[rows[i].Unit] = append(out[rows[i].Unit], engine.Reservation{Start: rows[i].Start, End: rows[i].End, Units: rows[i].Units})
	}
	return out, nil
}

// ClosuresForUnits batches Closures across many units, expanding each unit's
// date-range closures in that unit's timezone (from tzByUnit).
func (r *AvailabilityReader) ClosuresForUnits(ctx context.Context, unitIDs []string, tzByUnit map[string]string) (map[string][]engine.Closure, error) {
	out := map[string][]engine.Closure{}
	if len(unitIDs) == 0 {
		return out, nil
	}
	var rows []schedule.AvailabilityException
	if err := r.db.WithContext(ctx).
		Preload("Window").Preload("DateRange").
		Where("unit_id IN ? AND kind = ?", unitIDs, schedule.ExceptionKindClosure).
		Find(&rows).Error; err != nil {
		return nil, mapErr(err)
	}
	for i := range rows {
		loc, err := time.LoadLocation(tzByUnit[rows[i].UnitID])
		if err != nil {
			loc = time.UTC
		}
		if c, ok := closureOf(&rows[i], loc); ok {
			out[rows[i].UnitID] = append(out[rows[i].UnitID], c)
		}
	}
	return out, nil
}
