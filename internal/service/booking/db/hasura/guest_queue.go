// Mutation-batch helpers inserting and deleting guest graphs.
package hasura

import (
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	identityschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/schemaql"
	"github.com/the-protobuf-project/runtime-go/network/runtime"
)

// queueGuestInserts appends each guest graph to tx in foreign-key order: the
// belongs-to sub-rows first, then the guest row that references them and the
// booking.
func queueGuestInserts(tx *runtime.Tx, r *BookingRepository, graphs []guestGraph) {
	for i := range graphs {
		g := &graphs[i]
		if g.idDoc != nil {
			var res identityschema.InsertIdentityIdDocumentsResponse
			tx.Add(r.svc.Mutation.Identity.IdDocuments.CreateOp(*g.idDoc, &res))
		}
		if g.foreigner != nil {
			var res identityschema.InsertIdentityForeignerDetailsResponse
			tx.Add(r.svc.Mutation.Identity.ForeignerDetails.CreateOp(*g.foreigner, &res))
		}
		if g.prefs != nil {
			var res identityschema.InsertIdentityGuestPreferencesResponse
			tx.Add(r.svc.Mutation.Identity.GuestPreferences.CreateOp(*g.prefs, &res))
		}
		if g.permanent != nil {
			var res commonschema.InsertCommonPostalAddressResponse
			tx.Add(r.svc.Mutation.Common.PostalAddress.CreateOp(*g.permanent, &res))
		}
		if g.local != nil {
			var res commonschema.InsertCommonPostalAddressResponse
			tx.Add(r.svc.Mutation.Common.PostalAddress.CreateOp(*g.local, &res))
		}
		var gRes identityschema.InsertIdentityGuestsResponse
		tx.Add(r.svc.Mutation.Identity.Guests.CreateOp(g.guest, &gRes))
	}
}

// queueGuestDeletes appends deletes for a booking's existing guest party: one
// predicate delete (a native mutation, delete_identity_guests_by_booking_id)
// removes every guest row on the booking — including rows a stale snapshot
// missed — then the snapshot's belongs-to sub-rows (ID documents, foreigner
// details, preferences, addresses) are deleted by id.
func queueGuestDeletes(tx *runtime.Tx, r *BookingRepository, bookingID string, guests []identityschema.IdentityGuests) {
	var delAll identityschema.DeleteIdentityGuestsByBookingIdResponse
	tx.Add(r.svc.Mutation.Identity.Guests.DeleteByBookingIdOp(bookingID, &delAll))
	for i := range guests {
		g := &guests[i]
		if g.IdDocumentId != nil {
			var res identityschema.DeleteIdentityIdDocumentsByIdResponse
			tx.Add(r.svc.Mutation.Identity.IdDocuments.DeleteOp(*g.IdDocumentId, &res))
		}
		if g.ForeignerId != nil {
			var res identityschema.DeleteIdentityForeignerDetailsByIdResponse
			tx.Add(r.svc.Mutation.Identity.ForeignerDetails.DeleteOp(*g.ForeignerId, &res))
		}
		if g.PreferencesId != nil {
			var res identityschema.DeleteIdentityGuestPreferencesByIdResponse
			tx.Add(r.svc.Mutation.Identity.GuestPreferences.DeleteOp(*g.PreferencesId, &res))
		}
		if g.PermanentAddressId != nil {
			var res commonschema.DeleteCommonPostalAddressByIdResponse
			tx.Add(r.svc.Mutation.Common.PostalAddress.DeleteOp(*g.PermanentAddressId, &res))
		}
		if g.LocalAddressId != nil {
			var res commonschema.DeleteCommonPostalAddressByIdResponse
			tx.Add(r.svc.Mutation.Common.PostalAddress.DeleteOp(*g.LocalAddressId, &res))
		}
	}
}
