package gorm

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/discount"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure (side-effect-free) conversions between the protobuf
// PromoCode and the GORM storage models. The protobuf API embeds Money value
// objects and string lists; the schema normalizes those into a separate Money
// table (booking schema) and into applicable-resources / applicable-offerings
// join rows. The join columns store the full API resource name verbatim so the
// list values round-trip exactly (the schema has no column for an offering's
// parent resource).

// Enum maps translate between the protobuf int32 enums and the string values
// persisted by GORM (which match the database CHECK constraints).
var (
	discountTypeToDB = map[promocodepbv1.DiscountType]promocode.DiscountType{
		promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE:   promocode.DiscountTypePercentage,
		promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT: promocode.DiscountTypeFixedAmount,
	}
	discountTypeFromDB = map[promocode.DiscountType]promocodepbv1.DiscountType{
		promocode.DiscountTypePercentage:  promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE,
		promocode.DiscountTypeFixedAmount: promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT,
	}
	stateToDB = map[promocodepbv1.PromoCodeState]promocode.PromoCodeState{
		promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE:   promocode.PromoCodeStateActive,
		promocodepbv1.PromoCodeState_PROMO_CODE_STATE_DISABLED: promocode.PromoCodeStateDisabled,
		promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED:  promocode.PromoCodeStateExpired,
	}
)

func ptr[T any](v T) *T { return &v }

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// strOrNil maps an empty proto string (which cannot represent NULL) to a nil
// column pointer, so unset optional strings stay NULL in the database.
func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func tsToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func timeToTS(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// moneyToModel builds a new Money row (with a fresh ULID id) from a proto Money,
// or returns nil when m is nil.
func moneyToModel(m *money.Money) *booking.Money {
	if m == nil {
		return nil
	}
	return &booking.Money{
		ID:           ulid.GenerateString(),
		CurrencyCode: strOrNil(m.GetCurrencyCode()),
		Units:        ptr(m.GetUnits()),
		Nanos:        ptr(m.GetNanos()),
	}
}

func moneyFromModel(m *booking.Money) *money.Money {
	if m == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: deref(m.CurrencyCode),
		Units:        deref(m.Units),
		Nanos:        deref(m.Nanos),
	}
}

// toPromoModel maps the scalar, enum, and timestamp fields of a proto PromoCode
// onto a fresh GORM model. Identity (ID/Name), Money foreign keys, etag, and the
// join rows are set by the repository, which owns id generation and the
// transaction.
func toPromoModel(pc *promocodepbv1.PromoCode) *promocode.PromoCode {
	dt, ok := discountTypeToDB[pc.GetDiscountType()]
	if !ok {
		dt = promocode.DiscountTypePercentage // not-null column default
	}
	state := promocode.PromoCodeStateActive
	if pc.GetDisabled() {
		state = promocode.PromoCodeStateDisabled
	}
	return &promocode.PromoCode{
		Code:             pc.GetCode(),
		DisplayName:      strOrNil(pc.GetDisplayName()),
		Description:      strOrNil(pc.GetDescription()),
		DiscountType:     dt,
		PercentOff:       ptr(pc.GetPercentOff()),
		RedeemStartTime:  tsToTime(pc.GetRedeemStartTime()),
		RedeemEndTime:    tsToTime(pc.GetRedeemEndTime()),
		MaxRedemptions:   ptr(pc.GetMaxRedemptions()),
		PerCustomerLimit: ptr(pc.GetPerCustomerLimit()),
		RedemptionCount:  ptr(pc.GetRedemptionCount()),
		State:            &state,
		Disabled:         ptr(pc.GetDisabled()),
	}
}

// fromPromoModel assembles the protobuf PromoCode from the stored model, its
// resolved Money value-objects, and the applicable-resources / -offerings names.
func fromPromoModel(m *promocode.PromoCode, amount, minSub *booking.Money, resNames, offNames []string) *promocodepbv1.PromoCode {
	pc := &promocodepbv1.PromoCode{
		Name:                m.Name,
		Code:                m.Code,
		DisplayName:         deref(m.DisplayName),
		Description:         deref(m.Description),
		DiscountType:        discountTypeFromDB[m.DiscountType],
		PercentOff:          deref(m.PercentOff),
		AmountOff:           moneyFromModel(amount),
		RedeemStartTime:     timeToTS(m.RedeemStartTime),
		RedeemEndTime:       timeToTS(m.RedeemEndTime),
		MaxRedemptions:      deref(m.MaxRedemptions),
		PerCustomerLimit:    deref(m.PerCustomerLimit),
		MinSubtotal:         moneyFromModel(minSub),
		ApplicableResources: resNames,
		ApplicableOfferings: offNames,
		RedemptionCount:     deref(m.RedemptionCount),
		Disabled:            deref(m.Disabled),
		CreateTime:          timeToTS(&m.CreateTime),
		UpdateTime:          timeToTS(&m.UpdateTime),
		Etag:                deref(m.Etag),
	}
	// Derive the lifecycle state from the window/flags rather than trusting the
	// possibly-stale stored value (a code becomes EXPIRED purely with time).
	pc.State = discount.EffectiveState(pc, time.Now().UTC())
	return pc
}

// buildApplicableResources / buildApplicableOfferings turn the proto name lists
// into join rows (each with a fresh ULID id) pointing back at promoID. The full
// API name is stored verbatim in the foreign-key column so reads round-trip.
func buildApplicableResources(promoID string, names []string) []*promocode.PromoCodeApplicableResources {
	rows := make([]*promocode.PromoCodeApplicableResources, 0, len(names))
	for _, name := range names {
		rows = append(rows, &promocode.PromoCodeApplicableResources{
			ID:          ulid.GenerateString(),
			PromoCodeID: promoID,
			ResourceID:  name,
		})
	}
	return rows
}

func buildApplicableOfferings(promoID string, names []string) []*promocode.PromoCodeApplicableOfferings {
	rows := make([]*promocode.PromoCodeApplicableOfferings, 0, len(names))
	for _, name := range names {
		rows = append(rows, &promocode.PromoCodeApplicableOfferings{
			ID:          ulid.GenerateString(),
			PromoCodeID: promoID,
			OfferingID:  name,
		})
	}
	return rows
}
