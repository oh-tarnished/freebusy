package hasura

import (
	"context"
	"fmt"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"time"

	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/propertiesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitsql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"github.com/the-protobuf-project/runtime-go/network/runtime"
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
		var out commonschema.InsertCommonPostalAddressResponse
		tx.Add(r.svc.Mutation.Common.PostalAddress.CreateOp(*g.address, &out))
	}
	if g.policy != nil {
		var out pschema.InsertPropertyPoliciesResponse
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
	var updRes pschema.UpdatePropertyPropertiesByIdResponse
	tx.Add(r.svc.Mutation.Property.Properties.UpdateOp(id, patch, &updRes))

	for i := range g.medias {
		g.medias[i].PropertyId = id
		var out pschema.InsertPropertyMediasResponse
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

// UpdateUnit applies the masked fields of u and returns the result. The pricing
// children, media, and applicable-promo-code rows are rebuilt from the merged
// proto; the superseded rows and their Money/DateRange value-objects are deleted
// in the same batch.
func (r *PropertyRepository) UpdateUnit(ctx context.Context, u *propertypbv1.Unit, paths []string) (*propertypbv1.Unit, error) {
	id, err := types.UnitID(u.GetName())
	if err != nil {
		return nil, err
	}
	res, err := r.svc.Query.Property.Units.Get(ctx, id)
	if err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	if u.GetEtag() != "" && res.Etag != nil && u.GetEtag() != *res.Etag {
		return nil, types.ErrConflict
	}
	parts, old, err := r.fetchUnitParts(ctx, res)
	if err != nil {
		return nil, err
	}

	merged := unitFromParts(parts)
	applyUnitMask(merged, u, paths)
	now := time.Now().UTC()
	g := buildUnitGraph(merged, res.PropertyId, now)
	g.unit.Id = id

	tx := r.svc.Mutation.Tx()
	if g.price != nil {
		var out commonschema.InsertCommonMoneysResponse
		tx.Add(r.svc.Mutation.Common.Moneys.CreateOp(*g.price, &out))
	}
	patch := unitsql.UpdateInput{
		DisplayName:  graphql.Value(g.unit.DisplayName),
		Description:  dbutil.NullableStr(g.unit.Description),
		Type:         graphql.Value(g.unit.Type),
		Capacity:     graphql.Value(g.unit.Capacity),
		MaxOccupancy: graphql.Value(g.unit.MaxOccupancy),
		TimeZone:     graphql.Value(g.unit.TimeZone),
		PricingUnit:  dbutil.NullableStr(g.unit.PricingUnit),
		Duration:     dbutil.NullableStr(g.unit.Duration),
		Tags:         graphql.Value(g.unit.Tags),
		Attributes:   graphql.Value(g.unit.Attributes),
		PriceId:      dbutil.NullableStr(g.unit.PriceId),
		Etag:         graphql.Value(ulid.GenerateString()),
		UpdateTime:   graphql.Value(dbutil.TsToStr(timestamppb.New(now))),
	}
	var updRes pschema.UpdatePropertyUnitsByIdResponse
	tx.Add(r.svc.Mutation.Property.Units.UpdateOp(id, patch, &updRes))

	queueUnitChildInserts(tx, r, g, id)
	queueUnitChildDeletes(tx, r, old)

	if err := tx.Commit(ctx); err != nil {
		return nil, dbutil.MapHasuraErr(err)
	}
	return r.GetUnit(ctx, u.GetName())
}

// DeleteUnit removes a unit (its pricing/media/promo children cascade in the DB)
// DeleteUnit removes a unit, its pricing children, and the Money/DateRange
// value-objects those children referenced. Child licences block the delete
// unless force is set, in which case they (and their attachment rows) go too.
func (r *PropertyRepository) DeleteUnit(ctx context.Context, name string, force bool) error {
	id, err := types.UnitID(name)
	if err != nil {
		return err
	}
	res, err := r.svc.Query.Property.Units.Get(ctx, id)
	if err != nil {
		return dbutil.MapHasuraErr(err)
	}
	if res == nil {
		return types.ErrNotFound
	}
	_, refs, err := r.fetchUnitParts(ctx, res)
	if err != nil {
		return err
	}
	licences, err := r.svc.Query.Property.Licences.List(ctx, licencesql.List().Where(licencesql.Unit.Eq(id)))
	if err != nil {
		return dbutil.MapHasuraErr(err)
	}
	if len(licences) > 0 && !force {
		return fmt.Errorf("%w: unit has %d licences; set force to delete them too",
			types.ErrInvalidArgument, len(licences))
	}
	tx := r.svc.Mutation.Tx()
	for i := range licences {
		var out pschema.DeletePropertyLicencesByIdResponse
		tx.Add(r.svc.Mutation.Property.Licences.DeleteOp(licences[i].Id, &out))
		if aid := licences[i].AttachmentId; aid != nil {
			var aOut sharedschema.DeleteSharedAttachmentsByIdResponse
			tx.Add(r.svc.Mutation.Shared.Attachments.DeleteOp(*aid, &aOut))
		}
	}
	var delRes pschema.DeletePropertyUnitsByIdResponse
	tx.Add(r.svc.Mutation.Property.Units.DeleteOp(id, &delRes))
	queueValueObjectDeletes(tx, r, refs.moneyIDs, refs.dateIDs)
	return dbutil.MapHasuraErr(tx.Commit(ctx))
}

// queuePropertyChildDeletes appends deletes for a property's superseded media
// rows and its now-unreferenced address / policy rows.
func queuePropertyChildDeletes(tx *runtime.Tx, r *PropertyRepository, refs propertyRefs) {
	for _, mid := range refs.mediaIDs {
		var out pschema.DeletePropertyMediasByIdResponse
		tx.Add(r.svc.Mutation.Property.Medias.DeleteOp(mid, &out))
	}
	if refs.addressID != nil {
		var out commonschema.DeleteCommonPostalAddressByIdResponse
		tx.Add(r.svc.Mutation.Common.PostalAddress.DeleteOp(*refs.addressID, &out))
	}
	if refs.policyID != nil {
		var out pschema.DeletePropertyPoliciesByIdResponse
		tx.Add(r.svc.Mutation.Property.Policies.DeleteOp(*refs.policyID, &out))
	}
}

// queueUnitChildDeletes appends deletes for a unit's superseded pricing/media/
// promo child rows and the Money/DateRange value-objects they referenced.
func queueUnitChildDeletes(tx *runtime.Tx, r *PropertyRepository, refs unitRefs) {
	for _, jid := range refs.rateIDs {
		var out pschema.DeletePropertyRateOverridesByIdResponse
		tx.Add(r.svc.Mutation.Property.RateOverrides.DeleteOp(jid, &out))
	}
	for _, jid := range refs.losIDs {
		var out pschema.DeletePropertyLosDiscountsByIdResponse
		tx.Add(r.svc.Mutation.Property.LosDiscounts.DeleteOp(jid, &out))
	}
	for _, jid := range refs.feeIDs {
		var out pschema.DeletePropertyFeesByIdResponse
		tx.Add(r.svc.Mutation.Property.Fees.DeleteOp(jid, &out))
	}
	for _, jid := range refs.taxIDs {
		var out pschema.DeletePropertyTaxesByIdResponse
		tx.Add(r.svc.Mutation.Property.Taxes.DeleteOp(jid, &out))
	}
	for _, jid := range refs.mediaIDs {
		var out pschema.DeletePropertyUnitMediasByIdResponse
		tx.Add(r.svc.Mutation.Property.UnitMedias.DeleteOp(jid, &out))
	}
	for _, jid := range refs.promoIDs {
		var out pschema.DeletePropertyUnitApplicablePromoCodesByIdResponse
		tx.Add(r.svc.Mutation.Property.UnitApplicablePromoCodes.DeleteOp(jid, &out))
	}
	queueValueObjectDeletes(tx, r, refs.moneyIDs, refs.dateIDs)
}

func queueValueObjectDeletes(tx *runtime.Tx, r *PropertyRepository, moneyIDs, dateIDs []string) {
	for _, mid := range moneyIDs {
		var out commonschema.DeleteCommonMoneysByIdResponse
		tx.Add(r.svc.Mutation.Common.Moneys.DeleteOp(mid, &out))
	}
	for _, did := range dateIDs {
		var out sharedschema.DeleteSharedDateRangesByIdResponse
		tx.Add(r.svc.Mutation.Shared.DateRanges.DeleteOp(did, &out))
	}
}
