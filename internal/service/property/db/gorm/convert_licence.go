package gorm

import (
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// Conversions between the protobuf Licence domain type and its GORM storage
// model. The generated converters carry the field mass; this file wires what
// the schema cannot know: the target derivation, the unit reference, and the
// attachment row's identity and server-set upload time.

// licenceGraph is a Licence row plus the attachment row it references, built
// together and persisted in one transaction.
type licenceGraph struct {
	licence    *property.Licence
	attachment *shared.Attachment
}

// buildLicenceGraph materializes the proto into its row graph. The target
// derives from whether unitID is set. Identity (id, name), state, timestamps,
// and etag stay with the caller.
func buildLicenceGraph(l *propertypbv1.Licence, propertyID string, unitID *string) *licenceGraph {
	target := property.LicenceTargetProperty
	if unitID != nil {
		target = property.LicenceTargetUnit
	}
	row := property.LicenceFromProto(l)
	row.Name = "" // identity is the repository's
	row.Etag = nil
	row.State = nil
	row.Target = &target
	row.UnitID = unitID
	row.PropertyID = propertyID
	g := &licenceGraph{licence: row}
	if a := shared.AttachmentFromProto(l.GetAttachment()); a != nil {
		a.ID = ulid.GenerateString()
		now := time.Now().UTC()
		a.UploadTime = &now // server-set at creation
		g.attachment = a
		g.licence.AttachmentID = &a.ID
	}
	return g
}

// licenceFromModel re-hydrates the proto from a preloaded model. The generated
// converter covers everything except the unit resource name, rebuilt from the
// stored parent property and bare unit ids.
func licenceFromModel(m *property.Licence) *propertypbv1.Licence {
	l := property.LicenceToProto(m)
	if m.UnitID != nil {
		l.Unit, _ = types.UnitName(m.PropertyID, *m.UnitID)
	}
	return l
}
