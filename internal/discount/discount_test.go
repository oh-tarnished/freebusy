package discount

import (
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var now = time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)

func usd(units int64, nanos int32) *money.Money {
	return &money.Money{CurrencyCode: "USD", Units: units, Nanos: nanos}
}

func activePercent(pct int32) *promocodepbv1.PromoCode {
	return &promocodepbv1.PromoCode{
		Code:         "SAVE",
		DiscountType: promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE,
		PercentOff:   pct,
		State:        promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE,
	}
}

func TestEvaluateDiscountMath(t *testing.T) {
	t.Run("percentage", func(t *testing.T) {
		got := Evaluate(activePercent(10), usd(100, 0), "", "", now)
		assertValid(t, got, 10, 0, 90, 0)
	})

	t.Run("fixed amount", func(t *testing.T) {
		pc := activePercent(0)
		pc.DiscountType = promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT
		pc.AmountOff = usd(15, 500000000) // $15.50
		got := Evaluate(pc, usd(100, 0), "", "", now)
		assertValid(t, got, 15, 500000000, 84, 500000000)
	})

	t.Run("fixed amount capped at subtotal", func(t *testing.T) {
		pc := activePercent(0)
		pc.DiscountType = promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT
		pc.AmountOff = usd(200, 0)
		got := Evaluate(pc, usd(100, 0), "", "", now)
		assertValid(t, got, 100, 0, 0, 0)
	})

	t.Run("very large subtotal does not overflow", func(t *testing.T) {
		// sub = 9e18 nanos; a naive sub*percent would overflow int64. The split
		// math must still yield exactly 50%.
		got := Evaluate(activePercent(50), usd(9_000_000_000, 0), "", "", now)
		assertValid(t, got, 4_500_000_000, 0, 4_500_000_000, 0)
	})
}

func TestEvaluateRedeemability(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*promocodepbv1.PromoCode)
		reason string
	}{
		{"disabled", func(p *promocodepbv1.PromoCode) { p.Disabled = true }, "promo code is disabled"},
		{"expired state", func(p *promocodepbv1.PromoCode) {
			p.State = promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED
		}, "promo code has expired"},
		{"before window", func(p *promocodepbv1.PromoCode) {
			p.RedeemStartTime = timestamppb.New(now.Add(time.Hour))
		}, "promo code is not yet redeemable"},
		{"after window", func(p *promocodepbv1.PromoCode) {
			p.RedeemEndTime = timestamppb.New(now.Add(-time.Hour))
		}, "promo code has expired"},
		{"caps reached", func(p *promocodepbv1.PromoCode) {
			p.MaxRedemptions = 5
			p.RedemptionCount = 5
		}, "promo code redemption limit reached"},
		{"below minimum", func(p *promocodepbv1.PromoCode) {
			p.MinSubtotal = usd(50, 0)
		}, "subtotal is below the required minimum"},
		{"resource scope", func(p *promocodepbv1.PromoCode) {
			p.ApplicableResources = []string{"resources/allowed"}
		}, "promo code is not applicable to this resource"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pc := activePercent(10)
			tc.mutate(pc)
			got := Evaluate(pc, usd(40, 0), "resources/other", "", now)
			if got.Valid {
				t.Fatalf("expected invalid, got valid")
			}
			if got.Reason != tc.reason {
				t.Fatalf("reason = %q, want %q", got.Reason, tc.reason)
			}
			if got.Discount != nil || got.FinalTotal != nil {
				t.Fatalf("expected nil discount/total on invalid result")
			}
		})
	}
}

func TestEvaluateCurrencyMismatch(t *testing.T) {
	eur := &money.Money{CurrencyCode: "EUR", Units: 10}
	const reason = "promo code currency does not match the booking subtotal"

	t.Run("min_subtotal currency", func(t *testing.T) {
		pc := activePercent(10)
		pc.MinSubtotal = eur
		if got := Evaluate(pc, usd(100, 0), "", "", now); got.Valid || got.Reason != reason {
			t.Fatalf("got (valid=%v, reason=%q), want invalid currency mismatch", got.Valid, got.Reason)
		}
	})

	t.Run("amount_off currency", func(t *testing.T) {
		pc := activePercent(0)
		pc.DiscountType = promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT
		pc.AmountOff = eur
		if got := Evaluate(pc, usd(100, 0), "", "", now); got.Valid || got.Reason != reason {
			t.Fatalf("got (valid=%v, reason=%q), want invalid currency mismatch", got.Valid, got.Reason)
		}
	})
}

func TestEffectiveState(t *testing.T) {
	active := promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE
	disabled := promocodepbv1.PromoCodeState_PROMO_CODE_STATE_DISABLED
	expired := promocodepbv1.PromoCodeState_PROMO_CODE_STATE_EXPIRED

	cases := []struct {
		name   string
		pc     *promocodepbv1.PromoCode
		expect promocodepbv1.PromoCodeState
	}{
		{"active", &promocodepbv1.PromoCode{}, active},
		{"disabled", &promocodepbv1.PromoCode{Disabled: true}, disabled},
		{"past window", &promocodepbv1.PromoCode{RedeemEndTime: timestamppb.New(now.Add(-time.Hour))}, expired},
		{"redemptions exhausted", &promocodepbv1.PromoCode{MaxRedemptions: 5, RedemptionCount: 5}, expired},
		{"within window", &promocodepbv1.PromoCode{RedeemEndTime: timestamppb.New(now.Add(time.Hour))}, active},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EffectiveState(tc.pc, now); got != tc.expect {
				t.Fatalf("EffectiveState = %v, want %v", got, tc.expect)
			}
		})
	}
}

func TestEvaluateScopeMatchPasses(t *testing.T) {
	pc := activePercent(20)
	pc.ApplicableOfferings = []string{"resources/r/offerings/o"}
	got := Evaluate(pc, usd(50, 0), "resources/r", "resources/r/offerings/o", now)
	assertValid(t, got, 10, 0, 40, 0)
}

func assertValid(t *testing.T, r Result, discUnits int64, discNanos int32, finalUnits int64, finalNanos int32) {
	t.Helper()
	if !r.Valid {
		t.Fatalf("expected valid, got invalid: %q", r.Reason)
	}
	if r.Discount.GetUnits() != discUnits || r.Discount.GetNanos() != discNanos {
		t.Fatalf("discount = %d.%09d, want %d.%09d", r.Discount.GetUnits(), r.Discount.GetNanos(), discUnits, discNanos)
	}
	if r.FinalTotal.GetUnits() != finalUnits || r.FinalTotal.GetNanos() != finalNanos {
		t.Fatalf("final = %d.%09d, want %d.%09d", r.FinalTotal.GetUnits(), r.FinalTotal.GetNanos(), finalUnits, finalNanos)
	}
	if r.Discount.GetCurrencyCode() != "USD" || r.FinalTotal.GetCurrencyCode() != "USD" {
		t.Fatalf("currency not preserved: disc=%q final=%q", r.Discount.GetCurrencyCode(), r.FinalTotal.GetCurrencyCode())
	}
}
