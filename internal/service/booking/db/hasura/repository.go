package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	resourceql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/resourceql"
	bookingschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/schemaql"
	moneysql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/moneysql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	guestsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/guestsql"
	identityschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/schemaql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/booking/party"
	"github.com/oh-tarnished/freebusy/internal/service/booking/pricing"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"github.com/the-protobuf-project/runtime-go/network/runtime"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultHoldTTL = 15 * time.Minute

// BookingRepository is the Hasura-backed booking repository.
type BookingRepository struct {
	svc *freebusyql.Service
}

// NewBookingRepository returns a Hasura-backed BookingRepository bound to svc.
func NewBookingRepository(svc *freebusyql.Service) *BookingRepository {
	return &BookingRepository{svc: svc}
}

// CreateBooking places a PENDING_HOLD on a unit for a window. It loads the unit
// (for capacity, price, booking mode, timezone), computes the full price
// breakdown, verifies capacity against overlapping active bookings, and persists
// the booking with its window / contact / money value-objects in one batch.
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
	unit, err := r.svc.Query.Property.Units.Get(ctx, unitID)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if unit == nil {
		return nil, types.ErrNotFound
	}

	requested := b.GetUnits()
	if requested < 1 {
		requested = 1
	}

	// Occupancy: the staying party must fit the unit's max occupancy across the
	// reserved units. Zero max_occupancy means unbounded.
	if !party.Fits(deref(unit.MaxOccupancy), requested, b.GetOccupancy(), b.GetGuests()) {
		return nil, types.ErrInvalidArgument
	}
	occupancy := occupancyInput(b.GetOccupancy())
	guestGraphs := buildGuestGraphs(b.GetGuests(), id)

	promoID := lastSegment(b.GetPromoCode())

	// Full price breakdown (base × nights, LOS + promo discounts, fees, taxes).
	in, err := r.pricingInputs(ctx, unit, promoID)
	if err != nil {
		return nil, err
	}
	in.Nights = nightsBetween(b.GetWindow(), unit.TimeZone)
	in.Units = int64(requested)

	var priceIn, discountIn, totalIn *moneysql.CreateInput
	var components []*sharedpbv1.PriceComponent
	if in.Price != nil {
		res := pricing.Compute(in, unitID)
		pi := moneyInput(res.Base)
		priceIn = &pi
		ti := moneyInput(res.Total)
		totalIn = &ti
		if !pricing.IsZero(res.Discount) {
			di := moneyInput(res.Discount)
			discountIn = &di
		}
		components = res.Components
	}

	// Capacity check against overlapping active bookings on the unit.
	reserved, err := r.reservedUnits(ctx, unitID, b.GetWindow(), "")
	if err != nil {
		return nil, err
	}
	capacity := int64(1)
	if unit.Capacity != nil && *unit.Capacity > 0 {
		capacity = int64(*unit.Capacity)
	}
	if reserved+int64(requested) > capacity {
		return nil, types.ErrConflict
	}

	now := time.Now().UTC()
	ttl := defaultHoldTTL
	if d := b.GetHoldTtl(); d != nil && d.AsDuration() > 0 {
		ttl = d.AsDuration()
	}
	window := windowInput(b.GetWindow())
	contact := contactInput(b.GetContact())

	bi := resourceql.CreateInput{
		Id:             id,
		Name:           name,
		Unit:           unitID,
		Customer:       lastSegment(b.GetCustomer()),
		Units:          requested,
		State:          "PENDING_HOLD",
		HoldExpireTime: tsToStr(timestamppb.New(now.Add(ttl))),
		PromoCode:      promoID,
		Notes:          b.GetNotes(),
		Attributes:     structToJSON(b.GetAttributes()),
		HoldTtl:        durationToStr(b.GetHoldTtl()),
		Etag:           ulid.GenerateString(),
		WindowId:       window.Id,
		CreateTime:     tsToStr(timestamppb.New(now)),
		UpdateTime:     tsToStr(timestamppb.New(now)),
	}
	if contact != nil {
		bi.ContactId = contact.Id
	}
	if priceIn != nil {
		bi.PriceId = priceIn.Id
	}
	if discountIn != nil {
		bi.DiscountId = discountIn.Id
	}
	if totalIn != nil {
		bi.TotalId = totalIn.Id
	}
	if occupancy != nil {
		bi.OccupancyId = occupancy.Id
	}

	tx := r.svc.Mutation.Tx()
	var winRes sharedschema.InsertSharedTimeWindowsResponse
	tx.Add(r.svc.Mutation.Shared.TimeWindows.CreateOp(window, &winRes))
	if contact != nil {
		var cRes sharedschema.InsertSharedContactsResponse
		tx.Add(r.svc.Mutation.Shared.Contacts.CreateOp(*contact, &cRes))
	}
	queueMoneyInserts(tx, r, priceIn, discountIn, totalIn)
	// Occupancy is belongs-to (before the booking); guests are has-many (after,
	// carrying the booking_id FK).
	if occupancy != nil {
		var oRes bookingschema.InsertBookingOccupanciesResponse
		tx.Add(r.svc.Mutation.Booking.Occupancies.CreateOp(*occupancy, &oRes))
	}
	var bRes bookingschema.InsertBookingResourceResponse
	tx.Add(r.svc.Mutation.Booking.Resource.CreateOp(bi, &bRes))
	queueGuestInserts(tx, r, guestGraphs)
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}

	out, err := r.GetBooking(ctx, name)
	if err != nil {
		return nil, err
	}
	// price_components are computed, not persisted; ride them along on the response.
	out.PriceComponents = components
	return out, nil
}

