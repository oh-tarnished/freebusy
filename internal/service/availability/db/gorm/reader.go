// Package gorm is the GORM-backed AvailabilityReader: the read-only queries the
// availability engine runs on (unit config + schedule policy, a unit's active
// bookings and closures, and a catalog sweep for search). It converts GORM rows
// into the provider-neutral value types in the parent db package.
package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/shared/rrule"
	"google.golang.org/genproto/googleapis/type/money"
	"gorm.io/gorm"
)

// AvailabilityReader is the GORM-backed availability reader.
type AvailabilityReader struct {
	db *gorm.DB
}

// NewAvailabilityReader returns a GORM-backed AvailabilityReader bound to db.
func NewAvailabilityReader(db *gorm.DB) *AvailabilityReader {
	return &AvailabilityReader{db: db}
}

// activeBookingSQL selects the held/confirmed bookings on a unit whose window
// overlaps [start,end), for the engine's free-count math.
const activeBookingSQL = `
SELECT w.start_time AS start, w.end_time AS "end", COALESCE(b.units, 1) AS units
FROM "booking"."resource" b
JOIN "shared"."time_windows" w ON w.id = b.window_id
WHERE b.unit = ? AND b.state IN ('PENDING_HOLD','CONFIRMED')
  AND w.start_time < ? AND w.end_time > ?`

// activeBookingBatchSQL is activeBookingSQL across many units, carrying b.unit so
// the caller can group the rows by unit.
const activeBookingBatchSQL = `
SELECT b.unit AS unit, w.start_time AS start, w.end_time AS "end", COALESCE(b.units, 1) AS units
FROM "booking"."resource" b
JOIN "shared"."time_windows" w ON w.id = b.window_id
WHERE b.unit IN ? AND b.state IN ('PENDING_HOLD','CONFIRMED')
  AND w.start_time < ? AND w.end_time > ?`

func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return types.ErrNotFound
	default:
		return err
	}
}

// unitPreloads returns a query with the unit's price and pricing children (with
// their Money) preloaded — the data the lead-price computation needs.
func (r *AvailabilityReader) unitPreloads(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Preload("Price").
		Preload("Fees").Preload("Fees.Amount").
		Preload("Taxes").
		Preload("LosDiscounts").Preload("LosDiscounts.AmountOff")
}

// GetUnit loads the unit's config and, from its schedule, the stay/notice/buffer
// and open-hours policy.
func (r *AvailabilityReader) GetUnit(ctx context.Context, unitName string) (*engine.UnitInfo, error) {
	unitID, err := types.UnitID(unitName)
	if err != nil {
		return nil, err
	}
	var u property.Unit
	if err := r.unitPreloads(ctx).First(&u, "id = ?", unitID).Error; err != nil {
		return nil, mapErr(err)
	}
	scheduleName, err := types.ScheduleName(u.PropertyID, unitID)
	if err != nil {
		return nil, err
	}
	var sched *schedule.Schedule
	var s schedule.Schedule
	switch err := schedulePreloads(r.db.WithContext(ctx)).First(&s, "name = ?", scheduleName).Error; {
	case err == nil:
		sched = &s
	case errors.Is(err, gorm.ErrRecordNotFound):
		// No schedule configured; policy stays zero-valued.
	default:
		return nil, err
	}
	return buildUnitInfo(&u, sched), nil
}

func schedulePreloads(db *gorm.DB) *gorm.DB {
	return db.Preload("StayConstraints").Preload("Buffers").Preload("RecurringRules")
}

// buildUnitInfo assembles the engine UnitInfo from a preloaded unit and its
// (optional) preloaded schedule.
func buildUnitInfo(u *property.Unit, sched *schedule.Schedule) *engine.UnitInfo {
	info := &engine.UnitInfo{
		ID:           u.ID,
		Name:         u.Name,
		DisplayName:  u.DisplayName,
		Mode:         string(u.BookingMode),
		Capacity:     1,
		TimeZone:     u.TimeZone,
		Duration:     durationFromStr(u.Duration),
		Price:        moneyFromModel(u.Price),
		Archived:     u.State != nil && *u.State == property.UnitStateArchived,
		Fees:         feesOf(u),
		Taxes:        taxesOf(u),
		LosDiscounts: losOf(u),
	}
	if u.Capacity != nil && *u.Capacity > 0 {
		info.Capacity = *u.Capacity
	}
	if sched == nil {
		return info
	}
	if sc := sched.StayConstraints; sc != nil {
		info.MinNights = derefInt32(sc.MinNights)
		info.MaxNights = derefInt32(sc.MaxNights)
		info.CheckinWeekdays = weekdaysFromStr(sc.CheckinWeekdays)
		info.CheckoutWeekdays = weekdaysFromStr(sc.CheckoutWeekdays)
	}
	if b := sched.Buffers; b != nil {
		info.MinNotice = durationFromStr(b.MinNotice)
		info.MaxAdvance = durationFromStr(b.MaxAdvance)
		info.Gap = durationFromStr(b.Gap)
		info.StartDelta = durationFromStr(b.StartDelta)
		info.EndDelta = durationFromStr(b.EndDelta)
	}
	for i := range sched.RecurringRules {
		rr := &sched.RecurringRules[i]
		info.Recurring = append(info.Recurring, rrule.Rule{
			RRule:  rr.Rrule,
			Opens:  deref(rr.Opens),
			Closes: deref(rr.Closes),
		})
	}
	return info
}

// ActiveBookings returns the overlapping held/confirmed reservations on unitID.
func (r *AvailabilityReader) ActiveBookings(ctx context.Context, unitID string, start, end time.Time) ([]engine.Reservation, error) {
	var rows []engine.Reservation
	if err := r.db.WithContext(ctx).Raw(activeBookingSQL, unitID, end.UTC(), start.UTC()).Scan(&rows).Error; err != nil {
		return nil, mapErr(err)
	}
	return rows, nil
}

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
		return nil, types.MapFilterxErr(err)
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

// closureOf converts one exception row into an engine.Closure span in loc.
func closureOf(e *schedule.AvailabilityException, loc *time.Location) (engine.Closure, bool) {
	switch {
	case e.Window != nil:
		return engine.Closure{Start: e.Window.StartTime.UTC(), End: e.Window.EndTime.UTC()}, true
	case e.DateRange != nil:
		s, en := e.DateRange.StartDate, e.DateRange.EndDate
		start := time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, loc)
		end := time.Date(en.Year(), en.Month(), en.Day(), 0, 0, 0, 0, loc)
		return engine.Closure{Start: start.UTC(), End: end.UTC()}, true
	default:
		return engine.Closure{}, false
	}
}

// --- helpers -----------------------------------------------------------------

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

func durationFromStr(s *string) time.Duration {
	if s == nil || *s == "" {
		return 0
	}
	d, err := time.ParseDuration(*s)
	if err != nil {
		return 0
	}
	return d
}

func moneyFromModel(m *common.Money) *money.Money {
	if m == nil {
		return nil
	}
	out := &money.Money{}
	if m.CurrencyCode != nil {
		out.CurrencyCode = *m.CurrencyCode
	}
	if m.Units != nil {
		out.Units = *m.Units
	}
	if m.Nanos != nil {
		out.Nanos = *m.Nanos
	}
	return out
}
