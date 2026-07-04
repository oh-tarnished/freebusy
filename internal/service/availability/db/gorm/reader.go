// Package gorm is the GORM-backed AvailabilityReader: the read-only queries the
// availability engine runs on (unit config + schedule policy, a unit's active
// bookings and closures, and a catalog sweep for search). It converts GORM rows
// into the provider-neutral value types in the parent db package.
package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/types"
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

// GetUnit loads the unit's config and, from its schedule, the stay/notice policy.
func (r *AvailabilityReader) GetUnit(ctx context.Context, unitName string) (*engine.UnitInfo, error) {
	unitID, err := types.UnitID(unitName)
	if err != nil {
		return nil, err
	}
	var u property.Unit
	if err := r.db.WithContext(ctx).Preload("Price").First(&u, "id = ?", unitID).Error; err != nil {
		return nil, mapErr(err)
	}
	info := &engine.UnitInfo{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Mode:        string(u.BookingMode),
		Capacity:    1,
		TimeZone:    u.TimeZone,
		Duration:    durationFromStr(u.Duration),
		Price:       moneyFromModel(u.Price),
		Archived:    u.State != nil && *u.State == property.UnitStateArchived,
	}
	if u.Capacity != nil && *u.Capacity > 0 {
		info.Capacity = *u.Capacity
	}

	scheduleName, err := types.ScheduleName(u.PropertyID, unitID)
	if err != nil {
		return nil, err
	}
	var sched schedule.Schedule
	switch err := r.db.WithContext(ctx).Preload("StayConstraints").Preload("Buffers").First(&sched, "name = ?", scheduleName).Error; {
	case err == nil:
		if sc := sched.StayConstraints; sc != nil {
			info.MinNights = derefInt32(sc.MinNights)
			info.MaxNights = derefInt32(sc.MaxNights)
		}
		if b := sched.Buffers; b != nil {
			info.MinNotice = durationFromStr(b.MinNotice)
			info.MaxAdvance = durationFromStr(b.MaxAdvance)
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// No schedule configured; leave policy zero-valued.
	default:
		return nil, err
	}
	return info, nil
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

// SearchUnits returns active units for the storefront search, scoped and filtered.
func (r *AvailabilityReader) SearchUnits(ctx context.Context, propertyRef, organisation, filter string) ([]*engine.UnitInfo, error) {
	q := r.db.WithContext(ctx).Preload("Price").
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
	q, err := applyUnitFilter(q, filter)
	if err != nil {
		return nil, err
	}

	var units []property.Unit
	if err := q.Find(&units).Error; err != nil {
		return nil, mapErr(err)
	}
	out := make([]*engine.UnitInfo, 0, len(units))
	for i := range units {
		u := &units[i]
		info := &engine.UnitInfo{
			ID:          u.ID,
			Name:        u.Name,
			DisplayName: u.DisplayName,
			Mode:        string(u.BookingMode),
			Capacity:    1,
			TimeZone:    u.TimeZone,
			Duration:    durationFromStr(u.Duration),
			Price:       moneyFromModel(u.Price),
		}
		if u.Capacity != nil && *u.Capacity > 0 {
			info.Capacity = *u.Capacity
		}
		out = append(out, info)
	}
	return out, nil
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
