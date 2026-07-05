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
	"github.com/oh-tarnished/freebusy/internal/service/availability/db/hasura"
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
	// to a property or organisation and narrowed by an AIP-160 filter string. Each
	// returned unit is fully enriched (pricing + schedule policy).
	SearchUnits(ctx context.Context, property, organisation, filter string) ([]*engine.UnitInfo, error)

	// ActiveBookingsForUnits batches ActiveBookings over many units, keyed by unit
	// id — used by search to avoid a per-unit round trip.
	ActiveBookingsForUnits(ctx context.Context, unitIDs []string, start, end time.Time) (map[string][]engine.Reservation, error)

	// ClosuresForUnits batches Closures over many units (each expanded in its own
	// timezone, from tzByUnit), keyed by unit id.
	ClosuresForUnits(ctx context.Context, unitIDs []string, tzByUnit map[string]string) (map[string][]engine.Closure, error)
}

// Assert the provider implementations satisfy the contract here.
var (
	_ AvailabilityReader = (*gorm.AvailabilityReader)(nil)
	_ AvailabilityReader = (*hasura.AvailabilityReader)(nil)
)

// New returns the AvailabilityReader for the configured provider, built over the
// matching handle on conn ([database].provider; GORM by default, Hasura opt-in).
func New(conn *database.Connection) AvailabilityReader {
	if database.ProviderFromConfig() == database.ProviderHasura {
		return hasura.NewAvailabilityReader(conn.Hasura)
	}
	return gorm.NewAvailabilityReader(conn.PgSQLConn)
}
