package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"gorm.io/gorm"
)

// ExpireHolds flips every PENDING_HOLD booking whose hold has lapsed to EXPIRED,
// freeing the capacity it reserved. Returns the number of holds expired. Intended
// to be called periodically by the hold sweeper.
func (r *BookingRepository) ExpireHolds(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&booking.Booking{}).
		Where("state = ? AND hold_expire_time IS NOT NULL AND hold_expire_time < ?", booking.BookingStatePendingHold, now).
		Updates(map[string]any{
			"state":       booking.BookingStateExpired,
			"etag":        ulid.GenerateString(),
			"update_time": now,
		})
	if res.Error != nil {
		return 0, mapGormErr(res.Error)
	}
	return res.RowsAffected, nil
}

// UpdateBookingGuests replaces the whole staying party (guests + occupancy) on a
// booking. It is allowed only while the booking is PENDING_HOLD or CONFIRMED, and
// re-validates the new party against the unit's max occupancy. Old guest rows and
// their sub-rows, and the old occupancy, are removed in the same transaction.
func (r *BookingRepository) UpdateBookingGuests(ctx context.Context, name string, guests []*identitypbv1.Guest, occupancy *bookingpbv1.Occupancy) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var m booking.Booking
		if e := tx.WithContext(ctx).First(&m, "id = ?", id).Error; e != nil {
			return e
		}
		if m.State == nil || (*m.State != booking.BookingStatePendingHold && *m.State != booking.BookingStateConfirmed) {
			return types.ErrConflict
		}

		// Re-validate the party against the unit's max occupancy.
		var unit property.Unit
		if e := tx.WithContext(ctx).Select("id", "max_occupancy").First(&unit, "id = ?", m.UnitID).Error; e != nil {
			return e
		}
		units := deref(m.Units)
		if units < 1 {
			units = 1
		}
		if maxOcc := deref(unit.MaxOccupancy); maxOcc > 0 && partySize(occupancy, guests) > maxOcc*units {
			return types.ErrInvalidArgument
		}

		// Drop the old party, then repoint the occupancy and insert the new party.
		if e := deleteBookingGuests(ctx, tx, id); e != nil {
			return e
		}
		oldOccID := m.OccupancyID
		newOcc := occupancyToModel(occupancy)
		if newOcc != nil {
			if e := booking.NewOccupancyStore(tx).Create(ctx, newOcc); e != nil {
				return e
			}
			m.OccupancyID = &newOcc.ID
		} else {
			m.OccupancyID = nil
		}
		m.Etag = ptr(ulid.GenerateString())
		m.Contact, m.Window, m.Price, m.Discount, m.Total, m.RefundAmount, m.Occupancy = nil, nil, nil, nil, nil, nil, nil
		if e := booking.NewBookingStore(tx).Update(ctx, &m); e != nil {
			return e
		}
		if oldOccID != nil {
			if e := booking.NewOccupancyStore(tx).DeleteByID(ctx, *oldOccID); e != nil {
				return e
			}
		}
		graphs := make([]guestGraph, 0, len(guests))
		for _, g := range guests {
			graphs = append(graphs, buildGuestGraph(g, id))
		}
		return persistGuests(ctx, tx, graphs)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetBooking(ctx, name)
}

// ConfirmBooking flips a PENDING_HOLD booking to CONFIRMED.
func (r *BookingRepository) ConfirmBooking(ctx context.Context, name string) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var m booking.Booking
		if e := tx.WithContext(ctx).First(&m, "id = ?", id).Error; e != nil {
			return e
		}
		if m.State == nil || *m.State != booking.BookingStatePendingHold {
			return types.ErrConflict
		}
		now := time.Now().UTC()
		state := booking.BookingStateConfirmed
		m.State = &state
		m.ConfirmTime = &now
		m.HoldExpireTime = nil
		m.Etag = ptr(ulid.GenerateString())
		return booking.NewBookingStore(tx).Update(ctx, &m)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetBooking(ctx, name)
}

