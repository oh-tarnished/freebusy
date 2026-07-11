package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/postaladdressql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/mediasql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/policiesql"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UpdateProperty applies the masked fields of p and returns the result, as one
// atomic mutation batch: the merged proto is re-materialized into a fresh child
// graph, the property row is repointed at it, and the superseded media / address
// / policy rows are deleted in the same transaction.
func (r *PropertyRepository) UpdateProperty(ctx context.Context, p *propertypbv1.Property, paths []string) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(p.GetName())
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.Properties.Get(ctx, id)
	if err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if p.GetEtag() != "" && res.Etag != nil && p.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}
	parts, old, err := r.fetchPropertyParts(ctx, res)
	if err != nil {
		return nil, err
	}

	merged := propertyFromParts(parts)
	applyPropertyMask(merged, p, paths)
	now := time.Now().UTC()
	g := buildPropertyGraph(merged, now)
	g.property.Id = id

	tx := r.svc.Mutation.Tx()
	if g.address != nil {
		var out postaladdressql.InsertCommonPostalAddressResponse
		tx.Add(r.svc.Mutation.Common.PostalAddress.CreateOp(*g.address, &out))
	}
	if g.policy != nil {
		var out policiesql.InsertPropertyPoliciesResponse
		tx.Add(r.svc.Mutation.Property.Policies.CreateOp(*g.policy, &out))
	}
	patch := propertiesql.UpdateInput{
		Organisation: graphql.Value(g.property.Organisation),
		DisplayName:  graphql.Value(g.property.DisplayName),
		Description:  dbutil.NullableStr(g.property.Description),
		TimeZone:     graphql.Value(g.property.TimeZone),
		Tags:         graphql.Value(g.property.Tags),
		Attributes:   graphql.Value(g.property.Attributes),
		AddressId:    dbutil.NullableStr(g.property.AddressId),
		PolicyId:     dbutil.NullableStr(g.property.PolicyId),
		Etag:         graphql.Value(ulid.GenerateString()),
		UpdateTime:   graphql.Value(dbutil.TsToStr(timestamppb.New(now))),
	}
	var updRes propertiesql.UpdatePropertyPropertiesByIdResponse
	tx.Add(r.svc.Mutation.Property.Properties.UpdateOp(id, patch, &updRes))

	for i := range g.medias {
		g.medias[i].PropertyId = id
		var out mediasql.InsertPropertyMediasResponse
		tx.Add(r.svc.Mutation.Property.Medias.CreateOp(g.medias[i], &out))
	}
	queuePropertyChildDeletes(tx, r, old)

	if err := tx.Commit(ctx); err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	return r.GetProperty(ctx, p.GetName())
}

func (r *PropertyRepository) ArchiveProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	return r.setPropertyState(ctx, name, "ARCHIVED")
}

func (r *PropertyRepository) UnarchiveProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	return r.setPropertyState(ctx, name, "ACTIVE")
}

func (r *PropertyRepository) setPropertyState(ctx context.Context, name, state string) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(name)
	if err != nil {
		return nil, err
	}
	patch := propertiesql.UpdateInput{
		State:      graphql.Value(state),
		Etag:       graphql.Value(ulid.GenerateString()),
		UpdateTime: graphql.Value(dbutil.TsToStr(timestamppb.New(time.Now().UTC()))),
	}
	if _, err := r.svc.Mutation.Property.Properties.Update(ctx, id, patch); err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	return r.GetProperty(ctx, name)
}
