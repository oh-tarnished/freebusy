// Package db is the schedule persistence layer. It defines the provider-agnostic
// ScheduleRepository contract (spoken in protobuf domain types) and a factory
// that builds the implementation for the configured backend. Shared,
// provider-neutral vocabulary (errors, list params, names, field masks) lives in
// internal/types.
package db

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/service/schedule/db/gorm"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// ScheduleRepository provides persistence for a unit's availability configuration
// (a singleton Schedule) and its AvailabilityException resources. Errors are the
// sentinels in internal/types.
type ScheduleRepository interface {
	// GetSchedule returns a unit's schedule config ("…/units/{unit}/schedule");
	// an unconfigured unit reports an empty Schedule (name only).
	GetSchedule(ctx context.Context, name string) (*schedulepbv1.Schedule, error)

	// UpdateSchedule upserts the schedule config (created on first update). paths
	// is an AIP-134 field mask over the sections; an empty mask replaces all.
	UpdateSchedule(ctx context.Context, s *schedulepbv1.Schedule, paths []string) (*schedulepbv1.Schedule, error)

	// CreateAvailabilityException persists an exception under parent
	// ("properties/{property}/units/{unit}").
	CreateAvailabilityException(ctx context.Context, parent string, e *schedulepbv1.AvailabilityException) (*schedulepbv1.AvailabilityException, error)

	// GetAvailabilityException returns the exception by resource name.
	GetAvailabilityException(ctx context.Context, name string) (*schedulepbv1.AvailabilityException, error)

	// ListAvailabilityExceptions returns a page of exceptions under parent (a unit).
	ListAvailabilityExceptions(ctx context.Context, parent string, params types.ListParams) (items []*schedulepbv1.AvailabilityException, nextPageToken string, err error)

	// DeleteAvailabilityException removes the exception by resource name.
	DeleteAvailabilityException(ctx context.Context, name string) error
}

// Assert the provider implementation satisfies the contract here.
var _ ScheduleRepository = (*gorm.ScheduleRepository)(nil)

// New returns the ScheduleRepository for the configured provider. GORM is the
// default; the Hasura implementation is a follow-up increment (mirroring the
// property/organisation hasura adapters), so for now every provider resolves to
// the GORM adapter.
func New(conn *database.Connection) ScheduleRepository {
	return gorm.NewScheduleRepository(conn.PgSQLConn)
}
