// Package db is the availability read layer. AvailabilityService is a read-only
// compute engine, so this package exposes a narrow AvailabilityReader port — the
// exact reads the engine needs (unit config, its active bookings, its closures,
// and a catalog sweep for search) — plus a factory that builds the implementation
// for the configured backend. The read results are the provider-neutral value
// types the pure engine (internal/service/availability/engine) consumes; no
// protobuf or GORM leaks across this boundary.
package db

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/service/availability/db/gorm"
	"github.com/oh-tarnished/freebusy/internal/service/availability/engine"
)

// AvailabilityReader is the read-only port the availability engine runs on. It
// speaks in the neutral engine value types.
type AvailabilityReader interface {
	// GetUnit returns the unit's config and policy by resource name, or
	// types.ErrNotFound.
	GetUnit(ctx context.Context, unitName string) (*engine.UnitInfo, error)

	// ActiveBookings returns the held/confirmed bookings on unitID whose window
	// overlaps [start,end).
	ActiveBookings(ctx context.Context, unitID string, start, end time.Time) ([]engine.Reservation, error)

	// Closures returns the unit's CLOSURE exceptions as UTC spans. Date-range
	// closures are expanded to instants in tz (the unit's IANA timezone).
	Closures(ctx context.Context, unitID, tz string) ([]engine.Closure, error)

	// SearchUnits returns active units for the storefront search, optionally scoped
	// to a property or organisation and narrowed by an AIP-160 filter string.
	SearchUnits(ctx context.Context, property, organisation, filter string) ([]*engine.UnitInfo, error)
}

// Assert the provider implementation satisfies the contract here.
var _ AvailabilityReader = (*gorm.AvailabilityReader)(nil)

// New returns the AvailabilityReader for the configured provider. GORM is the
// default; a Hasura reader is a follow-up increment (the engine is provider-
// neutral), so for now every provider resolves to the GORM reader.
func New(conn *database.Connection) AvailabilityReader {
	return gorm.NewAvailabilityReader(conn.PgSQLConn)
}
