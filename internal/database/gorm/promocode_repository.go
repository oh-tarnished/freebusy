// Package gorm provides the GORM-backed implementations of the freebusy
// repository interfaces. It adapts the generated per-entity stores under
// internal/database/gorm/freebusy/... to the provider-agnostic contracts in
// internal/database/repository, converting between protobuf domain types and the
// relational storage models.
package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/promocode"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// PromoCodeRepository is the GORM-backed repository.PromoCodeRepository. Each
// write persists the promo code together with its normalized Money value-objects
// (booking schema) and applicable-resources / -offerings join rows inside a
// single transaction; each read re-hydrates them (see promocode_read.go).
type PromoCodeRepository struct {
	db *gorm.DB
}

var _ repository.PromoCodeRepository = (*PromoCodeRepository)(nil)

// NewPromoCodeRepository returns a GORM-backed PromoCodeRepository bound to db.
func NewPromoCodeRepository(db *gorm.DB) repository.PromoCodeRepository {
	return &PromoCodeRepository{db: db}
}

// Create persists pc and returns the stored record. The resource name is taken
// from pc.Name when present, otherwise a fresh ULID id is assigned.
func (r *PromoCodeRepository) Create(ctx context.Context, pc *promocodepbv1.PromoCode) (*promocodepbv1.PromoCode, error) {
	id, name, err := repository.ResolvePromoCodeName(pc.GetName())
	if err != nil {
		return nil, err
	}

	model := toPromoModel(pc)
	model.ID = id
	model.Name = name
	model.Etag = ptr(ulid.GenerateString())

	amount := moneyToModel(pc.GetAmountOff())
	if amount != nil {
		model.AmountOffID = &amount.ID
	}
	minSub := moneyToModel(pc.GetMinSubtotal())
	if minSub != nil {
		model.MinSubtotalID = &minSub.ID
	}
	resRows := buildApplicableResources(id, pc.GetApplicableResources())
	offRows := buildApplicableOfferings(id, pc.GetApplicableOfferings())

	err = r.db.Transaction(func(tx *gorm.DB) error {
		money := booking.NewMoneyStore(tx)
		if amount != nil {
			if e := money.Create(ctx, amount); e != nil {
				return e
			}
		}
		if minSub != nil {
			if e := money.Create(ctx, minSub); e != nil {
				return e
			}
		}
		if e := promocode.NewPromoCodeStore(tx).Create(ctx, model); e != nil {
			return e
		}
		res := promocode.NewPromoCodeApplicableResourcesStore(tx)
		for _, row := range resRows {
			if e := res.Create(ctx, row); e != nil {
				return e
			}
		}
		off := promocode.NewPromoCodeApplicableOfferingsStore(tx)
		for _, row := range offRows {
			if e := off.Create(ctx, row); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	return r.Get(ctx, name)
}

// Get returns the promo code addressed by its resource name.
func (r *PromoCodeRepository) Get(ctx context.Context, name string) (*promocodepbv1.PromoCode, error) {
	id, err := repository.PromoCodeID(name)
	if err != nil {
		return nil, err
	}
	return r.getByID(ctx, id)
}

// FindByCode returns the promo code with the given human-entered code.
func (r *PromoCodeRepository) FindByCode(ctx context.Context, code string) (*promocodepbv1.PromoCode, error) {
	models, err := promocode.NewPromoCodeStore(r.db).List(ctx, promocode.ListOptions{
		Where: "code = ?",
		Args:  []any{code},
		Limit: 1,
	})
	if err != nil {
		return nil, mapGormErr(err)
	}
	if len(models) == 0 {
		return nil, repository.ErrNotFound
	}
	return r.hydrate(ctx, &models[0])
}

// List returns a page of promo codes ordered by params.OrderBy. It fetches one
// extra row to decide whether a further page exists. To avoid an N+1, the join
// rows are preloaded in two batch queries and every page's Money value-objects are
// fetched with a single IN query (rather than per-row lookups).
func (r *PromoCodeRepository) List(ctx context.Context, params repository.ListParams) ([]*promocodepbv1.PromoCode, string, error) {
	order, err := orderClause(params.OrderBy)
	if err != nil {
		return nil, "", err
	}
	limit, offset := repository.PageBounds(params)

	q := r.db.WithContext(ctx).
		Preload("ApplicableResources").
		Preload("ApplicableOfferings").
		Limit(limit + 1).Offset(offset)
	if order != "" {
		q = q.Order(order)
	}
	var models []promocode.PromoCode
	if err := q.Find(&models).Error; err != nil {
		return nil, "", mapGormErr(err)
	}

	next := ""
	if len(models) > limit {
		models = models[:limit]
		next = repository.EncodeOffset(offset + limit)
	}

	moneys, err := r.loadMoneyMap(ctx, models)
	if err != nil {
		return nil, "", err
	}
	items := make([]*promocodepbv1.PromoCode, 0, len(models))
	for i := range models {
		m := &models[i]
		items = append(items, fromPromoModel(m,
			moneys[deref(m.AmountOffID)], moneys[deref(m.MinSubtotalID)],
			resourceJoinNames(m.ApplicableResources), offeringJoinNames(m.ApplicableOfferings)))
	}
	return items, next, nil
}