// CancelBooking flips a held or confirmed booking to CANCELLED, computing the
// refund from the unit's cancellation policy.
func (r *BookingRepository) CancelBooking(ctx context.Context, name string, reason bookingpbv1.CancelReason) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var m booking.Booking
		if e := preloadBooking(tx.WithContext(ctx)).First(&m, "id = ?", id).Error; e != nil {
			return e
		}
		if m.State != nil && (*m.State == booking.BookingStateCancelled || *m.State == booking.BookingStateExpired) {
			return types.ErrConflict
		}
		pct, amount, _, e := r.computeRefund(ctx, tx, &m)
		if e != nil {
			return e
		}
		now := time.Now().UTC()
		state := booking.BookingStateCancelled
		m.State = &state
		m.CancelTime = &now
		m.CancelReason = cancelReasonToModel(reason)
		m.RefundPercent = ptr(pct)
		m.Etag = ptr(ulid.GenerateString())
		m.Contact, m.Window, m.Price, m.Discount, m.Total, m.RefundAmount = nil, nil, nil, nil, nil, nil
		if amount != nil {
			refund := moneyToModel(amount)
			if e := common.NewMoneyStore(tx).Create(ctx, refund); e != nil {
				return e
			}
			m.RefundAmountID = &refund.ID
		}
		return booking.NewBookingStore(tx).Update(ctx, &m)
	})
	if err != nil {
		return nil, mapGormErr(err)
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
	var m booking.Booking
	if err = preloadBooking(r.db.WithContext(ctx)).First(&m, "id = ?", id).Error; err != nil {
		return false, 0, nil, nil, "", mapGormErr(err)
	}
	percent, amount, summary, err = r.computeRefund(ctx, r.db.WithContext(ctx), &m)
	if err != nil {
		return false, 0, nil, nil, "", mapGormErr(err)
	}
	total := moneyFromModel(m.Total)
	nonRefundable = moneySub(total, amount)
	return percent > 0, percent, amount, nonRefundable, summary, nil
}

