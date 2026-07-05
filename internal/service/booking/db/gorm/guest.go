package gorm

import (
	"context"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/booking"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/identity"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/booking/v1/bookingpbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/identity/v1/identitypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/postaladdress"
	"gorm.io/gorm"
)

// This file maps a booking's guest party and occupancy between the protobuf
// domain types and their storage models. Guests are stored in the identity.guests
// table (has-many by booking_id) with belongs-to sub-rows for the ID document,
// foreigner-registration details, preferences, and the permanent/local addresses.
// Occupancy is a booking-local belongs-to value.

// --- occupancy ---------------------------------------------------------------

func occupancyToModel(o *bookingpbv1.Occupancy) *booking.Occupancy {
	if o == nil {
		return nil
	}
	return &booking.Occupancy{
		ID:       ulid.GenerateString(),
		Adults:   ptr(o.GetAdults()),
		Children: ptr(o.GetChildren()),
		Infants:  ptr(o.GetInfants()),
	}
}

func occupancyFromModel(m *booking.Occupancy) *bookingpbv1.Occupancy {
	if m == nil {
		return nil
	}
	return &bookingpbv1.Occupancy{
		Adults:   deref(m.Adults),
		Children: deref(m.Children),
		Infants:  deref(m.Infants),
	}
}

// partySize is the headcount charged against occupancy: the adults+children of
// the explicit Occupancy when given, else the non-infant guests. Infants are not
// counted.
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

// guestGraph is the set of rows one Guest materializes into: the guest row plus
// its belongs-to sub-rows (created before it, since the guest holds their FKs).
type guestGraph struct {
	guest       *identity.Guest
	idDocument  *identity.IdDocument
	foreigner   *identity.ForeignerDetails
	preferences *identity.GuestPreferences
	permanent   *common.PostalAddress
	local       *common.PostalAddress
}

// buildGuestGraph turns a proto Guest into its row graph under bookingID.
func buildGuestGraph(g *identitypbv1.Guest, bookingID string) guestGraph {
	graph := guestGraph{
		guest: &identity.Guest{
			ID:          ulid.GenerateString(),
			BookingID:   bookingID,
			DisplayName: g.GetDisplayName(),
			Primary:     ptr(g.GetPrimary()),
			Gender:      genderToModel(g.GetGender()),
			BirthDate:   dateToTime(g.GetBirthDate()),
			AgeGroup:    ageGroupToModel(g.GetAgeGroup()),
			Nationality: strOrNil(g.GetNationality()),
			Email:       strOrNil(g.GetEmail()),
			PhoneNumber: strOrNil(g.GetPhoneNumber()),
		},
	}
	if d := g.GetIdDocument(); d != nil {
		graph.idDocument = &identity.IdDocument{
			ID:             ulid.GenerateString(),
			Type:           idDocTypeToModel(d.GetType()),
			Number:         d.GetNumber(),
			IssuingCountry: strOrNil(d.GetIssuingCountry()),
			IssuePlace:     strOrNil(d.GetIssuePlace()),
			IssueDate:      dateToTime(d.GetIssueDate()),
			ExpiryDate:     dateToTime(d.GetExpiryDate()),
		}
		graph.guest.IDDocumentID = &graph.idDocument.ID
	}
	if f := g.GetForeigner(); f != nil {
		graph.foreigner = &identity.ForeignerDetails{
			ID:              ulid.GenerateString(),
			VisaNumber:      strOrNil(f.GetVisaNumber()),
			VisaType:        strOrNil(f.GetVisaType()),
			VisaIssuePlace:  strOrNil(f.GetVisaIssuePlace()),
			VisaIssueDate:   dateToTime(f.GetVisaIssueDate()),
			VisaExpiryDate:  dateToTime(f.GetVisaExpiryDate()),
			ArrivalDate:     dateToTime(f.GetArrivalDate()),
			EntryPort:       strOrNil(f.GetEntryPort()),
			Origin:          strOrNil(f.GetOrigin()),
			NextDestination: strOrNil(f.GetNextDestination()),
			VisitPurpose:    strOrNil(f.GetVisitPurpose()),
		}
		graph.guest.ForeignerID = &graph.foreigner.ID
	}
	if p := g.GetPreferences(); p != nil {
		graph.preferences = &identity.GuestPreferences{
			ID:              ulid.GenerateString(),
			Smoking:         smokingToModel(p.GetSmoking()),
			Bed:             bedToModel(p.GetBed()),
			Dietary:         p.GetDietary(),
			Accessibility:   p.GetAccessibility(),
			FloorPreference: ptr(p.GetFloorPreference()),
			LoyaltyNumber:   strOrNil(p.GetLoyaltyNumber()),
			SpecialRequests: p.GetSpecialRequests(),
			Notes:           strOrNil(p.GetNotes()),
		}
		graph.guest.PreferencesID = &graph.preferences.ID
	}
	if a := addressToModel(g.GetPermanentAddress()); a != nil {
		graph.permanent = a
		graph.guest.PermanentAddressID = &a.ID
	}
	if a := addressToModel(g.GetLocalAddress()); a != nil {
		graph.local = a
		graph.guest.LocalAddressID = &a.ID
	}
	return graph
}

