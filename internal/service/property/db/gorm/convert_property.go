package gorm

import (
	"github.com/lib/pq"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// propertyGraph is the set of rows a single Property materializes into: the
// property row, its belongs-to address and policy (created before it), and its
// has-many media rows (created after it, since they carry the property_id FK).
type propertyGraph struct {
	property *property.Property
	address  *common.PostalAddress
	policy   *property.Policy
	medias   []*property.Media
}

// buildPropertyGraph turns a proto Property into its row graph, minting a fresh
// ULID for every child row. The property row's identity (ID/Name/Etag) and the
// media rows' PropertyID FK are stamped by the repository, which owns id
// assignment and the transaction.
func buildPropertyGraph(pc *propertypbv1.Property) *propertyGraph {
	g := &propertyGraph{}
	state := property.PropertyStateActive
	g.property = &property.Property{
		OrganisationID: lastSegment(pc.GetOrganisation()),
		DisplayName:    pc.GetDisplayName(),
		Description:    strOrNil(pc.GetDescription()),
		TimeZone:       pc.GetTimeZone(),
		Tags:           pq.StringArray(pc.GetTags()),
		Attributes:     structToJSON(pc.GetAttributes()),
		State:          &state,
	}
	if a := addressToModel(pc.GetAddress()); a != nil {
		g.address = a
		g.property.AddressID = &a.ID
	}
	if pol := policyToModel(pc.GetPolicy()); pol != nil {
		g.policy = pol
		g.property.PolicyID = &pol.ID
	}
	for _, m := range pc.GetMedia() {
		g.medias = append(g.medias, mediaToModel(m))
	}
	return g
}

// propertyFromModel assembles the protobuf Property from a stored row and its
// preloaded associations (address, policy, media, child units).
func propertyFromModel(m *property.Property) *propertypbv1.Property {
	p := &propertypbv1.Property{
		Name:         m.Name,
		Organisation: orgName(m.OrganisationID),
		DisplayName:  m.DisplayName,
		Description:  deref(m.Description),
		Address:      addressFromModel(m.Address),
		TimeZone:     m.TimeZone,
		Policy:       policyFromModel(m.Policy),
		Tags:         []string(m.Tags),
		Attributes:   jsonToStruct(m.Attributes),
		State:        propertyStateFromModel(m.State),
		CreateTime:   timeToTS(&m.CreateTime),
		UpdateTime:   timeToTS(&m.UpdateTime),
		Etag:         deref(m.Etag),
	}
	for i := range m.Medias {
		p.Media = append(p.Media, mediaFromModel(&m.Medias[i]))
	}
	for i := range m.Units {
		p.Units = append(p.Units, m.Units[i].Name)
	}
	return p
}

func policyToModel(p *propertypbv1.Policy) *property.Policy {
	if p == nil {
		return nil
	}
	return &property.Policy{
		ID:           ulid.GenerateString(),
		CheckinTime:  todToTime(p.GetCheckinTime()),
		CheckoutTime: todToTime(p.GetCheckoutTime()),
		HouseRules:   pq.StringArray(p.GetHouseRules()),
		Notes:        strOrNil(p.GetNotes()),
	}
}

func policyFromModel(p *property.Policy) *propertypbv1.Policy {
	if p == nil {
		return nil
	}
	return &propertypbv1.Policy{
		CheckinTime:  timeToTOD(p.CheckinTime),
		CheckoutTime: timeToTOD(p.CheckoutTime),
		HouseRules:   []string(p.HouseRules),
		Notes:        deref(p.Notes),
	}
}

func mediaToModel(m *propertypbv1.Media) *property.Media {
	return &property.Media{
		ID:          ulid.GenerateString(),
		URI:         m.GetUri(),
		Type:        mediaTypeToModel(m.GetType()),
		Title:       strOrNil(m.GetTitle()),
		Description: strOrNil(m.GetDescription()),
		MimeType:    strOrNil(m.GetMimeType()),
		SortOrder:   ptr(m.GetSortOrder()),
		Primary:     ptr(m.GetPrimary()),
	}
}

func mediaFromModel(m *property.Media) *propertypbv1.Media {
	return &propertypbv1.Media{
		Uri:         m.URI,
		Type:        mediaTypeFromModel(m.Type),
		Title:       deref(m.Title),
		Description: deref(m.Description),
		MimeType:    deref(m.MimeType),
		SortOrder:   deref(m.SortOrder),
		Primary:     deref(m.Primary),
	}
}

func propertyStateFromModel(s *property.PropertyState) propertypbv1.PropertyState {
	if s == nil {
		return propertypbv1.PropertyState_PROPERTY_STATE_UNSPECIFIED
	}
	return propertypbv1.PropertyState(propertypbv1.PropertyState_value["PROPERTY_STATE_"+string(*s)])
}
