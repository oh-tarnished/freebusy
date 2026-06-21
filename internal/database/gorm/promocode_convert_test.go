package gorm

import (
	"testing"
	"time"

	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// roundTrip mirrors what the repository does around the pure converters: map the
// proto onto a model, set the identity/foreign keys it owns, then read it back.
func roundTrip(in *promocodepbv1.PromoCode) *promocodepbv1.PromoCode {
	m := toPromoModel(in)
	m.ID = "ID123"
	m.Name = in.GetName()
	amount := moneyToModel(in.GetAmountOff())
	minSub := moneyToModel(in.GetMinSubtotal())
	return fromPromoModel(m, amount, minSub, in.GetApplicableResources(), in.GetApplicableOfferings())
}

func TestPromoConvertPercentageRoundTrip(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	in := &promocodepbv1.PromoCode{
		Name:                "promoCodes/ID123",
		Code:                "SUMMER25",
		DisplayName:         "Summer Sale",
		Description:         "25% off everything",
		DiscountType:        promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE,
		PercentOff:          25,
		MaxRedemptions:      100,
		PerCustomerLimit:    2,
		RedemptionCount:     5,
		MinSubtotal:         &money.Money{CurrencyCode: "USD", Units: 50},
		ApplicableResources: []string{"resources/room-1"},
		ApplicableOfferings: []string{"resources/room-1/offerings/night"},
		RedeemStartTime:     timestamppb.New(start),
	}

	out := roundTrip(in)

	if out.GetCode() != "SUMMER25" || out.GetDisplayName() != "Summer Sale" || out.GetDescription() != "25% off everything" {
		t.Fatalf("scalar fields not preserved: %+v", out)
	}
	if out.GetDiscountType() != promocodepbv1.DiscountType_DISCOUNT_TYPE_PERCENTAGE {
		t.Fatalf("discount type = %v", out.GetDiscountType())
	}
	if out.GetPercentOff() != 25 || out.GetMaxRedemptions() != 100 || out.GetPerCustomerLimit() != 2 || out.GetRedemptionCount() != 5 {
		t.Fatalf("numeric fields not preserved: %+v", out)
	}
	if out.GetMinSubtotal().GetUnits() != 50 || out.GetMinSubtotal().GetCurrencyCode() != "USD" {
		t.Fatalf("min subtotal not preserved: %+v", out.GetMinSubtotal())
	}
	if got := out.GetApplicableResources(); len(got) != 1 || got[0] != "resources/room-1" {
		t.Fatalf("applicable resources = %v", got)
	}
	if got := out.GetApplicableOfferings(); len(got) != 1 || got[0] != "resources/room-1/offerings/night" {
		t.Fatalf("applicable offerings = %v", got)
	}
	if !out.GetRedeemStartTime().AsTime().Equal(start) {
		t.Fatalf("redeem start = %v, want %v", out.GetRedeemStartTime().AsTime(), start)
	}
	if out.GetState() != promocodepbv1.PromoCodeState_PROMO_CODE_STATE_ACTIVE {
		t.Fatalf("state = %v, want ACTIVE", out.GetState())
	}
	if out.GetName() != "promoCodes/ID123" {
		t.Fatalf("name = %q", out.GetName())
	}
}

func TestPromoConvertFixedAmountDisabled(t *testing.T) {
	in := &promocodepbv1.PromoCode{
		Code:         "FLAT10",
		DiscountType: promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT,
		AmountOff:    &money.Money{CurrencyCode: "EUR", Units: 10, Nanos: 990000000},
		Disabled:     true,
	}

	out := roundTrip(in)

	if out.GetDiscountType() != promocodepbv1.DiscountType_DISCOUNT_TYPE_FIXED_AMOUNT {
		t.Fatalf("discount type = %v", out.GetDiscountType())
	}
	if out.GetAmountOff().GetUnits() != 10 || out.GetAmountOff().GetNanos() != 990000000 || out.GetAmountOff().GetCurrencyCode() != "EUR" {
		t.Fatalf("amount off not preserved: %+v", out.GetAmountOff())
	}
	if !out.GetDisabled() {
		t.Fatal("disabled flag not preserved")
	}
	if out.GetState() != promocodepbv1.PromoCodeState_PROMO_CODE_STATE_DISABLED {
		t.Fatalf("state = %v, want DISABLED", out.GetState())
	}
	// Unset optional money should stay nil.
	if out.GetMinSubtotal() != nil {
		t.Fatalf("expected nil min subtotal, got %+v", out.GetMinSubtotal())
	}
}
