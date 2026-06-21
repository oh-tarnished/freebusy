package hasura

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// buildUpdatePatch sets the masked scalar/enum/timestamp fields of an UpdateInput.
//
// KNOWN LIMITATION (generator-side): the generated input uses omitzero json tags,
// so a field set to its zero value is omitted from the mutation — a masked Update
// on the Hasura provider therefore CANNOT clear an optional field back to
// empty/null. The GORM adapter can (it writes NULL), so clear-to-empty via
// update_mask diverges between providers. Removing the divergence requires the
// generator to emit nullable (pointer) update inputs; until then callers should
// not rely on clearing optional fields through Hasura. The Money foreign keys and
// etag are set by the caller (see Update).
func buildUpdatePatch(pc *promocodepbv1.PromoCode, paths []string) resourceql.UpdateInput {
	p := resourceql.UpdateInput{UpdateTime: time.Now().UTC().Format(time.RFC3339Nano)}
	if inMask(paths, "code") {
		p.Code = pc.GetCode()
	}
	if inMask(paths, "display_name") {
		p.DisplayName = pc.GetDisplayName()
	}
	if inMask(paths, "description") {
		p.Description = pc.GetDescription()
	}
	if inMask(paths, "discount_type") {
		p.DiscountType = discountTypeToStr[pc.GetDiscountType()]
	}
	if inMask(paths, "percent_off") {
		p.PercentOff = pc.GetPercentOff()
	}
	if inMask(paths, "redeem_start_time") {
		p.RedeemStartTime = tsToRFC(pc.GetRedeemStartTime())
	}
	if inMask(paths, "redeem_end_time") {
		p.RedeemEndTime = tsToRFC(pc.GetRedeemEndTime())
	}
	if inMask(paths, "max_redemptions") {
		p.MaxRedemptions = graphql.Int64(pc.GetMaxRedemptions())
	}
	if inMask(paths, "per_customer_limit") {
		p.PerCustomerLimit = pc.GetPerCustomerLimit()
	}
	if inMask(paths, "redemption_count") {
		p.RedemptionCount = graphql.Int64(pc.GetRedemptionCount())
	}
	if inMask(paths, "disabled") {
		p.Disabled = pc.GetDisabled()
		p.State = stateForWrite(pc)
	}
	if inMask(paths, "state") {
		p.State = stateToStr[pc.GetState()]
	}
	return p
}
