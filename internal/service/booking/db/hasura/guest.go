package hasura

import (
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"strings"
	"time"

	occupanciesql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/occupanciesql"
	bookingschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/bookingql/schemaql"
	postaladdressql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/postaladdressql"
	commonschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/commonql/schemaql"
	foreignerdetailsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/foreignerdetailsql"
	guestpreferencesql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/guestpreferencesql"
	guestsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/guestsql"
	iddocumentsql "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/iddocumentsql"
	identityschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/identityql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/runtime"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/postaladdress"
)

const dateLayout = "2006-01-02"

// --- occupancy ---------------------------------------------------------------

func occupancyInput(o *bookingpbv1.Occupancy) *occupanciesql.CreateInput {
	if o == nil {
		return nil
	}
	return &occupanciesql.CreateInput{
		Id:       ulid.GenerateString(),
		Adults:   o.GetAdults(),
		Children: o.GetChildren(),
		Infants:  o.GetInfants(),
	}
}

func occupancyFromSchema(o *bookingschema.BookingOccupancies) *bookingpbv1.Occupancy {
	if o == nil {
		return nil
	}
	return &bookingpbv1.Occupancy{Adults: repox.Deref(o.Adults), Children: repox.Deref(o.Children), Infants: repox.Deref(o.Infants)}
}

// --- guest graph -------------------------------------------------------------

// guestGraph is the set of GraphQL insert inputs one Guest materializes into.
type guestGraph struct {
	guest     guestsql.CreateInput
	idDoc     *iddocumentsql.CreateInput
	foreigner *foreignerdetailsql.CreateInput
	prefs     *guestpreferencesql.CreateInput
	permanent *postaladdressql.CreateInput
	local     *postaladdressql.CreateInput
}

func buildGuestGraph(g *identitypbv1.Guest, bookingID string) guestGraph {
	graph := guestGraph{
		guest: guestsql.CreateInput{
			Id:          ulid.GenerateString(),
			BookingId:   bookingID,
			DisplayName: g.GetDisplayName(),
			Primary:     g.GetPrimary(),
			Gender:      bare(g.GetGender().String(), "GENDER_"),
			BirthDate:   dateToStr(g.GetBirthDate()),
			AgeGroup:    bare(g.GetAgeGroup().String(), "AGE_GROUP_"),
			Nationality: g.GetNationality(),
			Email:       g.GetEmail(),
			PhoneNumber: g.GetPhoneNumber(),
		},
	}
	if d := g.GetIdDocument(); d != nil {
		id := ulid.GenerateString()
		graph.idDoc = &iddocumentsql.CreateInput{
			Id:             id,
			Type:           bare(d.GetType().String(), "ID_DOCUMENT_TYPE_"),
			Number:         d.GetNumber(),
			IssuingCountry: d.GetIssuingCountry(),
			IssuePlace:     d.GetIssuePlace(),
			IssueDate:      dateToStr(d.GetIssueDate()),
			ExpiryDate:     dateToStr(d.GetExpiryDate()),
		}
		graph.guest.IdDocumentId = id
	}
	if f := g.GetForeigner(); f != nil {
		id := ulid.GenerateString()
		graph.foreigner = &foreignerdetailsql.CreateInput{
			Id:              id,
			VisaNumber:      f.GetVisaNumber(),
			VisaType:        f.GetVisaType(),
			VisaIssuePlace:  f.GetVisaIssuePlace(),
			VisaIssueDate:   dateToStr(f.GetVisaIssueDate()),
			VisaExpiryDate:  dateToStr(f.GetVisaExpiryDate()),
			ArrivalDate:     dateToStr(f.GetArrivalDate()),
			EntryPort:       f.GetEntryPort(),
			Origin:          f.GetOrigin(),
			NextDestination: f.GetNextDestination(),
			VisitPurpose:    f.GetVisitPurpose(),
		}
		graph.guest.ForeignerId = id
	}
	if p := g.GetPreferences(); p != nil {
		id := ulid.GenerateString()
		graph.prefs = &guestpreferencesql.CreateInput{
			Id:              id,
			Smoking:         bare(p.GetSmoking().String(), "SMOKING_PREFERENCE_"),
			Bed:             bare(p.GetBed().String(), "BED_PREFERENCE_"),
			Dietary:         toStrPtrs(p.GetDietary()),
			Accessibility:   toStrPtrs(p.GetAccessibility()),
			FloorPreference: p.GetFloorPreference(),
			LoyaltyNumber:   p.GetLoyaltyNumber(),
			SpecialRequests: toStrPtrs(p.GetSpecialRequests()),
			Notes:           p.GetNotes(),
		}
		graph.guest.PreferencesId = id
	}
	if a := addressInput(g.GetPermanentAddress()); a != nil {
		graph.permanent = a
		graph.guest.PermanentAddressId = a.Id
	}
	if a := addressInput(g.GetLocalAddress()); a != nil {
		graph.local = a
		graph.guest.LocalAddressId = a.Id
	}
	return graph
}