// RescheduleBooking atomically moves a booking to a new window (and optionally a
// new unit), re-checking capacity and recomputing the base price.
func (r *BookingRepository) RescheduleBooking(ctx context.Context, name string, w *bookingpbv1.Booking, newUnit string) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	if w.GetWindow() == nil {
		return nil, types.ErrInvalidArgument
	}
	var components []*sharedpbv1.PriceComponent
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var m booking.Booking
		if e := preloadBooking(tx.WithContext(ctx)).First(&m, "id = ?", id).Error; e != nil {
			return e
		}
		unitID := m.UnitID
		if newUnit != "" {
			_, uid, perr := types.ParseUnitParent(newUnit)
			if perr != nil {
				return perr
			}
			unitID = uid
		}
		var unit property.Unit
		if e := tx.WithContext(ctx).
			Preload("Price").
			Preload("Fees").Preload("Fees.Amount").
			Preload("Taxes").
			Preload("LosDiscounts").Preload("LosDiscounts.AmountOff").
			First(&unit, "id = ?", unitID).Error; e != nil {
			return e
		}
		var promo *promocode.PromoCode
		if pid := deref(m.PromoCodeID); pid != "" {
			var p promocode.PromoCode
			if e := tx.WithContext(ctx).
				Preload("Discount").Preload("Discount.AmountOff").
				Preload("Scope").Preload("Scope.MinSubtotal").Preload("Scope.ScopeApplicableUnits").
				First(&p, "id = ?", pid).Error; e != nil {
				return e
			}
			promo = &p
		}
		window := timeWindowToModel(w.GetWindow())

		// Capacity check on the new window/unit, excluding this booking.
		var reserved int64
		if e := tx.WithContext(ctx).Raw(overlapSQL+" AND b.id <> ?", unitID, window.EndTime, window.StartTime, id).Scan(&reserved).Error; e != nil {
			return e
		}
		capacity := int64(1)
		if unit.Capacity != nil && *unit.Capacity > 0 {
			capacity = int64(*unit.Capacity)
		}
		requested := deref(m.Units)
		if requested < 1 {
			requested = 1
		}
		if reserved+int64(requested) > capacity {
			return types.ErrConflict
		}

		if e := shared.NewTimeWindowStore(tx).Create(ctx, window); e != nil {
			return e
		}
		oldWindowID := m.WindowID
		oldPriceID, oldDiscountID, oldTotalID := m.PriceID, m.DiscountID, m.TotalID

		// Recompute the full price breakdown for the new window/unit (base, LOS +
		// promo discounts, fees, taxes), carrying the booking's promo code.
		var priceID, discountID, totalID *string
		if unit.Price != nil {
			nights := nightsBetween(w.GetWindow(), unit.TimeZone)
			p := computePricing(&unit, nights, int64(requested), promo)
			price := moneyToModel(p.base)
			total := moneyToModel(p.total)
			moneys := common.NewMoneyStore(tx)
			if e := moneys.Create(ctx, price); e != nil {
				return e
			}
			if e := moneys.Create(ctx, total); e != nil {
				return e
			}
			priceID, totalID = &price.ID, &total.ID
			if !isZeroMoney(p.discount) {
				discount := moneyToModel(p.discount)
				if e := moneys.Create(ctx, discount); e != nil {
					return e
				}
				discountID = &discount.ID
			}
			components = p.components
		}

		m.UnitID = unitID
		m.WindowID = window.ID
		m.PriceID = priceID
		m.DiscountID = discountID
		m.TotalID = totalID
		m.Etag = ptr(ulid.GenerateString())
		m.Contact, m.Window, m.Price, m.Discount, m.Total, m.RefundAmount = nil, nil, nil, nil, nil, nil
		if e := booking.NewBookingStore(tx).Update(ctx, &m); e != nil {
			return e
		}

		// Drop the superseded window (cascade would remove the booking, so delete
		// only after repointing) and the old price/discount/total Money rows.
		if e := shared.NewTimeWindowStore(tx).DeleteByID(ctx, oldWindowID); e != nil {
			return e
		}
		moneys := common.NewMoneyStore(tx)
		for _, mid := range []*string{oldPriceID, oldDiscountID, oldTotalID} {
			if mid != nil {
				if e := moneys.DeleteByID(ctx, *mid); e != nil {
					return e
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapGormErr(err)
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
func (r *BookingRepository) computeRefund(ctx context.Context, tx *gorm.DB, m *booking.Booking) (int32, *money.Money, string, error) {
	total := moneyFromModel(m.Total)
	if total == nil {
		return 0, nil, "non-refundable", nil
	}
	var unit property.Unit
	if err := tx.WithContext(ctx).Select("id", "property_id").First(&unit, "id = ?", m.UnitID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, "non-refundable", nil
		}
		return 0, nil, "", err
	}
	scheduleName, err := types.ScheduleName(unit.PropertyID, m.UnitID)
	if err != nil {
		return 0, nil, "", err
	}
	var sched schedule.Schedule
	switch err := tx.WithContext(ctx).Preload("CancellationPolicy.RefundTiers").First(&sched, "name = ?", scheduleName).Error; {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return 0, nil, "non-refundable (no cancellation policy)", nil
	case err != nil:
		return 0, nil, "", err
	}
	if sched.CancellationPolicy == nil || len(sched.CancellationPolicy.RefundTiers) == 0 {
		return 0, nil, "non-refundable", nil
	}

	var lead time.Duration
	if m.Window != nil {
		lead = time.Until(m.Window.StartTime)
	}
	// The satisfied tier with the largest cutoff wins (cancelled at least cutoff
	// before the booking start).
	var bestPct int32
	bestCutoff := time.Duration(-1)
	for i := range sched.CancellationPolicy.RefundTiers {
		cutoff, perr := time.ParseDuration(sched.CancellationPolicy.RefundTiers[i].Cutoff)
		if perr != nil {
			continue
		}
		if lead >= cutoff && cutoff > bestCutoff {
			bestCutoff = cutoff
			bestPct = sched.CancellationPolicy.RefundTiers[i].RefundPercent
		}
	}
	return bestPct, moneyPct(total, bestPct), fmt.Sprintf("%d%% refund for the applicable tier", bestPct), nil
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
