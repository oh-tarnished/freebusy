// Package repository defines the provider-agnostic persistence interfaces for the
// freebusy domain. Each interface speaks protobuf domain types (the wire/domain
// currency); concrete implementations live in sibling packages
// (internal/database/gorm and internal/database/hasura) and the factory in
// internal/database selects one at runtime. Shared, provider-neutral vocabulary
// (errors, list params, names, field masks) lives in internal/types.
package repository

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/promocode/v1/promocodepbv1"
)

// Provider identifies a database backend implementation.
type Provider string

const (
	// ProviderGorm is the default backend: GORM over the relational database.
	ProviderGorm Provider = "gorm"
	// ProviderHasura is the opt-in backend: Hasura GraphQL.
	ProviderHasura Provider = "hasura"
)

// PromoCodeRepository provides CRUD persistence for promo codes. All methods
// accept and return the protobuf PromoCode type; implementations own the
// conversion to and from their storage models, including the normalized Money
// value-objects and the applicable-resources / applicable-offerings join rows.
// Errors are the sentinels in internal/types (types.ErrNotFound, etc.).
type PromoCodeRepository interface {
	// Create persists pc and returns the stored record with server-set fields
	// (name, timestamps, etag, state) populated.
	Create(ctx context.Context, pc *promocodepbv1.PromoCode) (*promocodepbv1.PromoCode, error)

	// Get returns the promo code identified by its resource name
	// ("promoCodes/{promo_code}"), or types.ErrNotFound.
	Get(ctx context.Context, name string) (*promocodepbv1.PromoCode, error)

	// FindByCode returns the promo code with the given human-entered code (e.g.
	// "SUMMER25"), or types.ErrNotFound. Used by validation/redemption flows that
	// address a code rather than a resource name.
	FindByCode(ctx context.Context, code string) (*promocodepbv1.PromoCode, error)

	// List returns a page of promo codes and an opaque token for the next page
	// (empty when there are no further results).
	List(ctx context.Context, params types.ListParams) (items []*promocodepbv1.PromoCode, nextPageToken string, err error)

	// Update persists the fields named by paths (an AIP-134 field mask); an empty
	// paths slice means a full replace. The record is identified by pc.Name and
	// pc.Etag guards against concurrent writes (types.ErrConflict on mismatch).
	Update(ctx context.Context, pc *promocodepbv1.PromoCode, paths []string) (*promocodepbv1.PromoCode, error)

	// Delete removes the promo code identified by its resource name, returning
	// types.ErrNotFound when it does not exist.
	Delete(ctx context.Context, name string) error
}
