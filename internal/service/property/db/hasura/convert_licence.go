package hasura

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/unitlicencesql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/attachmentsql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Conversions between the protobuf PropertyLicence/UnitLicence domain types and
// the Hasura schema. The attachment (scanned certificate) is normalized into
// shared.attachments and referenced by FK; its bytea `content` crosses the
// GraphQL boundary as a base64 JSON string (ndc-postgres's bytea
// representation).

// --- enum <-> bare value-name conversions -----------------------------------

func licenceTypeToStr(t propertypbv1.LicenceType) string {
	return strings.TrimPrefix(t.String(), "LICENCE_TYPE_")
}

func licenceTypeFromStr(s string) propertypbv1.LicenceType {
	return propertypbv1.LicenceType(propertypbv1.LicenceType_value["LICENCE_TYPE_"+s])
}

func licenceStateFromStr(s *string) propertypbv1.LicenceState {
	if s == nil || *s == "" {
		return propertypbv1.LicenceState_LICENCE_STATE_UNSPECIFIED
	}
	return propertypbv1.LicenceState(propertypbv1.LicenceState_value["LICENCE_STATE_"+*s])
}

// --- attachment content (bytea) ----------------------------------------------

// bytesToRaw encodes raw file bytes as the Postgres hex-format JSON string
// ("\x<hex>") the bytea scalar expects — ndc-postgres passes the value through
// Postgres's bytea input syntax, so any other shape is stored as literal
// escape-format bytes. Nil in, nil out.
func bytesToRaw(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	out, _ := json.Marshal(`\x` + hex.EncodeToString(b))
	return out
}

// rawToBytes decodes the bytea scalar back into raw bytes: ndc-postgres emits
// the Postgres hex form ("\x<hex>"); a base64 fallback covers connectors that
// use the NDC-standard bytes representation instead.
func rawToBytes(r *json.RawMessage) []byte {
	if r == nil {
		return nil
	}
	var s string
	if err := json.Unmarshal(*r, &s); err != nil || s == "" {
		return nil
	}
	if strings.HasPrefix(s, `\x`) {
		if b, err := hex.DecodeString(s[2:]); err == nil {
			return b
		}
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}

// attachmentInput builds a fresh attachment row input from a proto Attachment,
// or nil. upload_time is server-set at creation.
func attachmentInput(a *sharedpbv1.Attachment, now time.Time) *attachmentsql.CreateInput {
	if a == nil {
		return nil
	}
	return &attachmentsql.CreateInput{
		Id:         ulid.GenerateString(),
		Filename:   a.GetFilename(),
		MimeType:   a.GetMimeType(),
		SizeBytes:  graphql.Int64(a.GetSizeBytes()),
		Content:    bytesToRaw(a.GetContent()),
		Uri:        a.GetUri(),
		UploadTime: tsToStr(timestamppb.New(now)),
	}
}

func attachmentFromModel(a *sharedschema.SharedAttachments) *sharedpbv1.Attachment {
	if a == nil {
		return nil
	}
	return &sharedpbv1.Attachment{
		Filename:   deref(a.Filename),
		MimeType:   deref(a.MimeType),
		SizeBytes:  int64(deref(a.SizeBytes)),
		Content:    rawToBytes(a.Content),
		Uri:        deref(a.Uri),
		UploadTime: strToTS(deref(a.UploadTime)),
	}
}

// --- PropertyLicence graph -----------------------------------------------------

type propertyLicenceGraph struct {
	licence    licencesql.CreateInput
	attachment *attachmentsql.CreateInput
}

// buildPropertyLicenceGraph materializes the proto into its row inputs.
// Identity (id, name), state, and etag stay with the caller.
func buildPropertyLicenceGraph(l *propertypbv1.PropertyLicence, propertyID string, now time.Time) *propertyLicenceGraph {
	nowStr := tsToStr(timestamppb.New(now))
	g := &propertyLicenceGraph{
		licence: licencesql.CreateInput{
			Type:             licenceTypeToStr(l.GetType()),
			LicenceNumber:    l.GetLicenceNumber(),
			IssuingAuthority: l.GetIssuingAuthority(),
			IssueDate:        dateToStr(l.GetIssueDate()),
			ExpiryDate:       dateToStr(l.GetExpiryDate()),
			Notes:            l.GetNotes(),
			PropertyId:       propertyID,
			CreateTime:       nowStr,
			UpdateTime:       nowStr,
		},
	}
	if a := attachmentInput(l.GetAttachment(), now); a != nil {
		g.attachment = a
		g.licence.AttachmentId = a.Id
	}
	return g
}

// propertyLicenceFromParts re-hydrates the proto from a licence row and its
// resolved attachment (nil when the licence has none).
func propertyLicenceFromParts(res *pschema.PropertyLicences, att *sharedschema.SharedAttachments) *propertypbv1.PropertyLicence {
	return &propertypbv1.PropertyLicence{
		Name:             res.Name,
		Type:             licenceTypeFromStr(res.Type),
		LicenceNumber:    deref(res.LicenceNumber),
		IssuingAuthority: deref(res.IssuingAuthority),
		IssueDate:        strToDate(deref(res.IssueDate)),
		ExpiryDate:       strToDate(deref(res.ExpiryDate)),
		Attachment:       attachmentFromModel(att),
		Notes:            deref(res.Notes),
		State:            licenceStateFromStr(res.State),
		CreateTime:       strToTS(res.CreateTime),
		UpdateTime:       strToTS(res.UpdateTime),
		Etag:             deref(res.Etag),
	}
}

// --- UnitLicence graph -----------------------------------------------------------

type unitLicenceGraph struct {
	licence    unitlicencesql.CreateInput
	attachment *attachmentsql.CreateInput
}

// buildUnitLicenceGraph materializes the proto into its row inputs.
func buildUnitLicenceGraph(l *propertypbv1.UnitLicence, propertyID, unitID string, now time.Time) *unitLicenceGraph {
	nowStr := tsToStr(timestamppb.New(now))
	g := &unitLicenceGraph{
		licence: unitlicencesql.CreateInput{
			Type:             licenceTypeToStr(l.GetType()),
			LicenceNumber:    l.GetLicenceNumber(),
			IssuingAuthority: l.GetIssuingAuthority(),
			IssueDate:        dateToStr(l.GetIssueDate()),
			ExpiryDate:       dateToStr(l.GetExpiryDate()),
			Notes:            l.GetNotes(),
			PropertyId:       propertyID,
			UnitId:           unitID,
			CreateTime:       nowStr,
			UpdateTime:       nowStr,
		},
	}
	if a := attachmentInput(l.GetAttachment(), now); a != nil {
		g.attachment = a
		g.licence.AttachmentId = a.Id
	}
	return g
}

// unitLicenceFromParts re-hydrates the proto from a licence row and its
// resolved attachment.
func unitLicenceFromParts(res *pschema.PropertyUnitLicences, att *sharedschema.SharedAttachments) *propertypbv1.UnitLicence {
	return &propertypbv1.UnitLicence{
		Name:             res.Name,
		Type:             licenceTypeFromStr(res.Type),
		LicenceNumber:    deref(res.LicenceNumber),
		IssuingAuthority: deref(res.IssuingAuthority),
		IssueDate:        strToDate(deref(res.IssueDate)),
		ExpiryDate:       strToDate(deref(res.ExpiryDate)),
		Attachment:       attachmentFromModel(att),
		Notes:            deref(res.Notes),
		State:            licenceStateFromStr(res.State),
		CreateTime:       strToTS(res.CreateTime),
		UpdateTime:       strToTS(res.UpdateTime),
		Etag:             deref(res.Etag),
	}
}
