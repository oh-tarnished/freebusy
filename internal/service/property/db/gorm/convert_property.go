package gorm

import (
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/common"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
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
// ULID for every child row. The generated converters carry the field mass; this
// only wires identities and foreign keys. The property row's identity
// (ID/Name/Etag) and the media rows' PropertyID FK are stamped by the
// repository, which owns id assignment and the transaction.
func buildPropertyGraph(pc *propertypbv1.Property) *propertyGraph {
	g := &propertyGraph{}
	g.property = property.PropertyFromProto(pc)
	g.property.Name = "" // identity is the repository's
	g.property.Etag = nil
	g.property.OrganisationID = repox.LastSegment(pc.GetOrganisation())
	if a := common.PostalAddressFromProto(pc.GetAddress()); a != nil {
		a.ID = ulid.GenerateString()
		g.address = a
		g.property.AddressID = &a.ID
	}
	if pol := property.PolicyFromProto(pc.GetPolicy()); pol != nil {
		pol.ID = ulid.GenerateString()
		g.policy = pol
		g.property.PolicyID = &pol.ID
	}
	for _, m := range pc.GetMedia() {
		row := property.MediaFromProto(m)
		row.ID = ulid.GenerateString()
		g.medias = append(g.medias, row)
	}
	return g
}

// propertyFromModel assembles the protobuf Property from a stored row and its
// preloaded associations. The generated converter covers the flat fields and
// the belongs-to address/policy; the organisation reference, media rows, and
// child name lists are layered on.
func propertyFromModel(m *property.Property) *propertypbv1.Property {
	p := property.PropertyToProto(m)
	p.Organisation = orgName(m.OrganisationID)
	for i := range m.Medias {
		p.Media = append(p.Media, property.MediaToProto(&m.Medias[i]))
	}
	for i := range m.Units {
		p.Units = append(p.Units, m.Units[i].Name)
	}
	for i := range m.Licences {
		p.Licences = append(p.Licences, m.Licences[i].Name)
	}
	return p
}
