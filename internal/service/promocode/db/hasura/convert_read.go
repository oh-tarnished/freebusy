// Read-side assembly: rows back to the PromoCode proto.
package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"time"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	pcschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/service/promocode/discount"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// parts holds a stored resource row and the child rows fetched to hydrate it.
type parts struct {
	res        *pcschema.PromocodeResource
	discount   *pcschema.PromocodeDiscounts
	amountOff  *commonschema.CommonMoneys
	window     *pcschema.PromocodeRedemptionWindows
	limits     *pcschema.PromocodeUsageLimits
	scope      *pcschema.PromocodeScopes
	minSub     *commonschema.CommonMoneys
	properties []pcschema.PromocodeScopeApplicableProperties
	units      []pcschema.PromocodeScopeApplicableUnits
}

// fromParts assembles the protobuf PromoCode from a stored resource row and its
// fetched child rows.
func fromParts(p parts) *promocodepbv1.PromoCode {
	res := p.res
	pc := &promocodepbv1.PromoCode{
		Name:            res.Name,
		Code:            res.Code,
		DisplayName:     repox.Deref(res.DisplayName),
		Description:     repox.Deref(res.Description),
		Discount:        discountFromModel(p.discount, p.amountOff),
		Window:          windowFromModel(p.window),
		Limits:          limitsFromModel(p.limits),
		Scope:           scopeFromParts(p.scope, p.minSub, p.properties, p.units),
		RedemptionCount: int64(repox.Deref(res.RedemptionCount)),
		Disabled:        repox.Deref(res.Disabled),
		CreateTime:      strToTS(&res.CreateTime),
		UpdateTime:      strToTS(&res.UpdateTime),
		Etag:            repox.Deref(res.Etag),
	}
	// Derive the lifecycle state from the window/flags rather than trusting the
	// possibly-stale stored value (a code becomes EXPIRED purely with time).
	pc.State = discount.EffectiveState(pc, time.Now().UTC())
	return pc
}

func discountFromModel(d *pcschema.PromocodeDiscounts, amountOff *commonschema.CommonMoneys) *promocodepbv1.Discount {
	if d == nil {
		return nil
	}
	out := &promocodepbv1.Discount{}
	if d.AmountCase != nil && *d.AmountCase == amountCaseAmountOff {
		out.Amount = &promocodepbv1.Discount_AmountOff{AmountOff: moneyFromModel(amountOff)}
	} else {
		out.Amount = &promocodepbv1.Discount_PercentOff{PercentOff: repox.Deref(d.PercentOff)}
	}
	return out
}

func windowFromModel(w *pcschema.PromocodeRedemptionWindows) *promocodepbv1.RedemptionWindow {
	if w == nil {
		return nil
	}
	return &promocodepbv1.RedemptionWindow{
		StartTime: strToTS(w.StartTime),
		EndTime:   strToTS(w.EndTime),
	}
}

func limitsFromModel(l *pcschema.PromocodeUsageLimits) *promocodepbv1.UsageLimits {
	if l == nil {
		return nil
	}
	out := &promocodepbv1.UsageLimits{}
	if l.MaxRedemptions != nil {
		out.MaxRedemptions = wrapperspb.Int64(int64(*l.MaxRedemptions))
	}
	if l.PerCustomerLimit != nil {
		out.PerCustomerLimit = wrapperspb.Int32(*l.PerCustomerLimit)
	}
	return out
}

func scopeFromParts(s *pcschema.PromocodeScopes, minSub *commonschema.CommonMoneys, res []pcschema.PromocodeScopeApplicableProperties, off []pcschema.PromocodeScopeApplicableUnits) *promocodepbv1.Scope {
	if s == nil {
		return nil
	}
	out := &promocodepbv1.Scope{MinSubtotal: moneyFromModel(minSub)}
	for i := range res {
		out.ApplicableProperties = append(out.ApplicableProperties, res[i].PropertyId)
	}
	for i := range off {
		out.ApplicableUnits = append(out.ApplicableUnits, off[i].UnitName)
	}
	return out
}
