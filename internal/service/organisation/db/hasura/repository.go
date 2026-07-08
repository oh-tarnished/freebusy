// Package hasura provides the Hasura/GraphQL-backed implementation of the
// organisation persistence contract. Organisation and Member are flat rows, so
// each write is a single mutation.
package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/membersql"
	orgschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// OrganisationRepository is the Hasura-backed organisation + member repository.
type OrganisationRepository struct {
	svc *freebusyql.Service
}

// NewOrganisationRepository returns a Hasura-backed OrganisationRepository.
func NewOrganisationRepository(svc *freebusyql.Service) *OrganisationRepository {
	return &OrganisationRepository{svc: svc}
}

// --- Organisation ------------------------------------------------------------

func (r *OrganisationRepository) CreateOrganisation(ctx context.Context, o *orgpbv1.Organisation) (*orgpbv1.Organisation, error) {
	id, name, err := types.ResolveOrganisationName(o.GetName())
	if err != nil {
		return nil, err
	}
	in := orgToCreateInput(o, time.Now().UTC())
	in.Id = id
	in.Name = name
	in.Etag = ulid.GenerateString()
	if _, err := r.svc.Mutation.Organisation.Resource.Create(ctx, in); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetOrganisation(ctx, name)
}

func (r *OrganisationRepository) GetOrganisation(ctx context.Context, name string) (*orgpbv1.Organisation, error) {
	id, err := types.OrganisationID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Organisation.Resource.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return orgFromModel(res), nil
}

func (r *OrganisationRepository) ListOrganisations(ctx context.Context, params types.ListParams) ([]*orgpbv1.Organisation, string, error) {
	rows, next, err := filterx.Hasura[orgschema.OrganisationResource](organisation.OrganisationFilterSpec, r.svc.Query.Organisation.Resource).
		List(ctx, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*orgpbv1.Organisation, 0, len(rows))
	for i := range rows {
		items = append(items, orgFromModel(&rows[i]))
	}
	return items, next, nil
}

// --- Member ------------------------------------------------------------------

func (r *OrganisationRepository) CreateMember(ctx context.Context, parent string, mem *orgpbv1.Member) (*orgpbv1.Member, error) {
	orgID, id, name, err := types.ResolveMemberName(parent, mem.GetName())
	if err != nil {
		return nil, err
	}
	in := memberToCreateInput(mem, time.Now().UTC())
	in.Id = id
	in.Name = name
	in.OrganisationId = orgID
	in.Etag = ulid.GenerateString()
	if _, err := r.svc.Mutation.Organisation.Members.Create(ctx, in); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetMember(ctx, name)
}

func (r *OrganisationRepository) GetMember(ctx context.Context, name string) (*orgpbv1.Member, error) {
	id, err := types.MemberID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Organisation.Members.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return memberFromModel(res), nil
}

func (r *OrganisationRepository) ListMembers(ctx context.Context, parent string, params types.ListParams) ([]*orgpbv1.Member, string, error) {
	orgID, err := types.OrganisationID(parent)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[orgschema.OrganisationMembers](organisation.MemberFilterSpec, r.svc.Query.Organisation.Members).
		Scope(membersql.OrganisationId.Eq(orgID)).
		List(ctx, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*orgpbv1.Member, 0, len(rows))
	for i := range rows {
		items = append(items, memberFromModel(&rows[i]))
	}
	return items, next, nil
}
