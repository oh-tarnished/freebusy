// Package repository defines provider-agnostic persistence interfaces for the
// freebusy domain.
//
// Every interface speaks protobuf domain types (the wire/domain currency), so the
// service layer never depends on a concrete ORM or GraphQL client. Concrete
// implementations live in sibling packages — internal/database/gorm and
// internal/database/hasura — and the factory in internal/database selects one at
// runtime based on the FREEBUSY_DB_PROVIDER environment variable.
package repository

import (
	"context"
	"errors"

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

// Sentinel errors returned by repositories. Adapters translate backend-specific
// failures into these values so the service layer can map them onto gRPC status
// codes without importing GORM or GraphQL packages. Callers should compare with
// errors.Is.
var (
	// ErrNotFound indicates the requested record does not exist.
	ErrNotFound = errors.New("repository: record not found")
	// ErrAlreadyExists indicates a uniqueness conflict on create.
	ErrAlreadyExists = errors.New("repository: record already exists")
	// ErrConflict indicates an optimistic-concurrency (etag) mismatch.
	ErrConflict = errors.New("repository: version conflict")
	// ErrInvalidArgument indicates a caller-supplied argument was rejected (e.g.
	// an order_by field outside the sortable allowlist).
	ErrInvalidArgument = errors.New("repository: invalid argument")
)

// ListParams carries pagination and ordering for List calls. PageToken is an
// opaque cursor produced by a prior List call; an empty token requests the first
// page. OrderBy is an AIP-132 order_by string validated by the adapter against a
// sortable-field allowlist.
type ListParams struct {
	PageSize  int32
	PageToken string
	OrderBy   string
}

// PromoCodeRepository provides CRUD persistence for promo codes. All methods
// accept and return the protobuf PromoCode type; implementations own the
// conversion to and from their storage models, including the normalized Money
// value-objects and the applicable-resources / applicable-offerings join rows.
type PromoCodeRepository interface {
	// Create persists pc and returns the stored record with server-set fields
	// (name, timestamps, etag, state) populated.
	Create(ctx context.Context, pc *promocodepbv1.PromoCode) (*promocodepbv1.PromoCode, error)

	// Get returns the promo code identified by its resource name
	// ("promoCodes/{promo_code}"), or ErrNotFound.
	Get(ctx context.Context, name string) (*promocodepbv1.PromoCode, error)

	// FindByCode returns the promo code with the given human-entered code (e.g.
	// "SUMMER25"), or ErrNotFound. Used by validation/redemption flows that
	// address a code rather than a resource name.
	FindByCode(ctx context.Context, code string) (*promocodepbv1.PromoCode, error)

	// List returns a page of promo codes and an opaque token for the next page
	// (empty when there are no further results).
	List(ctx context.Context, params ListParams) (items []*promocodepbv1.PromoCode, nextPageToken string, err error)

	// Update persists the fields named by paths (an AIP-134 field mask); an empty
	// paths slice means a full replace. The record is identified by pc.Name and
	// pc.Etag guards against concurrent writes (ErrConflict on mismatch).
	Update(ctx context.Context, pc *promocodepbv1.PromoCode, paths []string) (*promocodepbv1.PromoCode, error)

	// Delete removes the promo code identified by its resource name, returning
	// ErrNotFound when it does not exist.
	Delete(ctx context.Context, name string) error
}
