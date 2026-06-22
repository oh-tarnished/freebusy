package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/protobuf/generated/gorm/freebusy/promocode"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
	"gorm.io/gorm"
)

// Update applies the masked fields of pc to the stored record and returns the
// result. An empty paths slice replaces every mutable field. pc.Etag, when set,
// guards against concurrent writes (types.ErrConflict on mismatch). The
// whole update — scalars, Money value-objects, and join rows — runs in one
// transaction.
func (r *PromoCodeRepository) Update(ctx context.Context, pc *promocodepbv1.PromoCode, paths []string) (*promocodepbv1.PromoCode, error) {
	id, err := types.PromoCodeID(pc.GetName())
	if err != nil {
		return nil, err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		store := promocode.NewPromoCodeStore(tx)
		existing, e := store.GetByID(ctx, id)
		if e != nil {
			return e
		}
		if pc.GetEtag() != "" && existing.Etag != nil && pc.GetEtag() != *existing.Etag {
			return types.ErrConflict
		}

		applyMask(existing, pc, paths)
		existing.Etag = ptr(ulid.GenerateString())

		var staleMoney []string
		if inMask(paths, "amount_off") {
			stale, e := replaceMoney(ctx, tx, &existing.AmountOffID, pc.GetAmountOff())
			if e != nil {
				return e
			}
			if stale != "" {
				staleMoney = append(staleMoney, stale)
			}
		}
		if inMask(paths, "min_subtotal") {
			stale, e := replaceMoney(ctx, tx, &existing.MinSubtotalID, pc.GetMinSubtotal())
			if e != nil {
				return e
			}
			if stale != "" {
				staleMoney = append(staleMoney, stale)
			}
		}
		if e := store.Update(ctx, existing); e != nil {
			return e
		}
		// Delete superseded Money rows only now that the promo no longer points at them.
		stales := booking.NewMoneyStore(tx)
		for _, mid := range staleMoney {
			_ = stales.DeleteByID(ctx, mid)
		}
		if inMask(paths, "applicable_resources") {
			if e := replaceApplicableResources(ctx, tx, id, pc.GetApplicableResources()); e != nil {
				return e
			}
		}
		if inMask(paths, "applicable_offerings") {
			if e := replaceApplicableOfferings(ctx, tx, id, pc.GetApplicableOfferings()); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.Get(ctx, pc.GetName())
}

// Delete removes the promo code, its join rows, and its Money value-objects.
func (r *PromoCodeRepository) Delete(ctx context.Context, name string) error {
	id, err := types.PromoCodeID(name)
	if err != nil {
		return err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		store := promocode.NewPromoCodeStore(tx)
		existing, e := store.GetByID(ctx, id)
		if e != nil {
			return e
		}
		if e := replaceApplicableResources(ctx, tx, id, nil); e != nil {
			return e
		}
		if e := replaceApplicableOfferings(ctx, tx, id, nil); e != nil {
			return e
		}
		if e := store.DeleteByID(ctx, id); e != nil {
			return e
		}
		ms := booking.NewMoneyStore(tx)
		if existing.AmountOffID != nil {
			_ = ms.DeleteByID(ctx, *existing.AmountOffID)
		}
		if existing.MinSubtotalID != nil {
			_ = ms.DeleteByID(ctx, *existing.MinSubtotalID)
		}
		return nil
	})
	return mapGormErr(err)
}

// replaceMoney inserts a new Money row for m (when non-nil) and repoints *fk at
// it, returning the id of the now-superseded row ("" when there was none). The
// caller deletes stale rows only AFTER the new foreign key is persisted, so the
// promo never references a deleted row — safe even if FK constraints are enabled.
func replaceMoney(ctx context.Context, tx *gorm.DB, fk **string, m *money.Money) (stale string, err error) {
	if *fk != nil {
		stale = **fk
	}
	row := moneyToModel(m)
	if row == nil {
		*fk = nil
		return stale, nil
	}
	if e := booking.NewMoneyStore(tx).Create(ctx, row); e != nil {
		return "", e
	}
	*fk = &row.ID
	return stale, nil
}

// replaceApplicableResources deletes the promo code's resource join rows and
// recreates them from names (nil clears them).
func replaceApplicableResources(ctx context.Context, tx *gorm.DB, promoID string, names []string) error {
	if e := tx.WithContext(ctx).Where("promo_code_id = ?", promoID).Delete(&promocode.PromoCodeApplicableResources{}).Error; e != nil {
		return e
	}
	store := promocode.NewPromoCodeApplicableResourcesStore(tx)
	for _, row := range buildApplicableResources(promoID, names) {
		if e := store.Create(ctx, row); e != nil {
			return e
		}
	}
	return nil
}

// replaceApplicableOfferings deletes the promo code's offering join rows and
// recreates them from names (nil clears them).
func replaceApplicableOfferings(ctx context.Context, tx *gorm.DB, promoID string, names []string) error {
	if e := tx.WithContext(ctx).Where("promo_code_id = ?", promoID).Delete(&promocode.PromoCodeApplicableOfferings{}).Error; e != nil {
		return e
	}
	store := promocode.NewPromoCodeApplicableOfferingsStore(tx)
	for _, row := range buildApplicableOfferings(promoID, names) {
		if e := store.Create(ctx, row); e != nil {
			return e
		}
	}
	return nil
}
