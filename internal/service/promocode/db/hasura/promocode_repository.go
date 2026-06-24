// Package hasura provides the Hasura/GraphQL-backed implementation of the
// promocode persistence contract (internal/service/promocode/db.PromoCodeRepository).
// It adapts the generated freebusyql handlers to that contract, converting
// between protobuf domain types and the GraphQL schema types.
package hasura

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PromoCodeRepository is the Hasura-backed promo code repository.
//
// NOTE: the promo code schema was normalized — the flat discount / window /
// limits / scope columns became their own tables (promocode.discounts,
// promocode.redemption_windows, promocode.usage_limits, promocode.scopes, plus
// the redemptions sub-resource). The generated freebusyql client this adapter
// builds on is introspected from the live Hasura instance and still reflects the
// old flat schema, so it has no handlers for the new tables. Until the Hasura
// database is migrated to the normalized schema, its metadata reloaded, and the
// client regenerated (`generateql generate`), this provider returns Unimplemented.
// The GORM provider (the default) implements the new schema fully.
type PromoCodeRepository struct {
	svc *freebusyql.Service
}

// NewPromoCodeRepository returns a Hasura-backed PromoCodeRepository bound to svc.
// The parent db package asserts it satisfies db.PromoCodeRepository.
func NewPromoCodeRepository(svc *freebusyql.Service) *PromoCodeRepository {
	return &PromoCodeRepository{svc: svc}
}

// errUnimplemented is returned by every method until the freebusyql client is
// regenerated against the normalized promo code schema (see the type doc).
var errUnimplemented = status.Error(codes.Unimplemented,
	"hasura promo code provider is pending regeneration against the normalized schema; use the gorm provider")

func (r *PromoCodeRepository) Create(ctx context.Context, pc *promocodepbv1.PromoCode) (*promocodepbv1.PromoCode, error) {
	return nil, errUnimplemented
}

func (r *PromoCodeRepository) Get(ctx context.Context, name string) (*promocodepbv1.PromoCode, error) {
	return nil, errUnimplemented
}

func (r *PromoCodeRepository) FindByCode(ctx context.Context, code string) (*promocodepbv1.PromoCode, error) {
	return nil, errUnimplemented
}

func (r *PromoCodeRepository) List(ctx context.Context, params types.ListParams) ([]*promocodepbv1.PromoCode, string, error) {
	return nil, "", errUnimplemented
}

func (r *PromoCodeRepository) Update(ctx context.Context, pc *promocodepbv1.PromoCode, paths []string) (*promocodepbv1.PromoCode, error) {
	return nil, errUnimplemented
}

func (r *PromoCodeRepository) Delete(ctx context.Context, name string) error {
	return errUnimplemented
}
