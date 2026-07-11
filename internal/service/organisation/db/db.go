// Package db is the organisation persistence seam. The CRUD surface is the
// generated provider-agnostic repositories
// (internal/database/repository/freebusy/organisation — GORM or Hasura behind
// one interface); this package narrows them to the service-facing contract and
// layers the one behavior the generator does not know about: the force-delete
// member guard.
package db

import (
	"context"
	"fmt"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/database/repository/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
)

// OrganisationRepository provides CRUD persistence for organisations (chains)
// and their members. All methods accept and return the protobuf domain types;
// errors are the repox sentinels (aliased in internal/types).
type OrganisationRepository interface {
	CreateOrganisation(ctx context.Context, o *orgpbv1.Organisation) (*orgpbv1.Organisation, error)
	GetOrganisation(ctx context.Context, name string) (*orgpbv1.Organisation, error)
	ListOrganisations(ctx context.Context, in repox.ListInput) (items []*orgpbv1.Organisation, nextPageToken string, err error)
	UpdateOrganisation(ctx context.Context, o *orgpbv1.Organisation, paths []string) (*orgpbv1.Organisation, error)
	DeleteOrganisation(ctx context.Context, name string, force bool) error

	// CreateMember persists a member under parent ("organisations/{organisation}");
	// InviteMember maps onto it (a member starts in INVITED state, the column's
	// database default).
	CreateMember(ctx context.Context, parent string, m *orgpbv1.Member) (*orgpbv1.Member, error)
	GetMember(ctx context.Context, name string) (*orgpbv1.Member, error)
	ListMembers(ctx context.Context, parent string, in repox.ListInput) (items []*orgpbv1.Member, nextPageToken string, err error)
	UpdateMember(ctx context.Context, m *orgpbv1.Member, paths []string) (*orgpbv1.Member, error)
	DeleteMember(ctx context.Context, name string) error
}

// New returns the OrganisationRepository for the configured provider
// ([database].provider; GORM by default, Hasura opt-in), built on the generated
// repositories.
func New(conn *database.Connection) OrganisationRepository {
	c := repox.Conn{Gorm: conn.PgSQLConn}
	if database.ProviderFromConfig() == database.ProviderHasura {
		c = repox.Conn{GraphQL: conn.Hasura}
	}
	return &repos{gen: organisation.New(c)}
}

// repos maps the service contract onto the generated repository set.
type repos struct {
	gen organisation.Repositories
}

// --- Organisation ------------------------------------------------------------

func (r *repos) CreateOrganisation(ctx context.Context, o *orgpbv1.Organisation) (*orgpbv1.Organisation, error) {
	return r.gen.Organisations.Create(ctx, o)
}

func (r *repos) GetOrganisation(ctx context.Context, name string) (*orgpbv1.Organisation, error) {
	return r.gen.Organisations.Get(ctx, name)
}

func (r *repos) ListOrganisations(ctx context.Context, in repox.ListInput) ([]*orgpbv1.Organisation, string, error) {
	return r.gen.Organisations.List(ctx, in)
}

func (r *repos) UpdateOrganisation(ctx context.Context, o *orgpbv1.Organisation, paths []string) (*orgpbv1.Organisation, error) {
	return r.gen.Organisations.Update(ctx, o, paths)
}

// DeleteOrganisation removes an organisation. Without force it fails with a
// conflict while the organisation still has members; with force the members
// cascade in the DB.
func (r *repos) DeleteOrganisation(ctx context.Context, name string, force bool) error {
	if !force {
		members, _, err := r.gen.Members.List(ctx, name, repox.ListInput{PageSize: 1})
		if err != nil {
			return err
		}
		if len(members) > 0 {
			return fmt.Errorf("%w: organisation still has members (set force to cascade)", repox.ErrConflict)
		}
	}
	return r.gen.Organisations.Delete(ctx, name)
}

// --- Member ------------------------------------------------------------------

func (r *repos) CreateMember(ctx context.Context, parent string, m *orgpbv1.Member) (*orgpbv1.Member, error) {
	return r.gen.Members.Create(ctx, parent, m)
}

func (r *repos) GetMember(ctx context.Context, name string) (*orgpbv1.Member, error) {
	return r.gen.Members.Get(ctx, name)
}

func (r *repos) ListMembers(ctx context.Context, parent string, in repox.ListInput) ([]*orgpbv1.Member, string, error) {
	return r.gen.Members.List(ctx, parent, in)
}

func (r *repos) UpdateMember(ctx context.Context, m *orgpbv1.Member, paths []string) (*orgpbv1.Member, error) {
	return r.gen.Members.Update(ctx, m, paths)
}

func (r *repos) DeleteMember(ctx context.Context, name string) error {
	return r.gen.Members.Delete(ctx, name)
}