// queueMoneyInserts appends inserts for the non-nil Money value-objects.
func queueMoneyInserts(tx *runtime.Tx, r *BookingRepository, moneys ...*moneysql.CreateInput) {
	for _, mi := range moneys {
		if mi == nil {
			continue
		}
		var res commonschema.InsertCommonMoneysResponse
		tx.Add(r.svc.Mutation.Common.Moneys.CreateOp(*mi, &res))
	}
}

// GetBooking returns the booking addressed by its resource name.
func (r *BookingRepository) GetBooking(ctx context.Context, name string) (*bookingpbv1.Booking, error) {
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
	return r.hydrateBooking(ctx, res)
}

// ListBookings returns a page of bookings ordered by params.OrderBy.
func (r *BookingRepository) ListBookings(ctx context.Context, in repox.ListInput) ([]*bookingpbv1.Booking, string, error) {
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[bookingschema.BookingResource](booking.BookingFilterSpec, r.svc.Query.Booking.Resource).
		List(ctx, fin)
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*bookingpbv1.Booking, 0, len(rows))
	for i := range rows {
		out, err := r.hydrateBooking(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, out)
	}
	return items, next, nil
}

// hydrateBooking loads a booking row's value-objects and resolves its unit name.
func (r *BookingRepository) hydrateBooking(ctx context.Context, res *bookingschema.BookingResource) (*bookingpbv1.Booking, error) {
	parts := bookingParts{res: res}

	unitName, err := r.unitName(ctx, res.Unit)
	if err != nil {
		return nil, err
	}
	parts.unitName = unitName

	if res.ContactId != nil {
		c, err := r.svc.Query.Shared.Contacts.Get(ctx, *res.ContactId)
		if err != nil {
			return nil, mapHasuraErr(err)
		}
		parts.contact = c
	}
	if res.WindowId != "" {
		w, err := r.svc.Query.Shared.TimeWindows.Get(ctx, res.WindowId)
		if err != nil {
			return nil, mapHasuraErr(err)
		}
		parts.window = w
	}
	if res.PriceId != nil {
		if parts.price, err = r.money(ctx, *res.PriceId); err != nil {
			return nil, err
		}
	}
	if res.DiscountId != nil {
		if parts.discount, err = r.money(ctx, *res.DiscountId); err != nil {
			return nil, err
		}
	}
	if res.TotalId != nil {
		if parts.total, err = r.money(ctx, *res.TotalId); err != nil {
			return nil, err
		}
	}
	if res.RefundAmountId != nil {
		if parts.refund, err = r.money(ctx, *res.RefundAmountId); err != nil {
			return nil, err
		}
	}
	out := bookingFromParts(parts)

	if res.OccupancyId != nil {
		occ, err := r.svc.Query.Booking.Occupancies.Get(ctx, *res.OccupancyId)
		if err != nil {
			return nil, mapHasuraErr(err)
		}
		out.Occupancy = occupancyFromSchema(occ)
	}
	guests, err := r.loadGuests(ctx, res.Id)
	if err != nil {
		return nil, err
	}
	out.Guests = guests
	return out, nil
}