// buildGuestGraphs turns a proto guest party into its insert graphs under bookingID.
func buildGuestGraphs(guests []*identitypbv1.Guest, bookingID string) []guestGraph {
	graphs := make([]guestGraph, 0, len(guests))
	for _, g := range guests {
		graphs = append(graphs, buildGuestGraph(g, bookingID))
	}
	return graphs
}

// guestFromSchema hydrates a protobuf Guest from its stored rows.
func guestFromSchema(g *identityschema.IdentityGuests, doc *identityschema.IdentityIdDocuments, f *identityschema.IdentityForeignerDetails, p *identityschema.IdentityGuestPreferences, perm, loc *commonschema.CommonPostalAddress) *identitypbv1.Guest {
	out := &identitypbv1.Guest{
		DisplayName:      g.DisplayName,
		Primary:          repox.Deref(g.Primary),
		Gender:           genderFromStr(g.Gender),
		BirthDate:        strToDate(repox.Deref(g.BirthDate)),
		AgeGroup:         ageGroupFromStr(g.AgeGroup),
		Nationality:      repox.Deref(g.Nationality),
		Email:            repox.Deref(g.Email),
		PhoneNumber:      repox.Deref(g.PhoneNumber),
		PermanentAddress: addressFromSchema(perm),
		LocalAddress:     addressFromSchema(loc),
	}
	if doc != nil {
		out.IdDocument = &identitypbv1.IdDocument{
			Type:           idDocTypeFromStr(doc.Type),
			Number:         doc.Number,
			IssuingCountry: repox.Deref(doc.IssuingCountry),
			IssuePlace:     repox.Deref(doc.IssuePlace),
			IssueDate:      strToDate(repox.Deref(doc.IssueDate)),
			ExpiryDate:     strToDate(repox.Deref(doc.ExpiryDate)),
		}
	}
	if f != nil {
		out.Foreigner = &identitypbv1.ForeignerDetails{
			VisaNumber:      repox.Deref(f.VisaNumber),
			VisaType:        repox.Deref(f.VisaType),
			VisaIssuePlace:  repox.Deref(f.VisaIssuePlace),
			VisaIssueDate:   strToDate(repox.Deref(f.VisaIssueDate)),
			VisaExpiryDate:  strToDate(repox.Deref(f.VisaExpiryDate)),
			ArrivalDate:     strToDate(repox.Deref(f.ArrivalDate)),
			EntryPort:       repox.Deref(f.EntryPort),
			Origin:          repox.Deref(f.Origin),
			NextDestination: repox.Deref(f.NextDestination),
			VisitPurpose:    repox.Deref(f.VisitPurpose),
		}
	}
	if p != nil {
		out.Preferences = &identitypbv1.GuestPreferences{
			Smoking:         smokingFromStr(p.Smoking),
			Bed:             bedFromStr(p.Bed),
			Dietary:         fromStrPtrs(p.Dietary),
			Accessibility:   fromStrPtrs(p.Accessibility),
			FloorPreference: repox.Deref(p.FloorPreference),
			LoyaltyNumber:   repox.Deref(p.LoyaltyNumber),
			SpecialRequests: fromStrPtrs(p.SpecialRequests),
			Notes:           repox.Deref(p.Notes),
		}
	}
	return out
}

// --- postal address ----------------------------------------------------------

