package hasura

import (
	"context"
	"time"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	pcschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"github.com/oh-tarnished/generateql/runtime/go/runtime"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// promoRefs captures the ids of a promo code's child rows, Money rows, and scope
// join rows so a writer can delete the superseded ones after the resource no
// longer references them.
type promoRefs struct {
	discountID      string
	windowID        *string
	limitsID        *string
	scopeID         *string
	moneyIDs        []string
	resourceJoinIDs []string
	offeringJoinIDs []string
}

// Update applies the masked fields of pc to the stored record and returns the
// result. An empty paths slice replaces every mutable field; pc.Etag, when set,
// guards against concurrent writes (types.ErrConflict on mismatch). The whole
// update runs as one atomic mutation batch: the merged proto is re-materialized
// into a fresh child graph, the resource row is repointed at it, and the
// superseded child / Money / join rows are deleted in the same transaction.
func (r *PromoCodeRepository) Update(ctx context.Context, pc *promocodepbv1.PromoCode, paths []string) (*promocodepbv1.PromoCode, error) {
	id, err := types.PromoCodeID(pc.GetName())
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
	if pc.GetEtag() != "" && res.Etag != nil && pc.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}

	p, old, err := r.fetchParts(ctx, res)
	if err != nil {
		return nil, err
	}

	// Merge the masked fields onto a proto built from the stored record, then
	// rebuild the whole child graph from the result.
	merged := fromParts(p)
	applyMask(merged, pc, paths)
	now := time.Now().UTC()
	g := buildGraph(merged, now)

	tx := r.svc.Mutation.Tx()

	// 1. Insert the new Money rows and children.
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
		resRes := make([]pcschema.InsertPromocodeScopeApplicableResourcesResponse, len(g.resources))
		for i, row := range g.resources {
			tx.Add(r.svc.Mutation.Promocode.ScopeApplicableResources.CreateOp(row, &resRes[i]))
		}
		offRes := make([]pcschema.InsertPromocodeScopeApplicableOfferingsResponse, len(g.offerings))
		for i, row := range g.offerings {
			tx.Add(r.svc.Mutation.Promocode.ScopeApplicableOfferings.CreateOp(row, &offRes[i]))
		}
	}

	// 2. Repoint the resource row at the new children and bump the etag.
	patch := resourceql.UpdateInput{
		Code:        graphql.Value(g.resource.Code),
		DisplayName: nullableStr(g.resource.DisplayName),
		Description: nullableStr(g.resource.Description),
		Disabled:    graphql.Value(g.resource.Disabled),
		State:       graphql.Value(g.resource.State),
		DiscountId:  graphql.Value(g.resource.DiscountId),
		WindowId:    nullableStr(g.resource.WindowId),
		LimitsId:    nullableStr(g.resource.LimitsId),
		ScopeId:     nullableStr(g.resource.ScopeId),
		Etag:        graphql.Value(ulid.GenerateString()),
		UpdateTime:  graphql.Value(tsToStr(timestamppb.New(now))),
	}
	var updRes pcschema.UpdatePromocodeResourceByIdResponse
	tx.Add(r.svc.Mutation.Promocode.Resource.UpdateOp(id, patch, &updRes))

	// 3. Delete the superseded children in foreign-key order.
	queueChildDeletes(tx, r, old)

	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.Get(ctx, pc.GetName())
}

// Delete removes the promo code, its child rows, their Money value-objects, and
// the scope's join rows, in one atomic mutation batch.
func (r *PromoCodeRepository) Delete(ctx context.Context, name string) error {
	id, err := types.PromoCodeID(name)
	if err != nil {
		return err
	}
	res, err := r.svc.Query.Promocode.Resource.Get(ctx, id)
	if err != nil {
		return mapHasuraErr(err)
	}
	if res == nil {
		return types.ErrNotFound
	}
	_, refs, err := r.fetchParts(ctx, res)
	if err != nil {
		return err
	}

	tx := r.svc.Mutation.Tx()
	// Delete the resource first (its redemptions cascade in the DB), then the
	// now-unreferenced children.
	var delRes pcschema.DeletePromocodeResourceByIdResponse
	tx.Add(r.svc.Mutation.Promocode.Resource.DeleteOp(id, &delRes))
	queueChildDeletes(tx, r, refs)

	return mapHasuraErr(tx.Commit(ctx))
}

// queueChildDeletes appends deletes for a promo code's child rows to tx in
// foreign-key order: the scope's join rows, then the scope / window / limits /
// discount, then the Money rows those children referenced.
func queueChildDeletes(tx *runtime.Tx, r *PromoCodeRepository, refs promoRefs) {
	for _, jid := range refs.resourceJoinIDs {
		var out pcschema.DeletePromocodeScopeApplicableResourcesByIdResponse
		tx.Add(r.svc.Mutation.Promocode.ScopeApplicableResources.DeleteOp(jid, &out))
	}
	for _, jid := range refs.offeringJoinIDs {
		var out pcschema.DeletePromocodeScopeApplicableOfferingsByIdResponse
		tx.Add(r.svc.Mutation.Promocode.ScopeApplicableOfferings.DeleteOp(jid, &out))
	}
	if refs.scopeID != nil {
		var out pcschema.DeletePromocodeScopesByIdResponse
		tx.Add(r.svc.Mutation.Promocode.Scopes.DeleteOp(*refs.scopeID, &out))
	}
	if refs.windowID != nil {
		var out pcschema.DeletePromocodeRedemptionWindowsByIdResponse
		tx.Add(r.svc.Mutation.Promocode.RedemptionWindows.DeleteOp(*refs.windowID, &out))
	}
	if refs.limitsID != nil {
		var out pcschema.DeletePromocodeUsageLimitsByIdResponse
		tx.Add(r.svc.Mutation.Promocode.UsageLimits.DeleteOp(*refs.limitsID, &out))
	}
	if refs.discountID != "" {
		var out pcschema.DeletePromocodeDiscountsByIdResponse
		tx.Add(r.svc.Mutation.Promocode.Discounts.DeleteOp(refs.discountID, &out))
	}
	for _, mid := range refs.moneyIDs {
		var out commonschema.DeleteCommonMoneysByIdResponse
		tx.Add(r.svc.Mutation.Common.Moneys.DeleteOp(mid, &out))
	}
}

// nullableStr maps an empty optional string to a SQL NULL update and a non-empty
// one to a value update, so clearing a field in the proto clears the column.
func nullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}
