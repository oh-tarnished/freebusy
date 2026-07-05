package hasura

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/resourceql"
	bookingschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/schemaql"
	moneysql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	guestsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/guestsql"
	refundtiersql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/refundtiersql"
	schedresourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/scheduleql/resourceql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/service/booking/pricing"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ExpireHolds flips every PENDING_HOLD booking whose hold has lapsed to EXPIRED,
// freeing the capacity it reserved. Returns the number of holds expired. Timestamps
// are stored as fixed-width RFC 3339 UTC strings, so a lexical `<` matches
// chronological order. Intended to be called periodically by the hold sweeper.
func (r *BookingRepository) ExpireHolds(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	nowStr := tsToStr(timestamppb.New(now))
	rows, err := r.svc.Query.Booking.Resource.List(ctx, resourceql.List().Where(resourceql.And(
		resourceql.State.Eq("PENDING_HOLD"),
		resourceql.HoldExpireTime.Lt(nowStr),
	)))
	if err != nil {
		return 0, mapHasuraErr(err)
	}
	var expired int64
	for i := range rows {
		patch := resourceql.UpdateInput{
			State:      graphql.Value("EXPIRED"),
			Etag:       graphql.Value(ulid.GenerateString()),
			UpdateTime: graphql.Value(nowStr),
		}
		if _, err := r.svc.Mutation.Booking.Resource.Update(ctx, rows[i].Id, patch); err != nil {
			return expired, mapHasuraErr(err)
		}
		expired++
	}
	return expired, nil
}

// UpdateBookingGuests replaces the whole staying party (guests + occupancy) on a
// booking, allowed only while PENDING_HOLD or CONFIRMED and re-validated against
// the unit's max occupancy. The old occupancy and guest graph are dropped and the
// new ones inserted in one atomic mutation batch.
func (r *BookingRepository) UpdateBookingGuests(ctx context.Context, name string, guests []*identitypbv1.Guest, occupancy *bookingpbv1.Occupancy) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Booking.Resource.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if res.State == nil || (*res.State != "PENDING_HOLD" && *res.State != "CONFIRMED") {
		return nil, types.ErrConflict
	}

	// Re-validate the party against the unit's max occupancy.
	unit, err := r.svc.Query.Property.Units.Get(ctx, res.Unit)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if unit == nil {
		return nil, types.ErrNotFound
	}
	units := deref(res.Units)
	if units < 1 {
		units = 1
	}
	if unit.MaxOccupancy != nil && *unit.MaxOccupancy > 0 && partySize(occupancy, guests) > *unit.MaxOccupancy*units {
		return nil, types.ErrInvalidArgument
	}

	oldGuests, err := r.svc.Query.Identity.Guests.List(ctx, guestsql.List().Where(guestsql.BookingId.Eq(id)))
	if err != nil {
		return nil, mapHasuraErr(err)
	}

	now := time.Now().UTC()
	newOcc := occupancyInput(occupancy)
	graphs := make([]guestGraph, 0, len(guests))
	for _, g := range guests {
		graphs = append(graphs, buildGuestGraph(g, id))
	}

	tx := r.svc.Mutation.Tx()
	occID := ""
	if newOcc != nil {
		var oRes bookingschema.InsertBookingOccupanciesResponse
		tx.Add(r.svc.Mutation.Booking.Occupancies.CreateOp(*newOcc, &oRes))
		occID = newOcc.Id
	}
	patch := resourceql.UpdateInput{
		OccupancyId: nullableStr(occID),
		Etag:        graphql.Value(ulid.GenerateString()),
		UpdateTime:  graphql.Value(tsToStr(timestamppb.New(now))),
	}
	var updRes bookingschema.UpdateBookingResourceByIdResponse
	tx.Add(r.svc.Mutation.Booking.Resource.UpdateOp(id, patch, &updRes))
	if res.OccupancyId != nil {
		var delOcc bookingschema.DeleteBookingOccupanciesByIdResponse
		tx.Add(r.svc.Mutation.Booking.Occupancies.DeleteOp(*res.OccupancyId, &delOcc))
	}
	queueGuestDeletes(tx, r, oldGuests)
	queueGuestInserts(tx, r, graphs)
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetBooking(ctx, name)
}

// ConfirmBooking flips a PENDING_HOLD booking to CONFIRMED.
func (r *BookingRepository) ConfirmBooking(ctx context.Context, name string) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Booking.Resource.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if res.State == nil || *res.State != "PENDING_HOLD" {
		return nil, types.ErrConflict
	}
	now := time.Now().UTC()
	patch := resourceql.UpdateInput{
		State:          graphql.Value("CONFIRMED"),
		ConfirmTime:    graphql.Value(tsToStr(timestamppb.New(now))),
		HoldExpireTime: graphql.Null[string](),
		Etag:           graphql.Value(ulid.GenerateString()),
		UpdateTime:     graphql.Value(tsToStr(timestamppb.New(now))),
	}
	if _, err := r.svc.Mutation.Booking.Resource.Update(ctx, id, patch); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetBooking(ctx, name)
}

