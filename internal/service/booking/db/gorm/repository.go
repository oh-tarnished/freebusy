// Package gorm provides the GORM-backed implementation of the booking
// persistence contract (internal/service/booking/db.BookingRepository). It owns
// the hold lifecycle, the capacity/overlap check that prevents overbooking, and a
// base-price computation from the unit's price (evaluated in the unit timezone
// for nightly stays).
package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/booking/party"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// overlapSQL sums the reserved units of active bookings (held or confirmed) whose
// window overlaps [start,end) on a unit, for the capacity check. Windows are
// compared as UTC instants, so the check is timezone-safe.
const overlapSQL = `
SELECT COALESCE(SUM(COALESCE(b.units, 1)), 0)
FROM "booking"."resource" b
JOIN "shared"."time_windows" w ON w.id = b.window_id
WHERE b.unit = ? AND b.state IN ('PENDING_HOLD','CONFIRMED')
  AND w.start_time < ? AND w.end_time > ?`

// BookingRepository is the GORM-backed booking repository.
type BookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository returns a GORM-backed BookingRepository bound to db.
func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func preloadBooking(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Contact").
		Preload("Window").
		Preload("Price").
		Preload("Discount").
		Preload("Total").
		Preload("RefundAmount").
		Preload("Occupancy")
}

// mapGormErr translates GORM sentinel errors into the provider-neutral errors in
// internal/types.
func mapGormErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return types.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return types.ErrAlreadyExists
	default:
		return err
	}
}

