package hasura

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/oh-tarnished/freebusy/internal/database/repository/repox"
	"github.com/oh-tarnished/freebusy/internal/service/dbutil"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/licencesql"
	pschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/propertyql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/attachmentsql"
	sharedschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/sharedql/schemaql"
	"github.com/oh-tarnished/freebusy/internal/types"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/property/v1/propertypbv1"
	sharedpbv1 "github.com/oh-tarnished/freebusy/protobuf/generated/go/shared/v1/sharedpbv1"
	"github.com/oh-tarnished/runtime-go/ulid"
	"github.com/the-protobuf-project/runtime-go/network/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Conversions between the protobuf Licence domain type and the Hasura schema.
// The attachment (scanned certificate) is normalized into shared.attachments
// and referenced by FK; its bytea `content` crosses the GraphQL boundary as a
// base64 JSON string (ndc-postgres's bytea representation).

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
		UploadTime: dbutil.TsToStr(timestamppb.New(now)),
	}
}

func attachmentFromModel(a *sharedschema.SharedAttachments) *sharedpbv1.Attachment {
	if a == nil {
		return nil
	}
	return &sharedpbv1.Attachment{
		Filename:   repox.Deref(a.Filename),
		MimeType:   repox.Deref(a.MimeType),
		SizeBytes:  int64(repox.Deref(a.SizeBytes)),
		Content:    rawToBytes(a.Content),
		Uri:        repox.Deref(a.Uri),
		UploadTime: strToTS(repox.Deref(a.UploadTime)),
	}
}

// --- Licence graph -----------------------------------------------------------

func licenceTargetFromStr(s *string) propertypbv1.LicenceTarget {
	if s == nil || *s == "" {
		return propertypbv1.LicenceTarget_LICENCE_TARGET_UNSPECIFIED
	}
	return propertypbv1.LicenceTarget(propertypbv1.LicenceTarget_value["LICENCE_TARGET_"+*s])
}

type licenceGraph struct {
	licence    licencesql.CreateInput
	attachment *attachmentsql.CreateInput
}

// buildLicenceGraph materializes the proto into its row inputs. The target
// derives from whether unitID is set. Identity (id, name), state, and etag
// stay with the caller.
func buildLicenceGraph(l *propertypbv1.Licence, propertyID string, unitID *string, now time.Time) *licenceGraph {
	nowStr := dbutil.TsToStr(timestamppb.New(now))
	target := "PROPERTY"
	if unitID != nil {
		target = "UNIT"
	}
	g := &licenceGraph{
		licence: licencesql.CreateInput{
			Target:           target,
			Unit:             repox.Deref(unitID),
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

// licenceFromParts re-hydrates the proto from a licence row and its resolved
// attachment (nil when the licence has none). The unit resource name is
// rebuilt from the stored parent property and bare unit ids.
func licenceFromParts(res *pschema.PropertyLicences, att *sharedschema.SharedAttachments) *propertypbv1.Licence {
	var unit string
	if res.Unit != nil && *res.Unit != "" {
		unit, _ = types.UnitName(res.PropertyId, *res.Unit)
	}
	return &propertypbv1.Licence{
		Name:             res.Name,
		Target:           licenceTargetFromStr(res.Target),
		Unit:             unit,
		Type:             licenceTypeFromStr(res.Type),
		LicenceNumber:    repox.Deref(res.LicenceNumber),
		IssuingAuthority: repox.Deref(res.IssuingAuthority),
		IssueDate:        strToDate(repox.Deref(res.IssueDate)),
		ExpiryDate:       strToDate(repox.Deref(res.ExpiryDate)),
		Attachment:       attachmentFromModel(att),
		Notes:            repox.Deref(res.Notes),
		State:            licenceStateFromStr(res.State),
		CreateTime:       strToTS(res.CreateTime),
		UpdateTime:       strToTS(res.UpdateTime),
		Etag:             repox.Deref(res.Etag),
	}
}