func addressInput(a *postaladdress.PostalAddress) *postaladdressql.CreateInput {
	if a == nil {
		return nil
	}
	return &postaladdressql.CreateInput{
		Id:                 ulid.GenerateString(),
		Revision:           a.GetRevision(),
		RegionCode:         a.GetRegionCode(),
		LanguageCode:       a.GetLanguageCode(),
		PostalCode:         a.GetPostalCode(),
		SortingCode:        a.GetSortingCode(),
		AdministrativeArea: a.GetAdministrativeArea(),
		Locality:           a.GetLocality(),
		Sublocality:        a.GetSublocality(),
		AddressLines:       toStrPtrs(a.GetAddressLines()),
		Recipients:         toStrPtrs(a.GetRecipients()),
		Organization:       a.GetOrganization(),
	}
}

func addressFromSchema(a *commonschema.CommonPostalAddress) *postaladdress.PostalAddress {
	if a == nil {
		return nil
	}
	return &postaladdress.PostalAddress{
		Revision:           repox.Deref(a.Revision),
		RegionCode:         repox.Deref(a.RegionCode),
		LanguageCode:       repox.Deref(a.LanguageCode),
		PostalCode:         repox.Deref(a.PostalCode),
		SortingCode:        repox.Deref(a.SortingCode),
		AdministrativeArea: repox.Deref(a.AdministrativeArea),
		Locality:           repox.Deref(a.Locality),
		Sublocality:        repox.Deref(a.Sublocality),
		AddressLines:       fromStrPtrs(a.AddressLines),
		Recipients:         fromStrPtrs(a.Recipients),
		Organization:       repox.Deref(a.Organization),
	}
}

// --- date / list / enum helpers ----------------------------------------------

func dateToStr(d *date.Date) string {
	if d == nil || (d.GetYear() == 0 && d.GetMonth() == 0 && d.GetDay() == 0) {
		return ""
	}
	return time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC).Format(dateLayout)
}

func strToDate(s string) *date.Date {
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func toStrPtrs(ss []string) []*string {
	if len(ss) == 0 {
		return nil
	}
	out := make([]*string, len(ss))
	for i := range ss {
		v := ss[i]
		out[i] = &v
	}
	return out
}

func fromStrPtrs(ps []*string) []string {
	if len(ps) == 0 {
		return nil
	}
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			out = append(out, *p)
		}
	}
	return out
}

// bare strips the enum prefix from a proto enum name, returning "" for the
// unspecified value (so the column is omitted).
func bare(protoName, prefix string) string {
	if protoName == "" || strings.HasSuffix(protoName, "UNSPECIFIED") {
		return ""
	}
	return strings.TrimPrefix(protoName, prefix)
}

func genderFromStr(s *string) identitypbv1.Gender {
	if s == nil || *s == "" {
		return identitypbv1.Gender_GENDER_UNSPECIFIED
	}
	return identitypbv1.Gender(identitypbv1.Gender_value["GENDER_"+*s])
}

func ageGroupFromStr(s *string) identitypbv1.AgeGroup {
	if s == nil || *s == "" {
		return identitypbv1.AgeGroup_AGE_GROUP_UNSPECIFIED
	}
	return identitypbv1.AgeGroup(identitypbv1.AgeGroup_value["AGE_GROUP_"+*s])
}

func idDocTypeFromStr(s string) identitypbv1.IdDocumentType {
	if s == "" {
		return identitypbv1.IdDocumentType_ID_DOCUMENT_TYPE_UNSPECIFIED
	}
	return identitypbv1.IdDocumentType(identitypbv1.IdDocumentType_value["ID_DOCUMENT_TYPE_"+s])
}

func smokingFromStr(s *string) identitypbv1.SmokingPreference {
	if s == nil || *s == "" {
		return identitypbv1.SmokingPreference_SMOKING_PREFERENCE_UNSPECIFIED
	}
	return identitypbv1.SmokingPreference(identitypbv1.SmokingPreference_value["SMOKING_PREFERENCE_"+*s])
}

func bedFromStr(s *string) identitypbv1.BedPreference {
	if s == nil || *s == "" {
		return identitypbv1.BedPreference_BED_PREFERENCE_UNSPECIFIED
	}
	return identitypbv1.BedPreference(identitypbv1.BedPreference_value["BED_PREFERENCE_"+*s])
}

// --- persistence -------------------------------------------------------------

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
