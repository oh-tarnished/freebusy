package hasura

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/applicableofferingsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/applicableresourcesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// Update applies the masked fields of pc and returns the stored result. It reads
// the current row first to honor the etag guard and to learn the Money foreign
// keys and join-row ids it must replace. New Money rows are inserted before the
// resource patch points at them, and the superseded rows are deleted afterwards.
func (r *PromoCodeRepository) Update(ctx context.Context, pc *promocodepbv1.PromoCode, paths []string) (*promocodepbv1.PromoCode, error) {
	id, err := repository.PromoCodeID(pc.GetName())
	if err != nil {
		return nil, err
	}
	existing, err := r.svc.Query.Promocode.Resource.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, repository.ErrNotFound
	}
	if pc.GetEtag() != "" && existing.Etag != nil && pc.GetEtag() != *existing.Etag {
		return nil, repository.ErrConflict
	}

	patch := buildUpdatePatch(pc, paths)
	patch.Etag = ulid.GenerateString()

	var staleMoney []string
	if inMask(paths, "amount_off") {
		newID, err := r.insertMoney(ctx, pc.GetAmountOff())
		if err != nil {
			return nil, err
		}
		patch.AmountOffId = newID
		staleMoney = appendMoneyID(staleMoney, existing.AmountOffId)
	}
	if inMask(paths, "min_subtotal") {
		newID, err := r.insertMoney(ctx, pc.GetMinSubtotal())
		if err != nil {
			return nil, err
		}
		patch.MinSubtotalId = newID
		staleMoney = appendMoneyID(staleMoney, existing.MinSubtotalId)
	}

	// Enforce the etag server-side with a precheck so a concurrent writer that
	// changed it between the read above and this write is caught atomically (the
	// in-app check is only the fast path). A matching key with zero affected rows
	// means the precheck failed, i.e. the etag moved.
	var checks []*resourceql.UpdateRequest
	if pc.GetEtag() != "" {
		checks = append(checks, resourceql.Update().PreCheck(resourceql.Etag.Eq(pc.GetEtag())))
	}
	resp, err := r.svc.Mutation.Promocode.Resource.Update(ctx, id, patch, checks...)
	if err != nil {
		return nil, err
	}
	if pc.GetEtag() != "" && resp.AffectedRows == 0 {
		return nil, repository.ErrConflict
	}
	for _, mid := range staleMoney {
		_, _ = r.svc.Mutation.Booking.Moneys.Delete(ctx, mid)
	}

	if inMask(paths, "applicable_resources") {
		if err := r.replaceResourceJoins(ctx, id, resourceJoinIDs(existing), pc.GetApplicableResources()); err != nil {
			return nil, err
		}
	}
	if inMask(paths, "applicable_offerings") {
		if err := r.replaceOfferingJoins(ctx, id, offeringJoinIDs(existing), pc.GetApplicableOfferings()); err != nil {
			return nil, err
		}
	}
	return r.Get(ctx, pc.GetName())
}

// Delete removes the join rows, the promo code resource, and its Money rows.
func (r *PromoCodeRepository) Delete(ctx context.Context, name string) error {
	id, err := repository.PromoCodeID(name)
	if err != nil {
		return err
	}
	existing, err := r.svc.Query.Promocode.Resource.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return repository.ErrNotFound
	}
	for _, rowID := range resourceJoinIDs(existing) {
		if _, err := r.svc.Mutation.Promocode.ApplicableResources.Delete(ctx, rowID); err != nil {
			return err
		}
	}
	for _, rowID := range offeringJoinIDs(existing) {
		if _, err := r.svc.Mutation.Promocode.ApplicableOfferings.Delete(ctx, rowID); err != nil {
			return err
		}
	}
	if _, err := r.svc.Mutation.Promocode.Resource.Delete(ctx, id); err != nil {
		return err
	}
	for _, mid := range []*string{existing.AmountOffId, existing.MinSubtotalId} {
		if mid != nil && *mid != "" {
			_, _ = r.svc.Mutation.Booking.Moneys.Delete(ctx, *mid)
		}
	}
	return nil
}

// replaceResourceJoins deletes the given join rows and recreates them from names.
func (r *PromoCodeRepository) replaceResourceJoins(ctx context.Context, promoID string, oldRowIDs, names []string) error {
	for _, rowID := range oldRowIDs {
		if _, err := r.svc.Mutation.Promocode.ApplicableResources.Delete(ctx, rowID); err != nil {
			return err
		}
	}
	for _, name := range names {
		if _, err := r.svc.Mutation.Promocode.ApplicableResources.Create(ctx, applicableresourcesql.CreateInput{
			Id:          ulid.GenerateString(),
			PromoCodeId: promoID,
			ResourceId:  name,
		}); err != nil {
			return err
		}
	}
	return nil
}

// replaceOfferingJoins deletes the given join rows and recreates them from names.
func (r *PromoCodeRepository) replaceOfferingJoins(ctx context.Context, promoID string, oldRowIDs, names []string) error {
	for _, rowID := range oldRowIDs {
		if _, err := r.svc.Mutation.Promocode.ApplicableOfferings.Delete(ctx, rowID); err != nil {
			return err
		}
	}
	for _, name := range names {
		if _, err := r.svc.Mutation.Promocode.ApplicableOfferings.Create(ctx, applicableofferingsql.CreateInput{
			Id:          ulid.GenerateString(),
			PromoCodeId: promoID,
			OfferingId:  name,
		}); err != nil {
			return err
		}
	}
	return nil
}

func resourceJoinIDs(m *schemaql.PromocodeResource) []string {
	out := make([]string, 0, len(m.PromocodeApplicableResources))
	for _, row := range m.PromocodeApplicableResources {
		out = append(out, row.Id)
	}
	return out
}

func offeringJoinIDs(m *schemaql.PromocodeResource) []string {
	out := make([]string, 0, len(m.PromocodeApplicableOfferings))
	for _, row := range m.PromocodeApplicableOfferings {
		out = append(out, row.Id)
	}
	return out
}

func appendMoneyID(ids []string, id *string) []string {
	if id != nil && *id != "" {
		return append(ids, *id)
	}
	return ids
}
