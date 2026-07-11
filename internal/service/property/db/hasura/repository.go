// Package hasura provides the Hasura/GraphQL-backed implementation of the
// property persistence contract (internal/service/property/db.PropertyRepository).
// It adapts the generated freebusyql handlers to that contract, converting between
// protobuf domain types and the normalized GraphQL schema (the address, policy,
// media, and a unit's pricing child tables, plus the common Money/PostalAddress
// and shared DateRange value-objects).
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/mediasql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// PropertyRepository is the Hasura-backed property repository. Hasura exposes no
// client-side transactions across calls, but its mutation API can run several
// mutations as one atomic GraphQL document (svc.Mutation.Tx()); writes use that
// so a property/unit and its child rows commit together or not at all. Each read
// hydrates the row's children with follow-up queries.
type PropertyRepository struct {
	svc *freebusyql.Service
}

// NewPropertyRepository returns a Hasura-backed PropertyRepository bound to svc.
func NewPropertyRepository(svc *freebusyql.Service) *PropertyRepository {
	return &PropertyRepository{svc: svc}
}

// --- Property ----------------------------------------------------------------

func (r *PropertyRepository) CreateProperty(ctx context.Context, p *propertypbv1.Property) (*propertypbv1.Property, error) {
	id, name, err := types.ResolvePropertyName(p.GetName())
	if err != nil {
		return nil, err
	}
	g := buildPropertyGraph(p, time.Now().UTC())
	g.property.Id = id
	g.property.Name = name
	g.property.Etag = ulid.GenerateString()

	tx := r.svc.Mutation.Tx()
	if g.address != nil {
		var res commonschema.InsertCommonPostalAddressResponse
		tx.Add(r.svc.Mutation.Common.PostalAddress.CreateOp(*g.address, &res))
	}
	if g.policy != nil {
		var res pschema.InsertPropertyPoliciesResponse
		tx.Add(r.svc.Mutation.Property.Policies.CreateOp(*g.policy, &res))
	}
	var propRes pschema.InsertPropertyPropertiesResponse
	tx.Add(r.svc.Mutation.Property.Properties.CreateOp(g.property, &propRes))
	mediaRes := make([]pschema.InsertPropertyMediasResponse, len(g.medias))
	for i := range g.medias {
		g.medias[i].PropertyId = id
		tx.Add(r.svc.Mutation.Property.Medias.CreateOp(g.medias[i], &mediaRes[i]))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	return r.GetProperty(ctx, name)
}

func (r *PropertyRepository) GetProperty(ctx context.Context, name string) (*propertypbv1.Property, error) {
	id, err := types.PropertyID(name)
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
	parts, _, err := r.fetchPropertyParts(ctx, res)
	if err != nil {
		return nil, err
	}
	return propertyFromParts(parts), nil
}

func (r *PropertyRepository) ListProperties(ctx context.Context, in repox.ListInput) ([]*propertypbv1.Property, string, error) {
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura(property.PropertyFilterSpec, r.svc.Query.Property.Properties).
		List(ctx, fin)
	if err != nil {
		return nil, "", dbutil.MapHasuraErr(repox.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Property, 0, len(rows))
	for i := range rows {
		parts, _, err := r.fetchPropertyParts(ctx, &rows[i])
		if err != nil {
			return nil, "", err
		}
		items = append(items, propertyFromParts(parts))
	}
	return items, next, nil
}

// fetchPropertyParts loads a property row's address, policy, media, and child
// unit names, and returns the ids of the deletable child rows.
func (r *PropertyRepository) fetchPropertyParts(ctx context.Context, res *pschema.PropertyProperties) (propertyParts, propertyRefs, error) {
	p := propertyParts{res: res}
	refs := propertyRefs{addressID: res.AddressId, policyID: res.PolicyId}

	if res.AddressId != nil {
		a, err := r.svc.Query.Common.PostalAddress.Get(ctx, *res.AddressId)
		if err != nil {
			return propertyParts{}, propertyRefs{}, dbutil.MapHasuraErr(err)
		}
		p.address = a
	}
	if res.PolicyId != nil {
		pol, err := r.svc.Query.Property.Policies.Get(ctx, *res.PolicyId)
		if err != nil {
			return propertyParts{}, propertyRefs{}, dbutil.MapHasuraErr(err)
		}
		p.policy = pol
	}
	medias, err := r.svc.Query.Property.Medias.List(ctx, mediasList().Where(mediasql.PropertyId.Eq(res.Id)))
	if err != nil {
		return propertyParts{}, propertyRefs{}, dbutil.MapHasuraErr(err)
	}
	p.medias = medias
	for i := range medias {
		refs.mediaIDs = append(refs.mediaIDs, medias[i].Id)
	}
	units, err := r.svc.Query.Property.Units.List(ctx, unitsList().Where(unitsql.PropertyId.Eq(res.Id)))
	if err != nil {
		return propertyParts{}, propertyRefs{}, dbutil.MapHasuraErr(err)
	}
	for i := range units {
		p.unitNames = append(p.unitNames, units[i].Name)
	}
	licences, err := r.svc.Query.Property.Licences.List(ctx, licencesql.List().Where(licencesql.PropertyId.Eq(res.Id)))
	if err != nil {
		return propertyParts{}, propertyRefs{}, dbutil.MapHasuraErr(err)
	}
	for i := range licences {
		p.licenceNames = append(p.licenceNames, licences[i].Name)
	}
	return p, refs, nil
}

// --- Unit --------------------------------------------------------------------
