// Package hasura provides the Hasura/GraphQL-backed implementation of the
// promocode persistence contract (internal/service/promocode/db.PromoCodeRepository).
// It adapts the generated freebusyql handlers to that contract, converting
// between protobuf domain types and the normalized GraphQL schema (the discount,
// redemption window, usage limits, and scope child tables, their common.moneys
// value-objects, and the scope's applicable-resources / -offerings join rows).
package hasura

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	pcschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopeapplicablepropertiesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/scopeapplicableunitsql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
)

// PromoCodeRepository is the Hasura-backed promo code repository. Hasura exposes
// no client-side transactions across separate calls, but its mutation API can run
// several mutations as one atomic GraphQL document (svc.Mutation.Tx()); writes
// use that so a promo code and its child rows commit together or not at all. Each
// read hydrates the resource row's children with follow-up queries (the schema
// exposes only foreign-key ids on the resource, not nested relations).
type PromoCodeRepository struct {
	svc *freebusyql.Service
}

// NewPromoCodeRepository returns a Hasura-backed PromoCodeRepository bound to svc.
// The parent db package asserts it satisfies db.PromoCodeRepository.
func NewPromoCodeRepository(svc *freebusyql.Service) *PromoCodeRepository {
	return &PromoCodeRepository{svc: svc}
}

// Create persists pc and its child graph as a single atomic mutation batch, then
// re-reads the stored record. The resource name is taken from pc.Name when
// present, otherwise a fresh ULID id is assigned. Rows are queued in foreign-key
// order — Money rows and children before the resource that references them, the
// scope's join rows after the scope — so the in-transaction insert never violates
// a constraint.
func (r *PromoCodeRepository) Create(ctx context.Context, pc *promocodepbv1.PromoCode) (*promocodepbv1.PromoCode, error) {
	id, name, err := types.ResolvePromoCodeName(pc.GetName())
	if err != nil {
		return nil, err
	}

	g := buildGraph(pc, time.Now().UTC())
	g.resource.Id = id
	g.resource.Name = name
	g.resource.Etag = ulid.GenerateString()

	tx := r.svc.Mutation.Tx()

	moneyRes := make([]commonschema.InsertCommonMoneysResponse, len(g.moneys))
	for i, m := range g.moneys {
		tx.Add(r.svc.Mutation.Common.Moneys.CreateOp(m, &moneyRes[i]))
	}
	var discRes pcschema.InsertPromocodeDiscountsResponse
	tx.Add(r.svc.Mutation.Promocode.Discounts.CreateOp(g.discount, &discRes))
	if g.window != nil {
		var wRes pcschema.InsertPromocodeRedemptionWindowsResponse
		tx.Add(r.svc.Mutation.Promocode.RedemptionWindows.CreateOp(*g.window, &wRes))
	}
	if g.limits != nil {
		var lRes pcschema.InsertPromocodeUsageLimitsResponse
		tx.Add(r.svc.Mutation.Promocode.UsageLimits.CreateOp(*g.limits, &lRes))
	}
	if g.scope != nil {
		var sRes pcschema.InsertPromocodeScopesResponse
		tx.Add(r.svc.Mutation.Promocode.Scopes.CreateOp(*g.scope, &sRes))
		resRes := make([]pcschema.InsertPromocodeScopeApplicablePropertiesResponse, len(g.properties))
		for i, row := range g.properties {
			tx.Add(r.svc.Mutation.Promocode.ScopeApplicableProperties.CreateOp(row, &resRes[i]))
		}
		offRes := make([]pcschema.InsertPromocodeScopeApplicableUnitsResponse, len(g.units))
		for i, row := range g.units {
			tx.Add(r.svc.Mutation.Promocode.ScopeApplicableUnits.CreateOp(row, &offRes[i]))
		}
	}
	var resourceRes pcschema.InsertPromocodeResourceResponse
	tx.Add(r.svc.Mutation.Promocode.Resource.CreateOp(g.resource, &resourceRes))

	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.Get(ctx, name)
}

// Get returns the promo code addressed by its resource name.
func (r *PromoCodeRepository) Get(ctx context.Context, name string) (*promocodepbv1.PromoCode, error) {
	id, err := types.PromoCodeID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Promocode.Resource.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return r.hydrate(ctx, res)
}

// FindByCode returns the promo code with the given human-entered code.
func (r *PromoCodeRepository) FindByCode(ctx context.Context, code string) (*promocodepbv1.PromoCode, error) {
	res, err := r.svc.Query.Promocode.Resource.Find(ctx, resourceql.List().Where(resourceql.Code.Eq(code)))
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return r.hydrate(ctx, res)
}