// CreateBooking places a PENDING_HOLD on a unit for a window. It loads the unit
// (for capacity, price, booking mode, timezone), verifies capacity against
// overlapping active bookings, computes a base price, and persists the booking
// with its window / contact / price value-objects in one transaction.
func (r *BookingRepository) CreateBooking(ctx context.Context, b *bookingpbv1.Booking) (*bookingpbv1.Booking, error) {
	id, name, err := types.ResolveBookingName(b.GetName())
	if err != nil {
		return nil, err
	}
	_, unitID, err := types.ParseUnitParent(b.GetUnit())
	if err != nil {
		return nil, err
	}
	if b.GetWindow() == nil {
		return nil, types.ErrInvalidArgument
	}

	var unit property.Unit
	if err := r.db.WithContext(ctx).
		Preload("Price").
		Preload("Fees").Preload("Fees.Amount").
		Preload("Taxes").
		Preload("LosDiscounts").Preload("LosDiscounts.AmountOff").
		First(&unit, "id = ?", unitID).Error; err != nil {
		return nil, mapGormErr(err)
	}

	// Load the promo code (with its discount and scope) when one is applied, so the
	// pricing engine can evaluate its scope and discount.
	var promo *promocode.PromoCode
	if pid := lastSegment(b.GetPromoCode()); pid != "" {
		var p promocode.PromoCode
		if err := r.db.WithContext(ctx).
			Preload("Discount").Preload("Discount.AmountOff").
			Preload("Scope").Preload("Scope.MinSubtotal").Preload("Scope.ScopeApplicableUnits").
			First(&p, "id = ?", pid).Error; err != nil {
			return nil, mapGormErr(err)
		}
		promo = &p
	}

	requested := b.GetUnits()
	if requested < 1 {
		requested = 1
	}

	// Occupancy: the staying party must fit the unit's max occupancy across the
	// reserved units (guests × max_occupancy). Zero max_occupancy means unbounded.
	if !party.Fits(deref(unit.MaxOccupancy), requested, b.GetOccupancy(), b.GetGuests()) {
		return nil, types.ErrInvalidArgument
	}
	occupancy := occupancyToModel(b.GetOccupancy())
	guestGraphs := buildGuestGraphs(b.GetGuests(), id)

	window := timeWindowToModel(b.GetWindow())
	contact := contactToModel(b.GetContact())

	// Full price breakdown: base × nights, then LOS + promo discounts, fees, taxes.
	// Nights are counted in the unit's timezone; the itemized components ride along
	// on the create response (they are not persisted).
	var priceModel, discountModel, totalModel *common.Money
	var components []*sharedpbv1.PriceComponent
	if unit.Price != nil {
		nights := nightsBetween(b.GetWindow(), unit.TimeZone)
		p := computePricing(&unit, nights, int64(requested), promo)
		priceModel = moneyToModel(p.base)
		totalModel = moneyToModel(p.total)
		if !isZeroMoney(p.discount) {
			discountModel = moneyToModel(p.discount)
		}
		components = p.components
	}

	state := booking.BookingStatePendingHold
	ttl := defaultHoldTTL
	if d := b.GetHoldTtl(); d != nil && d.AsDuration() > 0 {
		ttl = d.AsDuration()
	}
	holdExpire := time.Now().UTC().Add(ttl)

	m := &booking.Booking{
		ID:             id,
		Name:           name,
		UnitID:         unitID,
		CustomerID:     strOrNil(lastSegment(b.GetCustomer())),
		Units:          ptr(requested),
		State:          &state,
		HoldExpireTime: &holdExpire,
		PromoCodeID:    strOrNil(lastSegment(b.GetPromoCode())),
		Notes:          strOrNil(b.GetNotes()),
		Attributes:     structToJSON(b.GetAttributes()),
		HoldTtl:        durationToStr(b.GetHoldTtl()),
		Etag:           ptr(ulid.GenerateString()),
		WindowID:       window.ID,
	}
	if contact != nil {
		m.ContactID = &contact.ID
	}
	if priceModel != nil {
		m.PriceID = &priceModel.ID
	}
	if discountModel != nil {
		m.DiscountID = &discountModel.ID
	}
	if totalModel != nil {
		m.TotalID = &totalModel.ID
	}
	if occupancy != nil {
		m.OccupancyID = &occupancy.ID
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		var reserved int64
		if e := tx.WithContext(ctx).Raw(overlapSQL, unitID, window.EndTime, window.StartTime).Scan(&reserved).Error; e != nil {
			return e
		}
		capacity := int64(1)
		if unit.Capacity != nil && *unit.Capacity > 0 {
			capacity = int64(*unit.Capacity)
		}
		if reserved+int64(requested) > capacity {
			return types.ErrConflict
		}
		if e := shared.NewTimeWindowStore(tx).Create(ctx, window); e != nil {
			return e
		}
		if contact != nil {
			if e := shared.NewContactStore(tx).Create(ctx, contact); e != nil {
				return e
			}
		}
		moneys := common.NewMoneyStore(tx)
		if priceModel != nil {
			if e := moneys.Create(ctx, priceModel); e != nil {
				return e
			}
		}
		if discountModel != nil {
			if e := moneys.Create(ctx, discountModel); e != nil {
				return e
			}
		}
		if totalModel != nil {
			if e := moneys.Create(ctx, totalModel); e != nil {
				return e
			}
		}
		// Occupancy is belongs-to (created before the booking); guests are has-many
		// (created after, carrying the booking_id FK).
		if occupancy != nil {
			if e := booking.NewOccupancyStore(tx).Create(ctx, occupancy); e != nil {
				return e
			}
		}
		if e := booking.NewBookingStore(tx).Create(ctx, m); e != nil {
			return e
		}
		return persistGuests(ctx, tx, guestGraphs)
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	out, err := r.GetBooking(ctx, name)
	if err != nil {
		return nil, err
	}
	// price_components are computed, not persisted; ride them along on the response.
	out.PriceComponents = components
	return out, nil
}

// GetBooking returns the booking addressed by its resource name.
func (r *BookingRepository) GetBooking(ctx context.Context, name string) (*bookingpbv1.Booking, error) {
	id, err := types.BookingID(name)
	if err != nil {
		return nil, err
	}
	var m booking.Booking
	if err := preloadBooking(r.db.WithContext(ctx)).First(&m, "id = ?", id).Error; err != nil {
		return nil, mapGormErr(err)
	}
	unitName, err := r.unitName(ctx, m.UnitID)
	if err != nil {
		return nil, err
	}
	out := bookingFromModel(&m, unitName)
	guests, err := r.loadGuests(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	out.Guests = guests
	return out, nil
}

// ListBookings returns a page of bookings ordered by params.OrderBy.
func (r *BookingRepository) ListBookings(ctx context.Context, in repox.ListInput) ([]*bookingpbv1.Booking, string, error) {
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	models, next, err := filterx.Gorm[booking.Booking](booking.BookingFilterSpec).
		List(ctx, preloadBooking(r.db), fin)
	if err != nil {
		return nil, "", mapGormErr(types.MapFilterxErr(err))
	}
	unitNames, err := r.unitNames(ctx, models)
	if err != nil {
		return nil, "", err
	}
	items := make([]*bookingpbv1.Booking, 0, len(models))
	for i := range models {
		out := bookingFromModel(&models[i], unitNames[models[i].UnitID])
		guests, err := r.loadGuests(ctx, models[i].ID)
		if err != nil {
			return nil, "", err
		}
		out.Guests = guests
		items = append(items, out)
	}
	return items, next, nil
}

// unitName resolves a bare unit id to its full resource name (the booking row
// stores only the id, since its FK targets property.units.id).
func (r *BookingRepository) unitName(ctx context.Context, unitID string) (string, error) {
	var u property.Unit
	if err := r.db.WithContext(ctx).Select("id", "name").First(&u, "id = ?", unitID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", mapGormErr(err)
	}
	return u.Name, nil
}

// unitNames batches the id→name resolution for a page of bookings.
func (r *BookingRepository) unitNames(ctx context.Context, bookings []booking.Booking) (map[string]string, error) {
	ids := make([]string, 0, len(bookings))
	seen := map[string]bool{}
	for i := range bookings {
		if id := bookings[i].UnitID; id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var units []property.Unit
	if err := r.db.WithContext(ctx).Select("id", "name").Where("id IN ?", ids).Find(&units).Error; err != nil {
		return nil, mapGormErr(err)
	}
	for i := range units {
		out[units[i].ID] = units[i].Name
	}
	return out, nil
}