// CancelBooking flips a held or confirmed booking to CANCELLED, computing the
// refund from the unit's cancellation policy, in one batch.
func (r *BookingRepository) CancelBooking(ctx context.Context, name string, reason bookingpbv1.CancelReason) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Booking.Resource.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if res.State != nil && (*res.State == "CANCELLED" || *res.State == "EXPIRED") {
		return nil, types.ErrConflict
	}
	pct, amount, _, err := r.computeRefund(ctx, res)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	patch := resourceql.UpdateInput{
		State:         graphql.Value("CANCELLED"),
		CancelTime:    graphql.Value(tsToStr(timestamppb.New(now))),
		CancelReason:  nullableStr(cancelReasonToStr(reason)),
		RefundPercent: graphql.Value(pct),
		Etag:          graphql.Value(ulid.GenerateString()),
		UpdateTime:    graphql.Value(tsToStr(timestamppb.New(now))),
	}
	tx := r.svc.Mutation.Tx()
	if amount != nil {
		mi := moneyInput(amount)
		var mRes commonschema.InsertCommonMoneysResponse
		tx.Add(r.svc.Mutation.Common.Moneys.CreateOp(mi, &mRes))
		patch.RefundAmountId = graphql.Value(mi.Id)
	}
	var updRes bookingschema.UpdateBookingResourceByIdResponse
	tx.Add(r.svc.Mutation.Booking.Resource.UpdateOp(id, patch, &updRes))
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetBooking(ctx, name)
}

// PreviewCancellation computes the refund a cancellation would yield now, without
// cancelling.
func (r *BookingRepository) PreviewCancellation(ctx context.Context, name string) (refundable bool, percent int32, amount, nonRefundable *money.Money, summary string, err error) {
	id, err := types.BookingID(name)
	if err != nil {
		return false, 0, nil, nil, "", err
	}
	res, err := r.svc.Query.Booking.Resource.Get(ctx, id)
	if err != nil {
		return false, 0, nil, nil, "", mapHasuraErr(err)
	}
	if res == nil {
		return false, 0, nil, nil, "", types.ErrNotFound
	}
	percent, amount, summary, err = r.computeRefund(ctx, res)
	if err != nil {
		return false, 0, nil, nil, "", err
	}
	var total *money.Money
	if res.TotalId != nil {
		if total, err = r.money(ctx, *res.TotalId); err != nil {
			return false, 0, nil, nil, "", err
		}
	}
	nonRefundable = moneySub(total, amount)
	return percent > 0, percent, amount, nonRefundable, summary, nil
}

