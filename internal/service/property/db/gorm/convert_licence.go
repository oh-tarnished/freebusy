package gorm

import (
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/property"
	"github.com/oh-tarnished/freebusy/internal/database/gorm/freebusy/shared"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
)

// Conversions between the protobuf Licence domain type and its GORM storage
// model. The attachment (scanned certificate) is normalized into the
// shared.attachments table and referenced by FK, mirroring how a unit's price
// references common.moneys.

func licenceTypeToModel(t propertypbv1.LicenceType) property.LicenceType {
	return property.LicenceType(strings.TrimPrefix(t.String(), "LICENCE_TYPE_"))
}

func licenceTypeFromModel(t property.LicenceType) propertypbv1.LicenceType {
	return propertypbv1.LicenceType(propertypbv1.LicenceType_value["LICENCE_TYPE_"+string(t)])
}

func licenceStateFromModel(s *property.LicenceState) propertypbv1.LicenceState {
	if s == nil {
		return propertypbv1.LicenceState_LICENCE_STATE_UNSPECIFIED
	}
	return propertypbv1.LicenceState(propertypbv1.LicenceState_value["LICENCE_STATE_"+string(*s)])
}

// attachmentToModel builds a fresh shared.Attachment row from a proto
// Attachment, or nil. upload_time is server-set at creation.
func attachmentToModel(a *sharedpbv1.Attachment) *shared.Attachment {
	if a == nil {
		return nil
	}
	now := time.Now().UTC()
	return &shared.Attachment{
		ID:         ulid.GenerateString(),
		Filename:   strOrNil(a.GetFilename()),
		MimeType:   strOrNil(a.GetMimeType()),
		SizeBytes:  ptr(a.GetSizeBytes()),
		Content:    a.GetContent(),
		URI:        strOrNil(a.GetUri()),
		UploadTime: &now,
	}
}

func attachmentFromModel(a *shared.Attachment) *sharedpbv1.Attachment {
	if a == nil {
		return nil
	}
	return &sharedpbv1.Attachment{
		Filename:   deref(a.Filename),
		MimeType:   deref(a.MimeType),
		SizeBytes:  deref(a.SizeBytes),
		Content:    a.Content,
		Uri:        deref(a.URI),
		UploadTime: timeToTS(a.UploadTime),
	}
}

// dateOrNil maps a proto Date onto the nullable date column (nil when unset).
func dateOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func timePtrToDate(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// licenceTargetFromModel maps the stored target column onto the proto enum.
func licenceTargetFromModel(t *property.LicenceTarget) propertypbv1.LicenceTarget {
	if t == nil {
		return propertypbv1.LicenceTarget_LICENCE_TARGET_UNSPECIFIED
	}
	return propertypbv1.LicenceTarget(propertypbv1.LicenceTarget_value["LICENCE_TARGET_"+string(*t)])
}

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
	g := &licenceGraph{
		licence: &property.Licence{
			Target:           &target,
			UnitID:           unitID,
			Type:             licenceTypeToModel(l.GetType()),
			LicenceNumber:    strOrNil(l.GetLicenceNumber()),
			IssuingAuthority: strOrNil(l.GetIssuingAuthority()),
			IssueDate:        dateOrNil(dateToTime(l.GetIssueDate())),
			ExpiryDate:       dateOrNil(dateToTime(l.GetExpiryDate())),
			Notes:            strOrNil(l.GetNotes()),
			PropertyID:       propertyID,
		},
	}
	if a := attachmentToModel(l.GetAttachment()); a != nil {
		g.attachment = a
		g.licence.AttachmentID = &a.ID
	}
	return g
}

// licenceFromModel re-hydrates the proto from a preloaded model. The unit
// resource name is rebuilt from the stored parent property and bare unit ids.
func licenceFromModel(m *property.Licence) *propertypbv1.Licence {
	var unit string
	if m.UnitID != nil {
		unit, _ = types.UnitName(m.PropertyID, *m.UnitID)
	}
	return &propertypbv1.Licence{
		Name:             m.Name,
		Target:           licenceTargetFromModel(m.Target),
		Unit:             unit,
		Type:             licenceTypeFromModel(m.Type),
		LicenceNumber:    deref(m.LicenceNumber),
		IssuingAuthority: deref(m.IssuingAuthority),
		IssueDate:        timeToDate(timePtrToDate(m.IssueDate)),
		ExpiryDate:       timeToDate(timePtrToDate(m.ExpiryDate)),
		Attachment:       attachmentFromModel(m.Attachment),
		Notes:            deref(m.Notes),
		State:            licenceStateFromModel(m.State),
		CreateTime:       timeToTS(&m.CreateTime),
		UpdateTime:       timeToTS(&m.UpdateTime),
		Etag:             deref(m.Etag),
	}
}
