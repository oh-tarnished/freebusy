// Licence mutations: masked update with attachment replacement, and delete.
package hasura

import (
	"context"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
		return nil, dbutil.MapHasuraErr(err)
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
		LicenceNumber:    dbutil.NullableStr(g.licence.LicenceNumber),
		IssuingAuthority: dbutil.NullableStr(g.licence.IssuingAuthority),
		IssueDate:        dbutil.NullableStr(g.licence.IssueDate),
		ExpiryDate:       dbutil.NullableStr(g.licence.ExpiryDate),
		Notes:            dbutil.NullableStr(g.licence.Notes),
		AttachmentId:     dbutil.NullableStr(g.licence.AttachmentId),
		Etag:             graphql.Value(ulid.GenerateString()),
		UpdateTime:       graphql.Value(dbutil.TsToStr(timestamppb.New(now))),
	}
	var updRes pschema.UpdatePropertyLicencesByIdResponse
	tx.Add(r.svc.Mutation.Property.Licences.UpdateOp(id, patch, &updRes))
	if res.AttachmentId != nil {
		var out sharedschema.DeleteSharedAttachmentsByIdResponse
		tx.Add(r.svc.Mutation.Shared.Attachments.DeleteOp(*res.AttachmentId, &out))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, dbutil.MapHasuraErr(err)
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
		return dbutil.MapHasuraErr(err)
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
	return dbutil.MapHasuraErr(tx.Commit(ctx))
}
