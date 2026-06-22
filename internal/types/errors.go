package types

import "errors"

// Sentinel errors returned by repositories. Adapters translate backend-specific
// failures into these values so the service layer can map them onto gRPC status
// codes without importing GORM or GraphQL packages. Compare with errors.Is.
var (
	// ErrNotFound indicates the requested record does not exist.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists indicates a uniqueness conflict on create.
	ErrAlreadyExists = errors.New("already exists")
	// ErrConflict indicates an optimistic-concurrency (etag) mismatch.
	ErrConflict = errors.New("version conflict")
	// ErrInvalidArgument indicates a caller-supplied argument was rejected (e.g.
	// an order_by field outside the sortable allowlist).
	ErrInvalidArgument = errors.New("invalid argument")
)