// guestFromModel assembles the protobuf Guest from a stored row and its preloaded
// sub-rows.
func guestFromModel(m *identity.Guest) *identitypbv1.Guest {
	out := &identitypbv1.Guest{
		DisplayName:      m.DisplayName,
		Primary:          deref(m.Primary),
		Gender:           genderFromModel(m.Gender),
		BirthDate:        timeToDate(m.BirthDate),
		AgeGroup:         ageGroupFromModel(m.AgeGroup),
		Nationality:      deref(m.Nationality),
		Email:            deref(m.Email),
		PhoneNumber:      deref(m.PhoneNumber),
		PermanentAddress: addressFromModel(m.PermanentAddress),
		LocalAddress:     addressFromModel(m.LocalAddress),
	}
	if d := m.IDDocument; d != nil {
		out.IdDocument = &identitypbv1.IdDocument{
			Type:           idDocTypeFromModel(d.Type),
			Number:         d.Number,
			IssuingCountry: deref(d.IssuingCountry),
			IssuePlace:     deref(d.IssuePlace),
			IssueDate:      timeToDate(d.IssueDate),
			ExpiryDate:     timeToDate(d.ExpiryDate),
		}
	}
	if f := m.Foreigner; f != nil {
		out.Foreigner = &identitypbv1.ForeignerDetails{
			VisaNumber:      deref(f.VisaNumber),
			VisaType:        deref(f.VisaType),
			VisaIssuePlace:  deref(f.VisaIssuePlace),
			VisaIssueDate:   timeToDate(f.VisaIssueDate),
			VisaExpiryDate:  timeToDate(f.VisaExpiryDate),
			ArrivalDate:     timeToDate(f.ArrivalDate),
			EntryPort:       deref(f.EntryPort),
			Origin:          deref(f.Origin),
			NextDestination: deref(f.NextDestination),
			VisitPurpose:    deref(f.VisitPurpose),
		}
	}
	if p := m.Preferences; p != nil {
		out.Preferences = &identitypbv1.GuestPreferences{
			Smoking:         smokingFromModel(p.Smoking),
			Bed:             bedFromModel(p.Bed),
			Dietary:         p.Dietary,
			Accessibility:   p.Accessibility,
			FloorPreference: deref(p.FloorPreference),
			LoyaltyNumber:   deref(p.LoyaltyNumber),
			SpecialRequests: p.SpecialRequests,
			Notes:           deref(p.Notes),
		}
	}
	return out
}

// --- value-object + enum + date converters -----------------------------------

func addressToModel(a *postaladdress.PostalAddress) *common.PostalAddress {
	if a == nil {
		return nil
	}
	return &common.PostalAddress{
		ID:                 ulid.GenerateString(),
		Revision:           ptr(a.GetRevision()),
		RegionCode:         strOrNil(a.GetRegionCode()),
		LanguageCode:       strOrNil(a.GetLanguageCode()),
		PostalCode:         strOrNil(a.GetPostalCode()),
		SortingCode:        strOrNil(a.GetSortingCode()),
		AdministrativeArea: strOrNil(a.GetAdministrativeArea()),
		Locality:           strOrNil(a.GetLocality()),
		Sublocality:        strOrNil(a.GetSublocality()),
		AddressLines:       a.GetAddressLines(),
		Recipients:         a.GetRecipients(),
		Organization:       strOrNil(a.GetOrganization()),
	}
}

func addressFromModel(a *common.PostalAddress) *postaladdress.PostalAddress {
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
		AddressLines:       a.AddressLines,
		Recipients:         a.Recipients,
		Organization:       deref(a.Organization),
	}
}

func dateToTime(d *date.Date) *time.Time {
	if d == nil || (d.GetYear() == 0 && d.GetMonth() == 0 && d.GetDay() == 0) {
		return nil
	}
	t := time.Date(int(d.GetYear()), time.Month(d.GetMonth()), int(d.GetDay()), 0, 0, 0, 0, time.UTC)
	return &t
}

func timeToDate(t *time.Time) *date.Date {
	if t == nil || t.IsZero() {
		return nil
	}
	return &date.Date{Year: int32(t.Year()), Month: int32(t.Month()), Day: int32(t.Day())}
}

func genderToModel(g identitypbv1.Gender) *identity.Gender {
	if g == identitypbv1.Gender_GENDER_UNSPECIFIED {
		return nil
	}
	v := identity.Gender(strings.TrimPrefix(g.String(), "GENDER_"))
	return &v
}

