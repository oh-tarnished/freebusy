package gorm

import (
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
)

// inMask aliases the shared field-mask predicate so update semantics stay
// identical across the gorm and hasura adapters.
var inMask = repository.InMask

// applyMask copies the masked scalar/enum fields from pc onto the existing model.
// Identity, timestamps, and etag are managed by the repository, and the Money and
// join fields are handled separately in Update.
func applyMask(m *promocode.PromoCode, pc *promocodepbv1.PromoCode, paths []string) {
	if inMask(paths, "code") {
		m.Code = pc.GetCode()
	}
	if inMask(paths, "display_name") {
		m.DisplayName = strOrNil(pc.GetDisplayName())
	}
	if inMask(paths, "description") {
		m.Description = strOrNil(pc.GetDescription())
	}
	if inMask(paths, "discount_type") {
		if dt, ok := discountTypeToDB[pc.GetDiscountType()]; ok {
			m.DiscountType = dt
		}
	}
	if inMask(paths, "percent_off") {
		m.PercentOff = ptr(pc.GetPercentOff())
	}
	if inMask(paths, "redeem_start_time") {
		m.RedeemStartTime = tsToTime(pc.GetRedeemStartTime())
	}
	if inMask(paths, "redeem_end_time") {
		m.RedeemEndTime = tsToTime(pc.GetRedeemEndTime())
	}
	if inMask(paths, "max_redemptions") {
		m.MaxRedemptions = ptr(pc.GetMaxRedemptions())
	}
	if inMask(paths, "per_customer_limit") {
		m.PerCustomerLimit = ptr(pc.GetPerCustomerLimit())
	}
	if inMask(paths, "redemption_count") {
		m.RedemptionCount = ptr(pc.GetRedemptionCount())
	}
	if inMask(paths, "disabled") {
		m.Disabled = ptr(pc.GetDisabled())
		state := promocode.PromoCodeStateActive
		if pc.GetDisabled() {
			state = promocode.PromoCodeStateDisabled
		}
		m.State = &state
	}
	if inMask(paths, "state") {
		if s, ok := stateToDB[pc.GetState()]; ok {
			m.State = &s
		}
	}
}