// RescheduleBooking atomically moves a booking to a new window (and optionally a
// new unit), re-checking capacity and recomputing the full price breakdown.
func (r *BookingRepository) RescheduleBooking(ctx context.Context, name string, w *bookingpbv1.Booking, newUnit string) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	if w.GetWindow() == nil {
		return nil, types.ErrInvalidArgument
	}
	res, err := r.svc.Query.Booking.Resource.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	unitID := res.Unit
	if newUnit != "" {
		_, uid, perr := types.ParseUnitParent(newUnit)
		if perr != nil {
			return nil, perr
		}
		unitID = uid
	}
	unit, err := r.svc.Query.Property.Units.Get(ctx, unitID)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if unit == nil {
		return nil, types.ErrNotFound
	}

	requested := int64(deref(res.Units))
	if requested < 1 {
		requested = 1
	}
	reserved, err := r.reservedUnits(ctx, unitID, w.GetWindow(), id)
	if err != nil {
		return nil, err
	}
	capacity := int64(1)
	if unit.Capacity != nil && *unit.Capacity > 0 {
		capacity = int64(*unit.Capacity)
	}
	if reserved+requested > capacity {
		return nil, types.ErrConflict
	}

	in, err := r.pricingInputs(ctx, unit, deref(res.PromoCode))
	if err != nil {
		return nil, err
	}
	in.Nights = nightsBetween(w.GetWindow(), unit.TimeZone)
	in.Units = requested

	var priceIn, discountIn, totalIn *moneysql.CreateInput
	var components []*sharedpbv1.PriceComponent
	priceID, discountID, totalID := "", "", ""
	if in.Price != nil {
		result := pricing.Compute(in, unitID)
		pi := moneyInput(result.Base)
		priceIn = &pi
		priceID = pi.Id
		ti := moneyInput(result.Total)
		totalIn = &ti
		totalID = ti.Id
		if !pricing.IsZero(result.Discount) {
			di := moneyInput(result.Discount)
			discountIn = &di
			discountID = di.Id
		}
		components = result.Components
	}

	now := time.Now().UTC()
	window := windowInput(w.GetWindow())
	patch := resourceql.UpdateInput{
		Unit:       graphql.Value(unitID),
		WindowId:   graphql.Value(window.Id),
		PriceId:    nullableStr(priceID),
		DiscountId: nullableStr(discountID),
		TotalId:    nullableStr(totalID),
		Etag:       graphql.Value(ulid.GenerateString()),
		UpdateTime: graphql.Value(tsToStr(timestamppb.New(now))),
	}

	tx := r.svc.Mutation.Tx()
	var winRes sharedschema.InsertSharedTimeWindowsResponse
	tx.Add(r.svc.Mutation.Shared.TimeWindows.CreateOp(window, &winRes))
	queueMoneyInserts(tx, r, priceIn, discountIn, totalIn)
	var updRes bookingschema.UpdateBookingResourceByIdResponse
	tx.Add(r.svc.Mutation.Booking.Resource.UpdateOp(id, patch, &updRes))
	// Drop the superseded window and old price/discount/total Money rows.
	var delW sharedschema.DeleteSharedTimeWindowsByIdResponse
	tx.Add(r.svc.Mutation.Shared.TimeWindows.DeleteOp(res.WindowId, &delW))
	for _, mid := range []*string{res.PriceId, res.DiscountId, res.TotalId} {
		if mid != nil {
			var delM commonschema.DeleteCommonMoneysByIdResponse
			tx.Add(r.svc.Mutation.Common.Moneys.DeleteOp(*mid, &delM))
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	out, err := r.GetBooking(ctx, name)
	if err != nil {
		return nil, err
	}
	out.PriceComponents = components
	return out, nil
}

// computeRefund resolves the unit's cancellation policy (from its schedule) and
// returns the refund percent, amount, and a human summary for the booking's lead
// time. No matching tier (or no policy) means non-refundable.
func (r *BookingRepository) computeRefund(ctx context.Context, res *bookingschema.BookingResource) (int32, *money.Money, string, error) {
	if res.TotalId == nil {
		return 0, nil, "non-refundable", nil
	}
	total, err := r.money(ctx, *res.TotalId)
	if err != nil {
		return 0, nil, "", err
	}
	unit, err := r.svc.Query.Property.Units.Get(ctx, res.Unit)
	if err != nil {
		return 0, nil, "", mapHasuraErr(err)
	}
	if unit == nil {
		return 0, nil, "non-refundable", nil
	}
	scheduleName, err := types.ScheduleName(unit.PropertyId, res.Unit)
	if err != nil {
		return 0, nil, "", err
	}
	sched, err := r.svc.Query.Schedule.Resource.Find(ctx, schedresourceql.List().Where(schedresourceql.Name.Eq(scheduleName)))
	if err != nil {
		return 0, nil, "", mapHasuraErr(err)
	}
	if sched == nil || sched.CancellationPolicyId == nil {
		return 0, nil, "non-refundable (no cancellation policy)", nil
	}
	tiers, err := r.svc.Query.Schedule.RefundTiers.List(ctx,
		refundtiersql.List().Where(refundtiersql.CancellationPolicyId.Eq(*sched.CancellationPolicyId)))
	if err != nil {
		return 0, nil, "", mapHasuraErr(err)
	}
	if len(tiers) == 0 {
		return 0, nil, "non-refundable", nil
	}

	var lead time.Duration
	if res.WindowId != "" {
		if w, werr := r.svc.Query.Shared.TimeWindows.Get(ctx, res.WindowId); werr == nil && w != nil {
			if st := strToTS(w.StartTime); st != nil {
				lead = time.Until(st.AsTime())
			}
		}
	}
	// The satisfied tier with the largest cutoff wins (cancelled at least cutoff
	// before the booking start).
	var bestPct int32
	bestCutoff := time.Duration(-1)
	for i := range tiers {
		cutoff, perr := time.ParseDuration(tiers[i].Cutoff)
		if perr != nil {
			continue
		}
		if lead >= cutoff && cutoff > bestCutoff {
			bestCutoff = cutoff
			bestPct = tiers[i].RefundPercent
		}
	}
	return bestPct, moneyPct(total, bestPct), fmt.Sprintf("%d%% refund for the applicable tier", bestPct), nil
}

// nullableStr maps an empty optional string to a SQL NULL update and a non-empty
// one to a value update.
func nullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}

// moneyPct returns pct percent of m.
func moneyPct(m *money.Money, pct int32) *money.Money {
	if m == nil {
		return nil
	}
	total := (m.GetUnits()*1_000_000_000 + int64(m.GetNanos())) * int64(pct) / 100
	return &money.Money{CurrencyCode: m.GetCurrencyCode(), Units: total / 1_000_000_000, Nanos: int32(total % 1_000_000_000)}
}

// moneySub returns a − b (used to split a total into refundable / retained).
func moneySub(a, b *money.Money) *money.Money {
	if a == nil {
		return nil
	}
	if b == nil {
		return a
	}
	total := (a.GetUnits()*1_000_000_000 + int64(a.GetNanos())) - (b.GetUnits()*1_000_000_000 + int64(b.GetNanos()))
	return &money.Money{CurrencyCode: a.GetCurrencyCode(), Units: total / 1_000_000_000, Nanos: int32(total % 1_000_000_000)}
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
