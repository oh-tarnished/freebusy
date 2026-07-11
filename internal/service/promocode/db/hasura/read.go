// List and row hydration.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopeapplicablepropertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopeapplicableunitsql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
)

// List returns a page of promo codes ordered by params.OrderBy and narrowed by
// params.Filter. It fetches one extra row to detect a further page, then hydrates
// each resource's children.
func (r *PromoCodeRepository) List(ctx context.Context, in repox.ListInput) ([]*promocodepbv1.PromoCode, string, error) {
	conds, err := filterx.Parse(in.Filter)
	if err != nil {
		return nil, "", repox.MapFilterxErr(err)
	}
	rows, next, err := filterx.Hasura(promocode.PromoCodeFilterSpec, r.svc.Query.Promocode.Resource).
		List(ctx, filterx.ListInput{
			PageSize:  in.PageSize,
			PageToken: in.PageToken,
			OrderBy:   in.OrderBy,
			Filter:    conds,
		})
	if err != nil {
		return nil, "", dbutil.MapHasuraErr(repox.MapFilterxErr(err))
	}

	items := make([]*promocodepbv1.PromoCode, 0, len(rows))
	for i := range rows {
		pc, err := r.hydrate(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, pc)
	}
	return items, next, nil
}

// hydrate fetches a resource row's child rows (discount + amount money, window,
// limits, scope + min money + applicable join rows) and assembles the proto.
func (r *PromoCodeRepository) hydrate(ctx context.Context, res *resourceql.PromocodeResource) (*promocodepbv1.PromoCode, error) {
	p, _, err := r.fetchParts(ctx, res)
	if err != nil {
		return nil, err
	}
	return fromParts(p), nil
}

// fetchParts loads the child rows for res and also returns the ids of those rows
// (refs) so writers can delete the superseded ones.
func (r *PromoCodeRepository) fetchParts(ctx context.Context, res *resourceql.PromocodeResource) (parts, promoRefs, error) {
	p := parts{res: res}
	refs := promoRefs{discountID: res.DiscountId, windowID: res.WindowId, limitsID: res.LimitsId, scopeID: res.ScopeId}

	d, err := r.svc.Query.Promocode.Discounts.Get(ctx, res.DiscountId)
	if err != nil {
		return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
	}
	p.discount = d
	if d != nil && d.AmountOffId != nil {
		m, err := r.svc.Query.Common.Moneys.Get(ctx, *d.AmountOffId)
		if err != nil {
			return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
		}
		p.amountOff = m
		refs.moneyIDs = append(refs.moneyIDs, *d.AmountOffId)
	}

	if res.WindowId != nil {
		w, err := r.svc.Query.Promocode.RedemptionWindows.Get(ctx, *res.WindowId)
		if err != nil {
			return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
		}
		p.window = w
	}
	if res.LimitsId != nil {
		l, err := r.svc.Query.Promocode.UsageLimits.Get(ctx, *res.LimitsId)
		if err != nil {
			return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
		}
		p.limits = l
	}
	if res.ScopeId != nil {
		s, err := r.svc.Query.Promocode.Scopes.Get(ctx, *res.ScopeId)
		if err != nil {
			return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
		}
		p.scope = s
		if s != nil && s.MinSubtotalId != nil {
			m, err := r.svc.Query.Common.Moneys.Get(ctx, *s.MinSubtotalId)
			if err != nil {
				return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
			}
			p.minSub = m
			refs.moneyIDs = append(refs.moneyIDs, *s.MinSubtotalId)
		}
		resRows, err := r.svc.Query.Promocode.ScopeApplicableProperties.List(ctx,
			scopeapplicablepropertiesql.List().Where(scopeapplicablepropertiesql.ScopeId.Eq(*res.ScopeId)))
		if err != nil {
			return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
		}
		p.properties = resRows
		for i := range resRows {
			refs.resourceJoinIDs = append(refs.resourceJoinIDs, resRows[i].Id)
		}
		offRows, err := r.svc.Query.Promocode.ScopeApplicableUnits.List(ctx,
			scopeapplicableunitsql.List().Where(scopeapplicableunitsql.ScopeId.Eq(*res.ScopeId)))
		if err != nil {
			return parts{}, promoRefs{}, dbutil.MapHasuraErr(err)
		}
		p.units = offRows
		for i := range offRows {
			refs.offeringJoinIDs = append(refs.offeringJoinIDs, offRows[i].Id)
		}
	}
	return p, refs, nil
}
