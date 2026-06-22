package hasura

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
)

// buildUpdatePatch sets the masked fields of an UpdateInput. The generated input
// is now nullable (graphql.Nullable[T]): a field left at its zero value is Unset
// (omitted from the mutation, leaving the column untouched), graphql.Value sets a
// value, and graphql.Null clears the column. Optional fields use nullableStr so an
// empty masked value clears the column to NULL — matching the GORM adapter, which
// resolves the old clear-to-null divergence between providers.
func buildUpdatePatch(pc *promocodepbv1.PromoCode, paths []string) resourceql.UpdateInput {
	var p resourceql.UpdateInput
	p.UpdateTime = graphql.Value(time.Now().UTC().Format(time.RFC3339Nano))
	if inMask(paths, "code") {
		p.Code = graphql.Value(pc.GetCode())
	}
	if inMask(paths, "display_name") {
		p.DisplayName = nullableStr(pc.GetDisplayName())
	}
	if inMask(paths, "description") {
		p.Description = nullableStr(pc.GetDescription())
	}
	if inMask(paths, "discount_type") {
		p.DiscountType = graphql.Value(discountTypeToStr[pc.GetDiscountType()])
	}
	if inMask(paths, "percent_off") {
		p.PercentOff = graphql.Value(pc.GetPercentOff())
	}
	if inMask(paths, "redeem_start_time") {
		p.RedeemStartTime = nullableStr(tsToRFC(pc.GetRedeemStartTime()))
	}
	if inMask(paths, "redeem_end_time") {
		p.RedeemEndTime = nullableStr(tsToRFC(pc.GetRedeemEndTime()))
	}
	if inMask(paths, "max_redemptions") {
		p.MaxRedemptions = graphql.Value(graphql.Int64(pc.GetMaxRedemptions()))
	}
	if inMask(paths, "per_customer_limit") {
		p.PerCustomerLimit = graphql.Value(pc.GetPerCustomerLimit())
	}
	if inMask(paths, "redemption_count") {
		p.RedemptionCount = graphql.Value(graphql.Int64(pc.GetRedemptionCount()))
	}
	if inMask(paths, "disabled") {
		p.Disabled = graphql.Value(pc.GetDisabled())
		p.State = graphql.Value(stateForWrite(pc))
	}
	if inMask(paths, "state") {
		p.State = graphql.Value(stateToStr[pc.GetState()])
	}
	return p
}

// nullableStr maps an empty string to a NULL column (clear) and a non-empty
// string to a set value, so masked optional fields clear consistently with GORM.
func nullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}
