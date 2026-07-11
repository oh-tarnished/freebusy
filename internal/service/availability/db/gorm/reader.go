// Package gorm is the GORM-backed AvailabilityReader: the read-only queries the
// availability engine runs on (unit config + schedule policy, a unit's active
// bookings and closures, and a catalog sweep for search). It converts GORM rows
// into the provider-neutral value types in the parent db package.
package gorm

import (
	"context"
	"errors"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/shared/rrule"
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
			Opens:  repox.Deref(rr.Opens),
			Closes: repox.Deref(rr.Closes),
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

// --- helpers -----------------------------------------------------------------
