// Package gorm provides the GORM-backed implementation of the organisation
// persistence contract (internal/service/organisation/db.OrganisationRepository).
// Organisation and Member are flat single-table rows, so writes are direct store
// calls rather than a transactional child graph.
package gorm

import (
	"context"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/organisation"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"gorm.io/gorm"
)

// OrganisationRepository is the GORM-backed organisation + member repository.
type OrganisationRepository struct {
	db *gorm.DB
}

// NewOrganisationRepository returns a GORM-backed OrganisationRepository bound to
// db. The parent db package asserts it satisfies db.OrganisationRepository.
func NewOrganisationRepository(db *gorm.DB) *OrganisationRepository {
	return &OrganisationRepository{db: db}
}

// --- Organisation ------------------------------------------------------------

func (r *OrganisationRepository) CreateOrganisation(ctx context.Context, o *orgpbv1.Organisation) (*orgpbv1.Organisation, error) {
	id, name, err := types.ResolveOrganisationName(o.GetName())
	if err != nil {
		return nil, err
	}
	m := orgToModel(o)
	m.ID = id
	m.Name = name
	m.Etag = ptr(ulid.GenerateString())
	if err := organisation.NewOrganisationStore(r.db).Create(ctx, m); err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetOrganisation(ctx, name)
}

func (r *OrganisationRepository) GetOrganisation(ctx context.Context, name string) (*orgpbv1.Organisation, error) {
	id, err := types.OrganisationID(name)
	if err != nil {
		return nil, err
	}
	var m organisation.Organisation
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, mapGormErr(err)
	}
	return orgFromModel(&m), nil
}

func (r *OrganisationRepository) ListOrganisations(ctx context.Context, params types.ListParams) ([]*orgpbv1.Organisation, string, error) {
	models, next, err := filterx.Gorm[organisation.Organisation](organisation.OrganisationFilterSpec).
		List(ctx, r.db, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapGormErr(types.MapFilterxErr(err))
	}
	items := make([]*orgpbv1.Organisation, 0, len(models))
	for i := range models {
		items = append(items, orgFromModel(&models[i]))
	}
	return items, next, nil
}

// --- Member ------------------------------------------------------------------

func (r *OrganisationRepository) CreateMember(ctx context.Context, parent string, mem *orgpbv1.Member) (*orgpbv1.Member, error) {
	orgID, id, name, err := types.ResolveMemberName(parent, mem.GetName())
	if err != nil {
		return nil, err
	}
	m := memberToModel(mem)
	m.ID = id
	m.Name = name
	m.OrganisationID = orgID
	m.Etag = ptr(ulid.GenerateString())
	if err := organisation.NewMemberStore(r.db).Create(ctx, m); err != nil {
		return nil, mapGormErr(err)
	}
	return r.GetMember(ctx, name)
}

func (r *OrganisationRepository) GetMember(ctx context.Context, name string) (*orgpbv1.Member, error) {
	id, err := types.MemberID(name)
	if err != nil {
		return nil, err
	}
	var m organisation.Member
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, mapGormErr(err)
	}
	return memberFromModel(&m), nil
}

func (r *OrganisationRepository) ListMembers(ctx context.Context, parent string, params types.ListParams) ([]*orgpbv1.Member, string, error) {
	orgID, err := types.OrganisationID(parent)
	if err != nil {
		return nil, "", err
	}
	models, next, err := filterx.Gorm[organisation.Member](organisation.MemberFilterSpec).
		List(ctx, r.db.Where("organisation_id = ?", orgID), types.FilterxInput(params))
	if err != nil {
		return nil, "", mapGormErr(types.MapFilterxErr(err))
	}
	items := make([]*orgpbv1.Member, 0, len(models))
	for i := range models {
		items = append(items, memberFromModel(&models[i]))
	}
	return items, next, nil
}
