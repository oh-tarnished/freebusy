package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitlicencesql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Licence persistence: PropertyLicence and UnitLicence follow the Unit
// repository pattern — the attachment (scanned certificate) is a belongs-to
// row in shared.attachments, batched with the licence row in one atomic
// mutation document and replaced wholesale on update.

func licencesList() *licencesql.ListRequest         { return licencesql.List() }
func unitLicencesList() *unitlicencesql.ListRequest { return unitlicencesql.List() }

// attachment resolves a licence's attachment row, or nil when it has none.
func (r *PropertyRepository) attachment(ctx context.Context, id *string) (*sharedschema.SharedAttachments, error) {
	if id == nil {
		return nil, nil
	}
	a, err := r.svc.Query.Shared.Attachments.Get(ctx, *id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	return a, nil
}

// --- PropertyLicence -----------------------------------------------------------

func (r *PropertyRepository) CreatePropertyLicence(ctx context.Context, parent string, l *propertypbv1.PropertyLicence) (*propertypbv1.PropertyLicence, error) {
	propertyID, id, name, err := types.ResolvePropertyLicenceName(parent, l.GetName())
	if err != nil {
		return nil, err
	}
	g := buildPropertyLicenceGraph(l, propertyID, time.Now().UTC())
	g.licence.Id = id
	g.licence.Name = name
	g.licence.State = stateActive
	g.licence.Etag = ulid.GenerateString()

	tx := r.svc.Mutation.Tx()
	if g.attachment != nil {
		var res sharedschema.InsertSharedAttachmentsResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.CreateOp(*g.attachment, &res))
	}
	var licRes pschema.InsertPropertyLicencesResponse
	tx.Add(r.svc.Mutation.Property.Licences.CreateOp(g.licence, &licRes))
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetPropertyLicence(ctx, name)
}

func (r *PropertyRepository) GetPropertyLicence(ctx context.Context, name string) (*propertypbv1.PropertyLicence, error) {
	id, err := types.PropertyLicenceID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.Licences.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	att, err := r.attachment(ctx, res.AttachmentId)
	if err != nil {
		return nil, err
	}
	return propertyLicenceFromParts(res, att), nil
}

func (r *PropertyRepository) ListPropertyLicences(ctx context.Context, parent string, params types.ListParams) ([]*propertypbv1.PropertyLicence, string, error) {
	propertyID, err := types.PropertyID(parent)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[pschema.PropertyLicences](property.PropertyLicenceFilterSpec, r.svc.Query.Property.Licences).
		Scope(licencesql.PropertyId.Eq(propertyID)).List(ctx, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.PropertyLicence, 0, len(rows))
	for i := range rows {
		att, err := r.attachment(ctx, rows[i].AttachmentId)
		if err != nil {
			return nil, "", err
		}
		items = append(items, propertyLicenceFromParts(&rows[i], att))
	}
	return items, next, nil
}

func (r *PropertyRepository) UpdatePropertyLicence(ctx context.Context, l *propertypbv1.PropertyLicence, paths []string) (*propertypbv1.PropertyLicence, error) {
	id, err := types.PropertyLicenceID(l.GetName())
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.Licences.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if l.GetEtag() != "" && res.Etag != nil && l.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}
	att, err := r.attachment(ctx, res.AttachmentId)
	if err != nil {
		return nil, err
	}

	merged := propertyLicenceFromParts(res, att)
	applyPropertyLicenceMask(merged, l, paths)
	now := time.Now().UTC()
	g := buildPropertyLicenceGraph(merged, res.PropertyId, now)

	tx := r.svc.Mutation.Tx()
	if g.attachment != nil {
		var out sharedschema.InsertSharedAttachmentsResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.CreateOp(*g.attachment, &out))
	}
	patch := licencesql.UpdateInput{
		Type:             graphql.Value(g.licence.Type),
		LicenceNumber:    nullableStr(g.licence.LicenceNumber),
		IssuingAuthority: nullableStr(g.licence.IssuingAuthority),
		IssueDate:        nullableStr(g.licence.IssueDate),
		ExpiryDate:       nullableStr(g.licence.ExpiryDate),
		Notes:            nullableStr(g.licence.Notes),
		AttachmentId:     nullableStr(g.licence.AttachmentId),
		Etag:             graphql.Value(ulid.GenerateString()),
		UpdateTime:       graphql.Value(tsToStr(timestamppb.New(now))),
	}
	var updRes pschema.UpdatePropertyLicencesByIdResponse
	tx.Add(r.svc.Mutation.Property.Licences.UpdateOp(id, patch, &updRes))
	if res.AttachmentId != nil {
		var out sharedschema.DeleteSharedAttachmentsByIdResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.DeleteOp(*res.AttachmentId, &out))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetPropertyLicence(ctx, l.GetName())
}

func (r *PropertyRepository) DeletePropertyLicence(ctx context.Context, name string) error {
	id, err := types.PropertyLicenceID(name)
	if err != nil {
		return err
	}
	res, err := r.svc.Query.Property.Licences.Get(ctx, id)
	if err != nil {
		return mapHasuraErr(err)
	}
	if res == nil {
		return types.ErrNotFound
	}
	tx := r.svc.Mutation.Tx()
	var delRes pschema.DeletePropertyLicencesByIdResponse
	tx.Add(r.svc.Mutation.Property.Licences.DeleteOp(id, &delRes))
	if res.AttachmentId != nil {
		var out sharedschema.DeleteSharedAttachmentsByIdResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.DeleteOp(*res.AttachmentId, &out))
	}
	return mapHasuraErr(tx.Commit(ctx))
}

