// Package hasura is the Hasura/GraphQL-backed AvailabilityReader: the same
// read-only queries the availability engine runs on as the GORM reader, expressed
// against the generated freebusyql client. It converts GraphQL rows into the
// provider-neutral engine value types.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/resourceql"
	feesql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/feesql"
	losdiscountsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/losdiscountsql"
	propertiesql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	propertyschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	taxesql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/taxesql"
	unitsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	availexcsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/availabilityexceptionsql"
	recurringsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/recurringrulesql"
	schedresourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/resourceql"
	scheduleschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
	"github.com/oh-tarnished/freebusy/internal/service/booking/pricing"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/shared/rrule"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/genproto/googleapis/type/money"
)

// AvailabilityReader is the Hasura-backed availability reader.
type AvailabilityReader struct {
	svc *freebusyql.Service
}

// NewAvailabilityReader returns a Hasura-backed AvailabilityReader bound to svc.
func NewAvailabilityReader(svc *freebusyql.Service) *AvailabilityReader {
	return &AvailabilityReader{svc: svc}
}

// GetUnit loads the unit's config, pricing children, and schedule policy.
func (r *AvailabilityReader) GetUnit(ctx context.Context, unitName string) (*engine.UnitInfo, error) {
	unitID, err := types.UnitID(unitName)
	if err != nil {
		return nil, err
	}
	u, err := r.svc.Query.Property.Units.Get(ctx, unitID)
	if err != nil {
		return nil, mapErr(err)
	}
	if u == nil {
		return nil, types.ErrNotFound
	}
	return r.buildUnitInfo(ctx, u)
}

// buildUnitInfo assembles the engine UnitInfo for one unit row.
func (r *AvailabilityReader) buildUnitInfo(ctx context.Context, u *propertyschema.PropertyUnits) (*engine.UnitInfo, error) {
	info := &engine.UnitInfo{
		ID:          u.Id,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Mode:        u.BookingMode,
		Capacity:    1,
		TimeZone:    u.TimeZone,
		Duration:    durationFromStr(repox.Deref(u.Duration)),
		Archived:    u.State != nil && *u.State == "ARCHIVED",
	}
	if u.Capacity != nil && *u.Capacity > 0 {
		info.Capacity = *u.Capacity
	}
	// price + pricing children
	price, fees, taxes, los, err := r.pricing(ctx, u)
	if err != nil {
		return nil, err
	}
	info.Price, info.Fees, info.Taxes, info.LosDiscounts = price, fees, taxes, los

	// schedule policy
	scheduleName, err := types.ScheduleName(u.PropertyId, u.Id)
	if err != nil {
		return nil, err
	}
	sched, err := r.svc.Query.Schedule.Resource.Find(ctx, schedresourceql.List().Where(schedresourceql.Name.Eq(scheduleName)))
	if err != nil {
		return nil, mapErr(err)
	}
	if sched != nil {
		if err := r.applySchedulePolicy(ctx, info, sched); err != nil {
			return nil, err
		}
	}
	return info, nil
}

// pricing loads a unit's base price and pricing children as engine inputs.
func (r *AvailabilityReader) pricing(ctx context.Context, u *propertyschema.PropertyUnits) (*money.Money, []pricing.Fee, []pricing.Tax, []pricing.LosDiscount, error) {
	var price *money.Money
	if u.PriceId != nil {
		m, err := r.money(ctx, *u.PriceId)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		price = m
	}
	feeRows, err := r.svc.Query.Property.Fees.List(ctx, feesql.List().Where(feesql.UnitId.Eq(u.Id)))
	if err != nil {
		return nil, nil, nil, nil, mapErr(err)
	}
	fees := make([]pricing.Fee, 0, len(feeRows))
	for i := range feeRows {
		f := &feeRows[i]
		var amt *money.Money
		if f.AmountId != nil {
			if amt, err = r.money(ctx, *f.AmountId); err != nil {
				return nil, nil, nil, nil, err
			}
		}
		fees = append(fees, pricing.Fee{Code: f.Code, DisplayName: repox.Deref(f.DisplayName), PricingUnit: repox.Deref(f.PricingUnit), Percent: f.Percent, Amount: amt, Taxable: repox.Deref(f.Taxable)})
	}
	taxRows, err := r.svc.Query.Property.Taxes.List(ctx, taxesql.List().Where(taxesql.UnitId.Eq(u.Id)))
	if err != nil {
		return nil, nil, nil, nil, mapErr(err)
	}
	taxes := make([]pricing.Tax, 0, len(taxRows))
	for i := range taxRows {
		taxes = append(taxes, pricing.Tax{Code: taxRows[i].Code, DisplayName: repox.Deref(taxRows[i].DisplayName), Percent: taxRows[i].Percent})
	}
	losRows, err := r.svc.Query.Property.LosDiscounts.List(ctx, losdiscountsql.List().Where(losdiscountsql.UnitId.Eq(u.Id)))
	if err != nil {
		return nil, nil, nil, nil, mapErr(err)
	}
	los := make([]pricing.LosDiscount, 0, len(losRows))
	for i := range losRows {
		d := &losRows[i]
		var amt *money.Money
		if d.AmountOffId != nil {
			if amt, err = r.money(ctx, *d.AmountOffId); err != nil {
				return nil, nil, nil, nil, err
			}
		}
		los = append(los, pricing.LosDiscount{MinNights: d.MinNights, PercentOff: d.PercentOff, AmountOff: amt})
	}
	return price, fees, taxes, los, nil
}

// applySchedulePolicy fills the stay/notice/buffer/open-hours policy from a
// schedule row.
func (r *AvailabilityReader) applySchedulePolicy(ctx context.Context, info *engine.UnitInfo, sched *scheduleschema.ScheduleResource) error {
	if sched.StayConstraintsId != nil {
		sc, err := r.svc.Query.Schedule.StayConstraints.Get(ctx, *sched.StayConstraintsId)
		if err != nil {
			return mapErr(err)
		}
		if sc != nil {
			info.MinNights = derefInt32(sc.MinNights)
			info.MaxNights = derefInt32(sc.MaxNights)
			info.CheckinWeekdays = weekdaysFromStr(sc.CheckinWeekdays)
			info.CheckoutWeekdays = weekdaysFromStr(sc.CheckoutWeekdays)
		}
	}
	if sched.BuffersId != nil {
		b, err := r.svc.Query.Schedule.BufferSettings.Get(ctx, *sched.BuffersId)
		if err != nil {
			return mapErr(err)
		}
		if b != nil {
			info.MinNotice = durationFromStr(repox.Deref(b.MinNotice))
			info.MaxAdvance = durationFromStr(repox.Deref(b.MaxAdvance))
			info.Gap = durationFromStr(repox.Deref(b.Gap))
			info.StartDelta = durationFromStr(repox.Deref(b.StartDelta))
			info.EndDelta = durationFromStr(repox.Deref(b.EndDelta))
		}
	}
	rules, err := r.svc.Query.Schedule.RecurringRules.List(ctx, recurringsql.List().Where(recurringsql.ScheduleId.Eq(sched.Id)))
	if err != nil {
		return mapErr(err)
	}
	for i := range rules {
		info.Recurring = append(info.Recurring, rrule.Rule{RRule: rules[i].Rrule, Opens: repox.Deref(rules[i].Opens), Closes: repox.Deref(rules[i].Closes)})
	}
	return nil
}

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
	rows, err := r.svc.Query.Schedule.AvailabilityExceptions.List(ctx, availexcsql.List().Where(availexcsql.And(
		availexcsql.UnitId.In(unitIDs...),
		availexcsql.Kind.Eq("CLOSURE"),
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
