package hasura

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/membersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *OrganisationRepository) UpdateOrganisation(ctx context.Context, o *orgpbv1.Organisation, paths []string) (*orgpbv1.Organisation, error) {
	id, err := types.OrganisationID(o.GetName())
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
	if o.GetEtag() != "" && res.Etag != nil && o.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}
	merged := orgFromModel(res)
	applyOrgMask(merged, o, paths)
	patch := resourceql.UpdateInput{
		DisplayName:  graphql.Value(merged.GetDisplayName()),
		Slug:         nullableStr(merged.GetSlug()),
		BillingEmail: nullableStr(merged.GetBillingEmail()),
		Settings:     nullableJSON(structToJSON(merged.GetSettings())),
		Etag:         graphql.Value(ulid.GenerateString()),
		UpdateTime:   graphql.Value(tsToStr(timestamppb.New(time.Now().UTC()))),
	}
	if _, err := r.svc.Mutation.Organisation.Resource.Update(ctx, id, patch); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetOrganisation(ctx, o.GetName())
}

// DeleteOrganisation removes an organisation. Without force it fails when the
// organisation still has members.
func (r *OrganisationRepository) DeleteOrganisation(ctx context.Context, name string, force bool) error {
	id, err := types.OrganisationID(name)
	if err != nil {
		return err
	}
	res, err := r.svc.Query.Organisation.Resource.Get(ctx, id)
	if err != nil {
		return mapHasuraErr(err)
	}
	if res == nil {
		return types.ErrNotFound
	}
	if !force {
		members, err := r.svc.Query.Organisation.Members.List(ctx, membersql.List().Where(membersql.OrganisationId.Eq(id)).Limit(1))
		if err != nil {
			return mapHasuraErr(err)
		}
		if len(members) > 0 {
			return types.ErrConflict
		}
	}
	if _, err := r.svc.Mutation.Organisation.Resource.Delete(ctx, id); err != nil {
		return mapHasuraErr(err)
	}
	return nil
}

func (r *OrganisationRepository) UpdateMember(ctx context.Context, mem *orgpbv1.Member, paths []string) (*orgpbv1.Member, error) {
	id, err := types.MemberID(mem.GetName())
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
	if mem.GetEtag() != "" && res.Etag != nil && mem.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}
	patch := membersql.UpdateInput{
		Etag:       graphql.Value(ulid.GenerateString()),
		UpdateTime: graphql.Value(tsToStr(timestamppb.New(time.Now().UTC()))),
	}
	if types.InMask(paths, "role") {
		patch.Role = graphql.Value(roleToStr(mem.GetRole()))
	}
	if _, err := r.svc.Mutation.Organisation.Members.Update(ctx, id, patch); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetMember(ctx, mem.GetName())
}

func (r *OrganisationRepository) DeleteMember(ctx context.Context, name string) error {
	id, err := types.MemberID(name)
	if err != nil {
		return err
	}
	if _, err := r.svc.Mutation.Organisation.Members.Delete(ctx, id); err != nil {
		return mapHasuraErr(err)
	}
	return nil
}

func nullableStr(s string) graphql.Nullable[string] {
	if s == "" {
		return graphql.Null[string]()
	}
	return graphql.Value(s)
}

func nullableJSON(b []byte) graphql.Nullable[json.RawMessage] {
	if len(b) == 0 {
		return graphql.Null[json.RawMessage]()
	}
	return graphql.Value(json.RawMessage(b))
}

// mapHasuraErr translates GraphQL/runtime errors into the repository sentinels.
func mapHasuraErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, graphql.ErrConflict):
		return types.ErrConflict
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") {
		return types.ErrAlreadyExists
	}
	return err
}