// List returns a page of promo codes ordered by params.OrderBy and narrowed by
// params.Filter. It fetches one extra row to detect a further page, then hydrates
// each resource's children.
func (r *PromoCodeRepository) List(ctx context.Context, params types.ListParams) ([]*promocodepbv1.PromoCode, string, error) {
	rows, next, err := filterx.Hasura[pcschema.PromocodeResource](promocode.PromoCodeFilterSpec, r.svc.Query.Promocode.Resource).
		List(ctx, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
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
func (r *PromoCodeRepository) hydrate(ctx context.Context, res *pcschema.PromocodeResource) (*promocodepbv1.PromoCode, error) {
	p, _, err := r.fetchParts(ctx, res)
	if err != nil {
		return nil, err
	}
	return fromParts(p), nil
}

// fetchParts loads the child rows for res and also returns the ids of those rows
// (refs) so writers can delete the superseded ones.
func (r *PromoCodeRepository) fetchParts(ctx context.Context, res *pcschema.PromocodeResource) (parts, promoRefs, error) {
	p := parts{res: res}
	refs := promoRefs{discountID: res.DiscountId, windowID: res.WindowId, limitsID: res.LimitsId, scopeID: res.ScopeId}

	d, err := r.svc.Query.Promocode.Discounts.Get(ctx, res.DiscountId)
	if err != nil {
		return parts{}, promoRefs{}, mapHasuraErr(err)
	}
	p.discount = d
	if d != nil && d.AmountOffId != nil {
		m, err := r.svc.Query.Common.Moneys.Get(ctx, *d.AmountOffId)
		if err != nil {
			return parts{}, promoRefs{}, mapHasuraErr(err)
		}
		p.amountOff = m
		refs.moneyIDs = append(refs.moneyIDs, *d.AmountOffId)
	}

	if res.WindowId != nil {
		w, err := r.svc.Query.Promocode.RedemptionWindows.Get(ctx, *res.WindowId)
		if err != nil {
			return parts{}, promoRefs{}, mapHasuraErr(err)
		}
		p.window = w
	}
	if res.LimitsId != nil {
		l, err := r.svc.Query.Promocode.UsageLimits.Get(ctx, *res.LimitsId)
		if err != nil {
			return parts{}, promoRefs{}, mapHasuraErr(err)
		}
		p.limits = l
	}
	if res.ScopeId != nil {
		s, err := r.svc.Query.Promocode.Scopes.Get(ctx, *res.ScopeId)
		if err != nil {
			return parts{}, promoRefs{}, mapHasuraErr(err)
		}
		p.scope = s
		if s != nil && s.MinSubtotalId != nil {
			m, err := r.svc.Query.Common.Moneys.Get(ctx, *s.MinSubtotalId)
			if err != nil {
				return parts{}, promoRefs{}, mapHasuraErr(err)
			}
			p.minSub = m
			refs.moneyIDs = append(refs.moneyIDs, *s.MinSubtotalId)
		}
		resRows, err := r.svc.Query.Promocode.ScopeApplicableProperties.List(ctx,
			scopeapplicablepropertiesql.List().Where(scopeapplicablepropertiesql.ScopeId.Eq(*res.ScopeId)))
		if err != nil {
			return parts{}, promoRefs{}, mapHasuraErr(err)
		}
		p.properties = resRows
		for i := range resRows {
			refs.resourceJoinIDs = append(refs.resourceJoinIDs, resRows[i].Id)
		}
		offRows, err := r.svc.Query.Promocode.ScopeApplicableUnits.List(ctx,
			scopeapplicableunitsql.List().Where(scopeapplicableunitsql.ScopeId.Eq(*res.ScopeId)))
		if err != nil {
			return parts{}, promoRefs{}, mapHasuraErr(err)
		}
		p.units = offRows
		for i := range offRows {
			refs.offeringJoinIDs = append(refs.offeringJoinIDs, offRows[i].Id)
		}
	}
	return p, refs, nil
}

// mapHasuraErr translates GraphQL/runtime errors into the repository sentinels so
// the service layer stays free of storage-specific error types.
func mapHasuraErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, graphql.ErrConflict):
		return types.ErrConflict
	}
	// Best-effort unique-violation detection (the engine surfaces it as a generic
	// constraint error in the GraphQL message).
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") {
		return types.ErrAlreadyExists
	}
	return err
}