// loadGuests returns a booking's guest party, each with its sub-rows hydrated,
// ordered by id (ULIDs preserve insertion order).
func (r *BookingRepository) loadGuests(ctx context.Context, bookingID string) ([]*identitypbv1.Guest, error) {
	rows, err := r.svc.Query.Identity.Guests.List(ctx, guestsql.List().Where(guestsql.BookingId.Eq(bookingID)).OrderBy(guestsql.Id.Asc()))
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	out := make([]*identitypbv1.Guest, 0, len(rows))
	for i := range rows {
		g := &rows[i]
		var doc *identityschema.IdentityIdDocuments
		var foreigner *identityschema.IdentityForeignerDetails
		var prefs *identityschema.IdentityGuestPreferences
		var perm, loc *commonschema.CommonPostalAddress
		if g.IdDocumentId != nil {
			if doc, err = r.svc.Query.Identity.IdDocuments.Get(ctx, *g.IdDocumentId); err != nil {
				return nil, mapHasuraErr(err)
			}
		}
		if g.ForeignerId != nil {
			if foreigner, err = r.svc.Query.Identity.ForeignerDetails.Get(ctx, *g.ForeignerId); err != nil {
				return nil, mapHasuraErr(err)
			}
		}
		if g.PreferencesId != nil {
			if prefs, err = r.svc.Query.Identity.GuestPreferences.Get(ctx, *g.PreferencesId); err != nil {
				return nil, mapHasuraErr(err)
			}
		}
		if g.PermanentAddressId != nil {
			if perm, err = r.svc.Query.Common.PostalAddress.Get(ctx, *g.PermanentAddressId); err != nil {
				return nil, mapHasuraErr(err)
			}
		}
		if g.LocalAddressId != nil {
			if loc, err = r.svc.Query.Common.PostalAddress.Get(ctx, *g.LocalAddressId); err != nil {
				return nil, mapHasuraErr(err)
			}
		}
		out = append(out, guestFromSchema(g, doc, foreigner, prefs, perm, loc))
	}
	return out, nil
}

// reservedUnits sums the units of active bookings (held or confirmed) on unitID
// whose window overlaps target, excluding excludeID (empty to exclude none).
// Windows are compared as UTC instants, so the check is timezone-safe.
func (r *BookingRepository) reservedUnits(ctx context.Context, unitID string, target *sharedpbv1.TimeWindow, excludeID string) (int64, error) {
	preds := []graphql.Predicate{resourceql.Unit.Eq(unitID), resourceql.State.In("PENDING_HOLD", "CONFIRMED")}
	if excludeID != "" {
		preds = append(preds, resourceql.Id.Neq(excludeID))
	}
	rows, err := r.svc.Query.Booking.Resource.List(ctx, resourceql.List().Where(resourceql.And(preds...)))
	if err != nil {
		return 0, mapHasuraErr(err)
	}
	var sum int64
	for i := range rows {
		if rows[i].WindowId == "" {
			continue
		}
		w, err := r.svc.Query.Shared.TimeWindows.Get(ctx, rows[i].WindowId)
		if err != nil {
			return 0, mapHasuraErr(err)
		}
		if w == nil || !overlaps(w, target) {
			continue
		}
		u := int64(1)
		if rows[i].Units != nil && *rows[i].Units > 0 {
			u = int64(*rows[i].Units)
		}
		sum += u
	}
	return sum, nil
}

// unitName resolves a bare unit id to its full resource name (the booking row
// stores only the id, since its FK targets property.units.id).
func (r *BookingRepository) unitName(ctx context.Context, unitID string) (string, error) {
	if unitID == "" {
		return "", nil
	}
	u, err := r.svc.Query.Property.Units.Get(ctx, unitID)
	if err != nil {
		return "", mapHasuraErr(err)
	}
	if u == nil {
		return "", nil
	}
	return u.Name, nil
}

// overlaps reports whether stored window w overlaps target [start,end) as UTC
// instants (half-open: touching endpoints do not overlap).
func overlaps(w *sharedschema.SharedTimeWindows, target *sharedpbv1.TimeWindow) bool {
	ws, we := strToTS(w.StartTime), strToTS(w.EndTime)
	if ws == nil || we == nil || target == nil {
		return false
	}
	return ws.AsTime().Before(target.GetEndTime().AsTime()) && we.AsTime().After(target.GetStartTime().AsTime())
}
