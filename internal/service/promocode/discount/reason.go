// The machine-readable reasons a code is not redeemable.
package discount

import ()

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
	ReasonNotApplicableProperty
	ReasonNotApplicableUnit
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
	case ReasonNotApplicableProperty:
		return "promo code is not applicable to this property"
	case ReasonNotApplicableUnit:
		return "promo code is not applicable to this unit"
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
