// Package discount evaluates a promo code against a prospective booking. It
// contains the pure, side-effect-free business rules behind
// PromoCodeService.ValidatePromoCode, so they can be unit-tested without a
// database, a server, or the observability stack.
package discount

import (
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/genproto/googleapis/type/money"
)

// nanosPerUnit is the number of nano-units in one currency unit (google.type.Money
// splits an amount into whole units plus nanos).
const nanosPerUnit = 1_000_000_000

// Result is the outcome of evaluating a promo code. When Valid is false, Reason
// classifies why (Reason.Message() gives the text) and Discount/FinalTotal are nil.
type Result struct {
	Valid      bool
	Reason     Reason
	Discount   *money.Money
	FinalTotal *money.Money
}

// Reason classifies why a promo code is not redeemable for a booking. It is
// transport-neutral; the service layer maps it to a gRPC status code.
// ReasonNone means the code is redeemable.
type Reason int

const (
	ReasonNone Reason = iota
	ReasonDisabled
	ReasonExpired
	ReasonNotYetRedeemable
	ReasonLimitReached
	ReasonCurrencyMismatch
	ReasonBelowMinimum
	ReasonNotApplicableResource
	ReasonNotApplicableOffering
)

// Message returns the human-readable explanation for r ("" when ReasonNone).
func (r Reason) Message() string {
	switch r {
	case ReasonDisabled:
		return "promo code is disabled"
	case ReasonExpired:
		return "promo code has expired"
	case ReasonNotYetRedeemable:
		return "promo code is not yet redeemable"
	case ReasonLimitReached:
		return "promo code redemption limit reached"
	case ReasonCurrencyMismatch:
		return "promo code currency does not match the booking subtotal"
	case ReasonBelowMinimum:
		return "subtotal is below the required minimum"
	case ReasonNotApplicableResource:
		return "promo code is not applicable to this resource"
	case ReasonNotApplicableOffering:
		return "promo code is not applicable to this offering"
	default:
		return ""
	}
}

// String implements fmt.Stringer for readable test/log output.
func (r Reason) String() string {
	if m := r.Message(); m != "" {
		return m
	}
	return "redeemable"
}

// Evaluate applies pc's redemption window, caps, minimum subtotal, and resource /
// offering scope to a prospective booking, then computes the discount and final
// total. now is injected so the caller (and tests) control the clock.
func Evaluate(pc *promocodepbv1.PromoCode, subtotal *money.Money, resource, offering string, now time.Time) Result {
	if reason := redeemable(pc, subtotal, resource, offering, now); reason != ReasonNone {
		return Result{Valid: false, Reason: reason}
	}
	discount := computeDiscount(pc, subtotal)
	return Result{
		Valid:      true,
		Discount:   discount,
		FinalTotal: subtract(subtotal, discount),
	}
}

// redeemable reports why pc may not be redeemed for this request, returning
// ReasonNone when it is redeemable.
func redeemable(pc *promocodepbv1.PromoCode, subtotal *money.Money, resource, offering string, now time.Time) Reason {
	if pc.GetDisabled() {
		return ReasonDisabled
	}
	if pc.GetState() == promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED {
		return ReasonExpired
	}
	if start := pc.GetWindow().GetStartTime(); start != nil && now.Before(start.AsTime()) {
		return ReasonNotYetRedeemable
	}
	if end := pc.GetWindow().GetEndTime(); end != nil && now.After(end.AsTime()) {
		return ReasonExpired
	}
	if max := pc.GetLimits().GetMaxRedemptions(); max != nil && pc.GetRedemptionCount() >= max.GetValue() {
		return ReasonLimitReached
	}
	// Money amounts are only comparable within the same currency; a mismatch
	// between the booking subtotal and the code's thresholds/amounts is treated as
	// not-applicable rather than silently computed across currencies.
	if min := pc.GetScope().GetMinSubtotal(); min != nil {
		if !sameCurrency(min, subtotal) {
			return ReasonCurrencyMismatch
		}
		if toNanos(subtotal) < toNanos(min) {
			return ReasonBelowMinimum
		}
	}
	if amt := pc.GetDiscount().GetAmountOff(); amt != nil && !sameCurrency(amt, subtotal) {
		return ReasonCurrencyMismatch
	}
	if res := pc.GetScope().GetApplicableResources(); len(res) > 0 && !contains(res, resource) {
		return ReasonNotApplicableResource
	}
	if off := pc.GetScope().GetApplicableOfferings(); len(off) > 0 && !contains(off, offering) {
		return ReasonNotApplicableOffering
	}
	return ReasonNone
}

// computeDiscount returns the discount amount for a redeemable code, clamped to
// the subtotal (a discount never exceeds the amount being discounted).
func computeDiscount(pc *promocodepbv1.PromoCode, subtotal *money.Money) *money.Money {
	sub := toNanos(subtotal)
	var d int64
	// The discount is a oneof: a fixed amount_off (Money) or a percent_off. A
	// non-nil amount_off means the fixed arm is set; otherwise treat it as a
	// percentage.
	if amt := pc.GetDiscount().GetAmountOff(); amt != nil {
		d = toNanos(amt)
	} else {
		// Split the multiply to avoid int64 overflow on very large subtotals:
		// percent_off <= 100, so (sub%100)*percent can't overflow and the result
		// never exceeds sub.
		pct := int64(pc.GetDiscount().GetPercentOff())
		d = sub/100*pct + sub%100*pct/100
	}
	if d > sub {
		d = sub
	}
	if d < 0 {
		d = 0
	}
	return fromNanos(d, currencyOf(subtotal))
}

// subtract returns subtotal minus discount, floored at zero, in the subtotal's
// currency.
func subtract(subtotal, discount *money.Money) *money.Money {
	total := toNanos(subtotal) - toNanos(discount)
	if total < 0 {
		total = 0
	}
	return fromNanos(total, currencyOf(subtotal))
}

// toNanos collapses a Money into a single nano-unit integer for arithmetic.
func toNanos(m *money.Money) int64 {
	if m == nil {
		return 0
	}
	return m.GetUnits()*nanosPerUnit + int64(m.GetNanos())
}

// fromNanos expands a nano-unit integer back into a Money of the given currency.
func fromNanos(total int64, currency string) *money.Money {
	return &money.Money{
		CurrencyCode: currency,
		Units:        total / nanosPerUnit,
		Nanos:        int32(total % nanosPerUnit),
	}
}

func currencyOf(m *money.Money) string {
	if m == nil {
		return ""
	}
	return m.GetCurrencyCode()
}

func sameCurrency(a, b *money.Money) bool {
	return a.GetCurrencyCode() == b.GetCurrencyCode()
}

// EffectiveState derives a promo code's lifecycle state from its flags and
// redemption window at time now. The stored state column can be stale (a code
// only becomes EXPIRED with the passage of time or as redemptions accrue), so
// reads should report this derived value rather than the persisted one.
func EffectiveState(pc *promocodepbv1.PromoCode, now time.Time) promocodepbv1.PromoCodeState {
	if pc.GetDisabled() {
		return promocodepbv1.PromoCodeState_PROMO_CODE_STATE_DISABLED
	}
	if end := pc.GetWindow().GetEndTime(); end != nil && now.After(end.AsTime()) {
		return promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED
	}
	if max := pc.GetLimits().GetMaxRedemptions(); max != nil && pc.GetRedemptionCount() >= max.GetValue() {
		return promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED
	}
	return promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}