func genderFromModel(g *identity.Gender) identitypbv1.Gender {
	if g == nil {
		return identitypbv1.Gender_GENDER_UNSPECIFIED
	}
	return identitypbv1.Gender(identitypbv1.Gender_value["GENDER_"+string(*g)])
}

func ageGroupToModel(a identitypbv1.AgeGroup) *identity.AgeGroup {
	if a == identitypbv1.AgeGroup_AGE_GROUP_UNSPECIFIED {
		return nil
	}
	v := identity.AgeGroup(strings.TrimPrefix(a.String(), "AGE_GROUP_"))
	return &v
}

func ageGroupFromModel(a *identity.AgeGroup) identitypbv1.AgeGroup {
	if a == nil {
		return identitypbv1.AgeGroup_AGE_GROUP_UNSPECIFIED
	}
	return identitypbv1.AgeGroup(identitypbv1.AgeGroup_value["AGE_GROUP_"+string(*a)])
}

func idDocTypeToModel(t identitypbv1.IdDocumentType) identity.IdDocumentType {
	return identity.IdDocumentType(strings.TrimPrefix(t.String(), "ID_DOCUMENT_TYPE_"))
}

func idDocTypeFromModel(t identity.IdDocumentType) identitypbv1.IdDocumentType {
	return identitypbv1.IdDocumentType(identitypbv1.IdDocumentType_value["ID_DOCUMENT_TYPE_"+string(t)])
}

func smokingToModel(s identitypbv1.SmokingPreference) *identity.SmokingPreference {
	if s == identitypbv1.SmokingPreference_SMOKING_PREFERENCE_UNSPECIFIED {
		return nil
	}
	v := identity.SmokingPreference(strings.TrimPrefix(s.String(), "SMOKING_PREFERENCE_"))
	return &v
}

func smokingFromModel(s *identity.SmokingPreference) identitypbv1.SmokingPreference {
	if s == nil {
		return identitypbv1.SmokingPreference_SMOKING_PREFERENCE_UNSPECIFIED
	}
	return identitypbv1.SmokingPreference(identitypbv1.SmokingPreference_value["SMOKING_PREFERENCE_"+string(*s)])
}

func bedToModel(b identitypbv1.BedPreference) *identity.BedPreference {
	if b == identitypbv1.BedPreference_BED_PREFERENCE_UNSPECIFIED {
		return nil
	}
	v := identity.BedPreference(strings.TrimPrefix(b.String(), "BED_PREFERENCE_"))
	return &v
}

func bedFromModel(b *identity.BedPreference) identitypbv1.BedPreference {
	if b == nil {
		return identitypbv1.BedPreference_BED_PREFERENCE_UNSPECIFIED
	}
	return identitypbv1.BedPreference(identitypbv1.BedPreference_value["BED_PREFERENCE_"+string(*b)])
}

// --- persistence -------------------------------------------------------------

// persistGuests inserts each guest graph in foreign-key order: the belongs-to
// sub-rows (ID document, foreigner details, preferences, addresses) first, then
// the guest row that references them and the booking.
func persistGuests(ctx context.Context, tx *gorm.DB, graphs []guestGraph) error {
	addrs := common.NewPostalAddressStore(tx)
	for i := range graphs {
		g := &graphs[i]
		if g.idDocument != nil {
			if e := identity.NewIdDocumentStore(tx).Create(ctx, g.idDocument); e != nil {
				return e
			}
		}
		if g.foreigner != nil {
			if e := identity.NewForeignerDetailsStore(tx).Create(ctx, g.foreigner); e != nil {
				return e
			}
		}
		if g.preferences != nil {
			if e := identity.NewGuestPreferencesStore(tx).Create(ctx, g.preferences); e != nil {
				return e
			}
		}
		if g.permanent != nil {
			if e := addrs.Create(ctx, g.permanent); e != nil {
				return e
			}
		}
		if g.local != nil {
			if e := addrs.Create(ctx, g.local); e != nil {
				return e
			}
		}
		if e := identity.NewGuestStore(tx).Create(ctx, g.guest); e != nil {
			return e
		}
	}
	return nil
}

// loadGuests returns a booking's guest party, with each guest's sub-rows
// preloaded, ordered by id (ULIDs preserve insertion order).
func (r *BookingRepository) loadGuests(ctx context.Context, bookingID string) ([]*identitypbv1.Guest, error) {
	var models []identity.Guest
	if err := r.db.WithContext(ctx).
		Preload("IDDocument").
		Preload("Foreigner").
		Preload("Preferences").
		Preload("PermanentAddress").
		Preload("LocalAddress").
		Where("booking_id = ?", bookingID).
		Order("id").
		Find(&models).Error; err != nil {
		return nil, mapGormErr(err)
	}
	out := make([]*identitypbv1.Guest, 0, len(models))
	for i := range models {
		out = append(out, guestFromModel(&models[i]))
	}
	return out, nil
}
