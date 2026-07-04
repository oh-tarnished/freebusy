// Package hasura provides the Hasura/GraphQL-backed implementation of the
// organisation persistence contract. Organisation and Member are flat rows, so
// each write is a single mutation.
package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/membersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
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
	order, err := orgOrderTerms(params.OrderBy)
	if err != nil {
		return nil, "", err
	}
	where, hasWhere, err := orgFilterPredicate(params.Filter)
	if err != nil {
		return nil, "", err
	}
	limit, offset := types.PageBounds(params)
	req := resourceql.List().Limit(limit + 1).Offset(offset)
	if len(order) > 0 {
		req = req.OrderBy(order...)
	}
	if hasWhere {
		req = req.Where(where)
	}
	rows, err := r.svc.Query.Organisation.Resource.List(ctx, req)
	if err != nil {
		return nil, "", mapHasuraErr(err)
	}
	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		next = types.EncodeOffset(offset + limit)
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
	order, err := memberOrderTerms(params.OrderBy)
	if err != nil {
		return nil, "", err
	}
	where, err := memberFilterPredicate(params.Filter, orgID)
	if err != nil {
		return nil, "", err
	}
	limit, offset := types.PageBounds(params)
	req := membersql.List().Limit(limit + 1).Offset(offset).Where(where)
	if len(order) > 0 {
		req = req.OrderBy(order...)
	}
	rows, err := r.svc.Query.Organisation.Members.List(ctx, req)
	if err != nil {
		return nil, "", mapHasuraErr(err)
	}
	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		next = types.EncodeOffset(offset + limit)
	}
	items := make([]*orgpbv1.Member, 0, len(rows))
	for i := range rows {
		items = append(items, memberFromModel(&rows[i]))
	}
	return items, next, nil
}
