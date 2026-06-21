// Package hasura provides the Hasura/GraphQL-backed implementations of the
// freebusy repository interfaces. It adapts the generated freebusyql handlers to
// the provider-agnostic contracts in internal/database/repository, converting
// between protobuf domain types and the GraphQL schema types.
package hasura

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/moneysql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/applicableofferingsql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/applicableresourcesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/promocodeql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/database/repository"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"github.com/oh-tarnished/generateql/runtime/go/graphql"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/money"
)

// PromoCodeRepository is the Hasura-backed repository.PromoCodeRepository. Each
// write fans out across the promo code resource mutation, the Money mutations in
// the booking schema, and the applicable-resources / -offerings join mutations;
// each read uses a single byId query that resolves all of those via relations.
type PromoCodeRepository struct {
	svc *freebusyql.Service
}

var _ repository.PromoCodeRepository = (*PromoCodeRepository)(nil)

// NewPromoCodeRepository returns a Hasura-backed PromoCodeRepository bound to svc.
func NewPromoCodeRepository(svc *freebusyql.Service) repository.PromoCodeRepository {
	return &PromoCodeRepository{svc: svc}
}

// Create inserts the Money value-objects, the promo code resource, and the join
// rows, then re-reads the stored record.
func (r *PromoCodeRepository) Create(ctx context.Context, pc *promocodepbv1.PromoCode) (*promocodepbv1.PromoCode, error) {
	id, name, err := repository.ResolvePromoCodeName(pc.GetName())
	if err != nil {
		return nil, err
	}
	// Hasura has no cross-mutation transaction, so each step compensates the
	// previous ones on failure (best-effort) to avoid orphaned Money/resource rows.
	amountID, err := r.insertMoney(ctx, pc.GetAmountOff())
	if err != nil {
		return nil, err
	}
	minID, err := r.insertMoney(ctx, pc.GetMinSubtotal())
	if err != nil {
		r.cleanupMoney(ctx, amountID)
		return nil, err
	}

	input := toCreateInput(pc, id, name, ulid.GenerateString(), amountID, minID)
	if _, err := r.svc.Mutation.Promocode.Resource.Create(ctx, input); err != nil {
		r.cleanupMoney(ctx, amountID, minID)
		return nil, err
	}
	if err := r.createJoins(ctx, id, pc.GetApplicableResources(), pc.GetApplicableOfferings()); err != nil {
		_, _ = r.svc.Mutation.Promocode.Resource.Delete(ctx, id)
		r.cleanupMoney(ctx, amountID, minID)
		return nil, err
	}
	return r.Get(ctx, name)
}

// cleanupMoney best-effort deletes Money rows inserted during a Create that later
// failed, so a partial fan-out doesn't leave them orphaned.
func (r *PromoCodeRepository) cleanupMoney(ctx context.Context, ids ...string) {
	for _, id := range ids {
		if id != "" {
			_, _ = r.svc.Mutation.Booking.Moneys.Delete(ctx, id)
		}
	}
}

// Get reads a promo code (with its Money and applicable relations) by name.
func (r *PromoCodeRepository) Get(ctx context.Context, name string) (*promocodepbv1.PromoCode, error) {
	id, err := repository.PromoCodeID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Promocode.Resource.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, repository.ErrNotFound
	}
	return resourceFromModel(res), nil
}

// FindByCode returns the promo code with the given human-entered code.
func (r *PromoCodeRepository) FindByCode(ctx context.Context, code string) (*promocodepbv1.PromoCode, error) {
	res, err := r.svc.Query.Promocode.Resource.Find(ctx, resourceql.List().Where(resourceql.Code.Eq(code)))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, repository.ErrNotFound
	}
	return resourceFromModel(res), nil
}

// List returns a page of promo codes, fetching one extra row to detect whether a
// further page exists.
func (r *PromoCodeRepository) List(ctx context.Context, params repository.ListParams) ([]*promocodepbv1.PromoCode, string, error) {
	order, err := orderTerms(params.OrderBy)
	if err != nil {
		return nil, "", err
	}
	limit, offset := repository.PageBounds(params)
	req := resourceql.List().Limit(limit + 1).Offset(offset)
	if len(order) > 0 {
		req = req.OrderBy(order...)
	}
	rows, err := r.svc.Query.Promocode.Resource.List(ctx, req)
	if err != nil {
		return nil, "", err
	}

	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		next = repository.EncodeOffset(offset + limit)
	}

	items := make([]*promocodepbv1.PromoCode, 0, len(rows))
	for i := range rows {
		items = append(items, resourceFromModel(&rows[i]))
	}
	return items, next, nil
}

// insertMoney inserts a Money row for m (when non-nil) and returns its new id, or
// "" when m is nil.
func (r *PromoCodeRepository) insertMoney(ctx context.Context, m *money.Money) (string, error) {
	if m == nil {
		return "", nil
	}
	id := ulid.GenerateString()
	_, err := r.svc.Mutation.Booking.Moneys.Create(ctx, moneysql.CreateInput{
		Id:           id,
		CurrencyCode: m.GetCurrencyCode(),
		Nanos:        m.GetNanos(),
		Units:        graphql.Int64(m.GetUnits()),
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// createJoins inserts the applicable-resources and applicable-offerings join rows
// for the promo code, storing each full API name verbatim.
func (r *PromoCodeRepository) createJoins(ctx context.Context, promoID string, resNames, offNames []string) error {
	for _, name := range resNames {
		if _, err := r.svc.Mutation.Promocode.ApplicableResources.Create(ctx, applicableresourcesql.CreateInput{
			Id:          ulid.GenerateString(),
			PromoCodeId: promoID,
			ResourceId:  name,
		}); err != nil {
			return err
		}
	}
	for _, name := range offNames {
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

// inMask aliases the shared field-mask predicate so update semantics stay
// identical across the gorm and hasura adapters.
var inMask = repository.InMask
