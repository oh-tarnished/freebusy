package hasura

import (
	"context"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/filterx"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Licence persistence follows the Unit repository pattern — the attachment
// (scanned certificate) is a belongs-to row in shared.attachments, batched
// with the licence row in one atomic mutation document and replaced wholesale
// on update. One table holds property-wide and per-unit licences alike; target
// and the nullable unit reference say which is which.

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

// CreateLicence persists l under parent ("properties/{property}") and returns
// the stored record. The caller (runtime layer) has already validated that a
// unit reference, if set, belongs to the parent property.
func (r *PropertyRepository) CreateLicence(ctx context.Context, parent string, l *propertypbv1.Licence) (*propertypbv1.Licence, error) {
	propertyID, id, name, err := types.ResolveLicenceName(parent, l.GetName())
	if err != nil {
		return nil, err
	}
	var unitID *string
	if l.GetUnit() != "" {
		_, uid, err := types.ParseUnitParent(l.GetUnit())
		if err != nil {
			return nil, err
		}
		unitID = &uid
	}
	g := buildLicenceGraph(l, propertyID, unitID, time.Now().UTC())
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
	return r.GetLicence(ctx, name)
}

// GetLicence returns the licence addressed by its resource name.
func (r *PropertyRepository) GetLicence(ctx context.Context, name string) (*propertypbv1.Licence, error) {
	id, err := types.LicenceID(name)
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
	return licenceFromParts(res, att), nil
}

// ListLicences returns a page of licences under parent
// ("properties/{property}") — property-wide and per-unit ones alike; the
// filter narrows by target, unit, type, state, or expiry_date.
func (r *PropertyRepository) ListLicences(ctx context.Context, parent string, in repox.ListInput) ([]*propertypbv1.Licence, string, error) {
	propertyID, err := types.PropertyID(parent)
	if err != nil {
		return nil, "", err
	}
	fin, err := types.FilterxFromRaw(in)
	if err != nil {
		return nil, "", err
	}
	rows, next, err := filterx.Hasura[pschema.PropertyLicences](property.LicenceFilterSpec, r.svc.Query.Property.Licences).
		Scope(licencesql.PropertyId.Eq(propertyID)).
		List(ctx, fin)
	if err != nil {
		return nil, "", mapHasuraErr(types.MapFilterxErr(err))
	}
	items := make([]*propertypbv1.Licence, 0, len(rows))
	for i := range rows {
		att, err := r.attachment(ctx, rows[i].AttachmentId)
		if err != nil {
			return nil, "", err
		}
		items = append(items, licenceFromParts(&rows[i], att))
	}
	return items, next, nil
}

// UpdateLicence applies the masked fields of l to the stored licence. The
// target and unit are immutable; the attachment is rebuilt from the merged
// proto, and the superseded attachment row is deleted once unreferenced.
func (r *PropertyRepository) UpdateLicence(ctx context.Context, l *propertypbv1.Licence, paths []string) (*propertypbv1.Licence, error) {
	id, err := types.LicenceID(l.GetName())
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

	merged := licenceFromParts(res, att)
	applyLicenceMask(merged, l, paths)
	now := time.Now().UTC()
	var unitID *string
	if res.Unit != nil && *res.Unit != "" {
		unitID = res.Unit
	}
	g := buildLicenceGraph(merged, res.PropertyId, unitID, now)

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
	return r.GetLicence(ctx, l.GetName())
}

// DeleteLicence removes a licence and its attachment row.
func (r *PropertyRepository) DeleteLicence(ctx context.Context, name string) error {
	id, err := types.LicenceID(name)
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
