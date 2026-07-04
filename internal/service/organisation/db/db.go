// Package db is the organisation persistence layer. It defines the
// provider-agnostic OrganisationRepository contract (spoken in protobuf domain
// types) and a factory that builds the implementation for the configured backend
// (GORM by default, Hasura opt-in). Shared, provider-neutral vocabulary (errors,
// list params, names, field masks) lives in internal/types.
package db

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database"
	"github.com/oh-tarnished/freebusy/internal/service/organisation/db/gorm"
	"github.com/oh-tarnished/freebusy/internal/service/organisation/db/hasura"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
)

// OrganisationRepository provides CRUD persistence for organisations (chains) and
// their members. All methods accept and return the protobuf domain types; errors
// are the sentinels in internal/types.
type OrganisationRepository interface {
	CreateOrganisation(ctx context.Context, o *orgpbv1.Organisation) (*orgpbv1.Organisation, error)
	GetOrganisation(ctx context.Context, name string) (*orgpbv1.Organisation, error)
	ListOrganisations(ctx context.Context, params types.ListParams) (items []*orgpbv1.Organisation, nextPageToken string, err error)
	UpdateOrganisation(ctx context.Context, o *orgpbv1.Organisation, paths []string) (*orgpbv1.Organisation, error)
	DeleteOrganisation(ctx context.Context, name string, force bool) error

	// CreateMember persists a member under parent ("organisations/{organisation}");
	// InviteMember maps onto it (a member starts in INVITED state).
	CreateMember(ctx context.Context, parent string, m *orgpbv1.Member) (*orgpbv1.Member, error)
	GetMember(ctx context.Context, name string) (*orgpbv1.Member, error)
	ListMembers(ctx context.Context, parent string, params types.ListParams) (items []*orgpbv1.Member, nextPageToken string, err error)
	UpdateMember(ctx context.Context, m *orgpbv1.Member, paths []string) (*orgpbv1.Member, error)
	DeleteMember(ctx context.Context, name string) error
}

// Assert the provider implementations satisfy the contract here.
var (
	_ OrganisationRepository = (*gorm.OrganisationRepository)(nil)
	_ OrganisationRepository = (*hasura.OrganisationRepository)(nil)
)

// New returns the OrganisationRepository for the configured provider ([database].
// provider; GORM by default, Hasura opt-in).
func New(conn *database.Connection) OrganisationRepository {
	if database.ProviderFromConfig() == database.ProviderHasura {
		return hasura.NewOrganisationRepository(conn.Hasura)
	}
	return gorm.NewOrganisationRepository(conn.PgSQLConn)
}
