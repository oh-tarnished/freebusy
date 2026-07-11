// Package db is the schedule persistence seam. AvailabilityException CRUD is
// the generated provider-agnostic repositories
// (internal/database/repository/freebusy/schedule — GORM or Hasura behind one
// interface, including the window/date_range span value objects); the Schedule
// itself is an AIP-156 singleton with a normalized section graph the generator
// does not express, so its upsert stays with the hand-written provider
// implementations in the gorm/ and hasura/ sub-packages.
package db

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database"
	schedgen "github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/schedule"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/schedule/db/gorm"
	"github.com/oh-tarnished/freebusy/internal/service/schedule/db/hasura"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/schedule/v1/schedulepbv1"
)

// ScheduleRepository provides persistence for a unit's availability
// configuration (a singleton Schedule) and its AvailabilityException
// resources. Errors are the repox sentinels (aliased in internal/types).
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
	ListAvailabilityExceptions(ctx context.Context, parent string, in repox.ListInput) (items []*schedulepbv1.AvailabilityException, nextPageToken string, err error)

	// DeleteAvailabilityException removes the exception (and its span rows) by
	// resource name.
	DeleteAvailabilityException(ctx context.Context, name string) error
}

// singleton is the hand-written slice of the contract: the Schedule upsert and
// read the providers keep implementing.
type singleton interface {
	GetSchedule(ctx context.Context, name string) (*schedulepbv1.Schedule, error)
	UpdateSchedule(ctx context.Context, s *schedulepbv1.Schedule, paths []string) (*schedulepbv1.Schedule, error)
}

// New returns the ScheduleRepository for the configured provider
// ([database].provider; GORM by default, Hasura opt-in): generated
// repositories for exceptions, the provider's singleton for the Schedule.
func New(conn *database.Connection) ScheduleRepository {
	if database.ProviderFromConfig() == database.ProviderHasura {
		return &repos{
			singleton: hasura.NewScheduleRepository(conn.Hasura),
			gen:       schedgen.New(repox.Conn{GraphQL: conn.Hasura}),
		}
	}
	return &repos{
		singleton: gorm.NewScheduleRepository(conn.PgSQLConn),
		gen:       schedgen.New(repox.Conn{Gorm: conn.PgSQLConn}),
	}
}

// repos joins the hand-written singleton with the generated exception
// repositories.
type repos struct {
	singleton singleton
	gen       schedgen.Repositories
}

func (r *repos) GetSchedule(ctx context.Context, name string) (*schedulepbv1.Schedule, error) {
	return r.singleton.GetSchedule(ctx, name)
}

func (r *repos) UpdateSchedule(ctx context.Context, s *schedulepbv1.Schedule, paths []string) (*schedulepbv1.Schedule, error) {
	return r.singleton.UpdateSchedule(ctx, s, paths)
}

func (r *repos) CreateAvailabilityException(ctx context.Context, parent string, e *schedulepbv1.AvailabilityException) (*schedulepbv1.AvailabilityException, error) {
	return r.gen.AvailabilityExceptions.Create(ctx, parent, e)
}

func (r *repos) GetAvailabilityException(ctx context.Context, name string) (*schedulepbv1.AvailabilityException, error) {
	return r.gen.AvailabilityExceptions.Get(ctx, name)
}

func (r *repos) ListAvailabilityExceptions(ctx context.Context, parent string, in repox.ListInput) ([]*schedulepbv1.AvailabilityException, string, error) {
	return r.gen.AvailabilityExceptions.List(ctx, parent, in)
}

func (r *repos) DeleteAvailabilityException(ctx context.Context, name string) error {
	return r.gen.AvailabilityExceptions.Delete(ctx, name)
}
