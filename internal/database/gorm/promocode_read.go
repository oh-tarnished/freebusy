package gorm

import (
	"context"
	"errors"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/protobuf/generated/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/protobuf/generated/gorm/gormx"
	"gorm.io/gorm"
)

// getByID loads and hydrates a single promo code by its bare id.
func (r *PromoCodeRepository) getByID(ctx context.Context, id string) (*promocodepbv1.PromoCode, error) {
	model, err := promocode.NewPromoCodeStore(r.db).GetByID(ctx, id)
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.hydrate(ctx, model)
}

// hydrate resolves a stored model's Money value-objects and applicable lists and
// converts the whole into a protobuf PromoCode.
func (r *PromoCodeRepository) hydrate(ctx context.Context, model *promocode.PromoCode) (*promocodepbv1.PromoCode, error) {
	amount, err := r.loadMoney(ctx, model.AmountOffID)
	if err != nil {
		return nil, err
	}
	minSub, err := r.loadMoney(ctx, model.MinSubtotalID)
	if err != nil {
		return nil, err
	}
	resNames, offNames, err := r.loadApplicable(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	return fromPromoModel(model, amount, minSub, resNames, offNames), nil
}

// loadMoney fetches a Money row by foreign key, treating a missing row as nil
// rather than an error (the value-object is optional).
func (r *PromoCodeRepository) loadMoney(ctx context.Context, id *string) (*booking.Money, error) {
	if id == nil || *id == "" {
		return nil, nil
	}
	m, err := booking.NewMoneyStore(r.db).GetByID(ctx, *id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, mapGormErr(err)
	}
	return m, nil
}

// loadMoneyMap fetches, in a single IN query, every Money row referenced by the
// page of models, keyed by id. It backs the batched List path (no per-row lookups).
func (r *PromoCodeRepository) loadMoneyMap(ctx context.Context, models []promocode.PromoCode) (map[string]*booking.Money, error) {
	ids := make([]string, 0, len(models)*2)
	for i := range models {
		if id := models[i].AmountOffID; id != nil && *id != "" {
			ids = append(ids, *id)
		}
		if id := models[i].MinSubtotalID; id != nil && *id != "" {
			ids = append(ids, *id)
		}
	}
	if len(ids) == 0 {
		return map[string]*booking.Money{}, nil
	}
	var rows []booking.Money
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, mapGormErr(err)
	}
	out := make(map[string]*booking.Money, len(rows))
	for i := range rows {
		out[rows[i].ID] = &rows[i]
	}
	return out, nil
}

// resourceJoinNames / offeringJoinNames extract the stored names from preloaded
// join rows.
func resourceJoinNames(rows []promocode.PromoCodeApplicableResources) []string {
	names := make([]string, 0, len(rows))
	for i := range rows {
		names = append(names, rows[i].ResourceID)
	}
	return names
}

func offeringJoinNames(rows []promocode.PromoCodeApplicableOfferings) []string {
	names := make([]string, 0, len(rows))
	for i := range rows {
		names = append(names, rows[i].OfferingID)
	}
	return names
}

// loadApplicable returns the applicable-resources and applicable-offerings names
// stored against the promo code.
func (r *PromoCodeRepository) loadApplicable(ctx context.Context, promoID string) (resNames, offNames []string, err error) {
	resRows, err := promocode.NewPromoCodeApplicableResourcesStore(r.db).ListByPromoCodeID(ctx, promoID, gormx.ListOptions{})
	if err != nil {
		return nil, nil, mapGormErr(err)
	}
	for _, row := range resRows {
		resNames = append(resNames, row.ResourceID)
	}
	offRows, err := promocode.NewPromoCodeApplicableOfferingsStore(r.db).ListByPromoCodeID(ctx, promoID, gormx.ListOptions{})
	if err != nil {
		return nil, nil, mapGormErr(err)
	}
	for _, row := range offRows {
		offNames = append(offNames, row.OfferingID)
	}
	return resNames, offNames, nil
}

// mapGormErr translates GORM sentinels into repository sentinels so the service
// layer stays free of storage-specific error types.
func mapGormErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return types.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return types.ErrAlreadyExists
	default:
		return err
	}
}