// --- UnitLicence ---------------------------------------------------------------

func (r *PropertyRepository) CreateUnitLicence(ctx context.Context, parent string, l *propertypbv1.UnitLicence) (*propertypbv1.UnitLicence, error) {
	propertyID, unitID, id, name, err := types.ResolveUnitLicenceName(parent, l.GetName())
	if err != nil {
		return nil, err
	}
	g := buildUnitLicenceGraph(l, propertyID, unitID, time.Now().UTC())
	g.licence.Id = id
	g.licence.Name = name
	g.licence.State = stateActive
	g.licence.Etag = ulid.GenerateString()

	tx := r.svc.Mutation.Tx()
	if g.attachment != nil {
		var res sharedschema.InsertSharedAttachmentsResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.CreateOp(*g.attachment, &res))
	}
	var licRes pschema.InsertPropertyUnitLicencesResponse
	tx.Add(r.svc.Mutation.Property.UnitLicences.CreateOp(g.licence, &licRes))
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetUnitLicence(ctx, name)
}

func (r *PropertyRepository) GetUnitLicence(ctx context.Context, name string) (*propertypbv1.UnitLicence, error) {
	id, err := types.UnitLicenceID(name)
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.UnitLicences.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	att, err := r.attachment(ctx, res.AttachmentId)
	if err != nil {
		return nil, err
	}
	return unitLicenceFromParts(res, att), nil
}

func (r *PropertyRepository) ListUnitLicences(ctx context.Context, parent string, params types.ListParams) ([]*propertypbv1.UnitLicence, string, error) {
	_, unitID, err := types.ParseUnitParent(parent)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[pschema.PropertyUnitLicences](property.UnitLicenceFilterSpec, r.svc.Query.Property.UnitLicences).
		Scope(unitlicencesql.UnitId.Eq(unitID)).List(ctx, types.FilterxInput(params))
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.UnitLicence, 0, len(rows))
	for i := range rows {
		att, err := r.attachment(ctx, rows[i].AttachmentId)
		if err != nil {
			return nil, "", err
		}
		items = append(items, unitLicenceFromParts(&rows[i], att))
	}
	return items, next, nil
}

func (r *PropertyRepository) UpdateUnitLicence(ctx context.Context, l *propertypbv1.UnitLicence, paths []string) (*propertypbv1.UnitLicence, error) {
	id, err := types.UnitLicenceID(l.GetName())
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.UnitLicences.Get(ctx, id)
	if err != nil {
		return nil, mapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if l.GetEtag() != "" && res.Etag != nil && l.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}
	att, err := r.attachment(ctx, res.AttachmentId)
	if err != nil {
		return nil, err
	}

	merged := unitLicenceFromParts(res, att)
	applyUnitLicenceMask(merged, l, paths)
	now := time.Now().UTC()
	g := buildUnitLicenceGraph(merged, res.PropertyId, res.UnitId, now)

	tx := r.svc.Mutation.Tx()
	if g.attachment != nil {
		var out sharedschema.InsertSharedAttachmentsResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.CreateOp(*g.attachment, &out))
	}
	patch := unitlicencesql.UpdateInput{
		Type:             graphql.Value(g.licence.Type),
		LicenceNumber:    nullableStr(g.licence.LicenceNumber),
		IssuingAuthority: nullableStr(g.licence.IssuingAuthority),
		IssueDate:        nullableStr(g.licence.IssueDate),
		ExpiryDate:       nullableStr(g.licence.ExpiryDate),
		Notes:            nullableStr(g.licence.Notes),
		AttachmentId:     nullableStr(g.licence.AttachmentId),
		Etag:             graphql.Value(ulid.GenerateString()),
		UpdateTime:       graphql.Value(tsToStr(timestamppb.New(now))),
	}
	var updRes pschema.UpdatePropertyUnitLicencesByIdResponse
	tx.Add(r.svc.Mutation.Property.UnitLicences.UpdateOp(id, patch, &updRes))
	if res.AttachmentId != nil {
		var out sharedschema.DeleteSharedAttachmentsByIdResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.DeleteOp(*res.AttachmentId, &out))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, mapHasuraErr(err)
	}
	return r.GetUnitLicence(ctx, l.GetName())
}

func (r *PropertyRepository) DeleteUnitLicence(ctx context.Context, name string) error {
	id, err := types.UnitLicenceID(name)
	if err != nil {
		return err
	}
	res, err := r.svc.Query.Property.UnitLicences.Get(ctx, id)
	if err != nil {
		return mapHasuraErr(err)
	}
	if res == nil {
		return types.ErrNotFound
	}
	tx := r.svc.Mutation.Tx()
	var delRes pschema.DeletePropertyUnitLicencesByIdResponse
	tx.Add(r.svc.Mutation.Property.UnitLicences.DeleteOp(id, &delRes))
	if res.AttachmentId != nil {
		var out sharedschema.DeleteSharedAttachmentsByIdResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.DeleteOp(*res.AttachmentId, &out))
	}
	return mapHasuraErr(tx.Commit(ctx))
}
