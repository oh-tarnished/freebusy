package hasura

import (
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
	"github.com/oh-tarnished/generateql/runtime/go/runtime"
	"github.com/oh-tarnished/runtime-go/ulid"
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
	return &bookingpbv1.Occupancy{Adults: deref(o.Adults), Children: deref(o.Children), Infants: deref(o.Infants)}
}

// partySize is the headcount charged against occupancy: adults+children of the
// explicit occupancy, else the non-infant guests.
func partySize(o *bookingpbv1.Occupancy, guests []*identitypbv1.Guest) int32 {
	if o != nil && o.GetAdults()+o.GetChildren() > 0 {
		return o.GetAdults() + o.GetChildren()
	}
	var n int32
	for _, g := range guests {
		if g.GetAgeGroup() != identitypbv1.AgeGroup_AGE_GROUP_INFANT {
			n++
		}
	}
	return n
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

// guestFromSchema hydrates a protobuf Guest from its stored rows.
func guestFromSchema(g *identityschema.IdentityGuests, doc *identityschema.IdentityIdDocuments, f *identityschema.IdentityForeignerDetails, p *identityschema.IdentityGuestPreferences, perm, loc *commonschema.CommonPostalAddress) *identitypbv1.Guest {
	out := &identitypbv1.Guest{
		DisplayName:      g.DisplayName,
		Primary:          deref(g.Primary),
		Gender:           genderFromStr(g.Gender),
		BirthDate:        strToDate(deref(g.BirthDate)),
		AgeGroup:         ageGroupFromStr(g.AgeGroup),
		Nationality:      deref(g.Nationality),
		Email:            deref(g.Email),
		PhoneNumber:      deref(g.PhoneNumber),
		PermanentAddress: addressFromSchema(perm),
		LocalAddress:     addressFromSchema(loc),
	}
	if doc != nil {
		out.IdDocument = &identitypbv1.IdDocument{
			Type:           idDocTypeFromStr(doc.Type),
			Number:         doc.Number,
			IssuingCountry: deref(doc.IssuingCountry),
			IssuePlace:     deref(doc.IssuePlace),
			IssueDate:      strToDate(deref(doc.IssueDate)),
			ExpiryDate:     strToDate(deref(doc.ExpiryDate)),
		}
	}
	if f != nil {
		out.Foreigner = &identitypbv1.ForeignerDetails{
			VisaNumber:      deref(f.VisaNumber),
			VisaType:        deref(f.VisaType),
			VisaIssuePlace:  deref(f.VisaIssuePlace),
			VisaIssueDate:   strToDate(deref(f.VisaIssueDate)),
			VisaExpiryDate:  strToDate(deref(f.VisaExpiryDate)),
			ArrivalDate:     strToDate(deref(f.ArrivalDate)),
			EntryPort:       deref(f.EntryPort),
			Origin:          deref(f.Origin),
			NextDestination: deref(f.NextDestination),
			VisitPurpose:    deref(f.VisitPurpose),
		}
	}
	if p != nil {
		out.Preferences = &identitypbv1.GuestPreferences{
			Smoking:         smokingFromStr(p.Smoking),
			Bed:             bedFromStr(p.Bed),
			Dietary:         fromStrPtrs(p.Dietary),
			Accessibility:   fromStrPtrs(p.Accessibility),
			FloorPreference: deref(p.FloorPreference),
			LoyaltyNumber:   deref(p.LoyaltyNumber),
			SpecialRequests: fromStrPtrs(p.SpecialRequests),
			Notes:           deref(p.Notes),
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
		Revision:           deref(a.Revision),
		RegionCode:         deref(a.RegionCode),
		LanguageCode:       deref(a.LanguageCode),
		PostalCode:         deref(a.PostalCode),
		SortingCode:        deref(a.SortingCode),
		AdministrativeArea: deref(a.AdministrativeArea),
		Locality:           deref(a.Locality),
		Sublocality:        deref(a.Sublocality),
		AddressLines:       fromStrPtrs(a.AddressLines),
		Recipients:         fromStrPtrs(a.Recipients),
		Organization:       deref(a.Organization),
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
