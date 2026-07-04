package hasura

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/membersql"
	"github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/resourceql"
	orgschema "github.com/oh-tarnished/freebusy/internal/database/hasura/freebusyql/organisationql/schemaql"
	"github.com/oh-tarnished/freebusy/protobuf/generated/go/organisation/v1/orgpbv1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file holds the pure conversions between the protobuf Organisation/Member
// domain types and the flat Hasura/GraphQL schema. Timestamps cross the boundary
// as RFC 3339 strings; enums as their bare value name; the settings jsonb as raw
// JSON.

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func tsToStr(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

func strToTS(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return timestamppb.New(t)
		}
	}
	return nil
}

func structToJSON(s *structpb.Struct) []byte {
	if s == nil {
		return nil
	}
	b, err := s.MarshalJSON()
	if err != nil {
		return nil
	}
	return b
}

func jsonBytes(r *json.RawMessage) []byte {
	if r == nil {
		return nil
	}
	return []byte(*r)
}

func jsonToStruct(b []byte) *structpb.Struct {
	if len(b) == 0 {
		return nil
	}
	s := &structpb.Struct{}
	if err := s.UnmarshalJSON(b); err != nil {
		return nil
	}
	return s
}

func userName(id *string) string {
	if id == nil || *id == "" {
		return ""
	}
	return "users/" + *id
}

// --- Organisation ------------------------------------------------------------

func orgToCreateInput(o *orgpbv1.Organisation, now time.Time) resourceql.CreateInput {
	nowStr := tsToStr(timestamppb.New(now))
	return resourceql.CreateInput{
		DisplayName:  o.GetDisplayName(),
		Slug:         o.GetSlug(),
		BillingEmail: o.GetBillingEmail(),
		Settings:     structToJSON(o.GetSettings()),
		State:        "ACTIVE",
		CreateTime:   nowStr,
		UpdateTime:   nowStr,
	}
}

func orgFromModel(m *orgschema.OrganisationResource) *orgpbv1.Organisation {
	return &orgpbv1.Organisation{
		Name:         m.Name,
		DisplayName:  m.DisplayName,
		Slug:         deref(m.Slug),
		BillingEmail: deref(m.BillingEmail),
		State:        orgStateFromStr(m.State),
		Settings:     jsonToStruct(jsonBytes(m.Settings)),
		MemberCount:  int64(deref(m.MemberCount)),
		CreateTime:   strToTS(m.CreateTime),
		UpdateTime:   strToTS(m.UpdateTime),
		Etag:         deref(m.Etag),
	}
}

func orgStateFromStr(s *string) orgpbv1.OrganisationState {
	if s == nil || *s == "" {
		return orgpbv1.OrganisationState_ORGANISATION_STATE_UNSPECIFIED
	}
	return orgpbv1.OrganisationState(orgpbv1.OrganisationState_value["ORGANISATION_STATE_"+*s])
}

// --- Member ------------------------------------------------------------------

func memberToCreateInput(mem *orgpbv1.Member, now time.Time) membersql.CreateInput {
	nowStr := tsToStr(timestamppb.New(now))
	return membersql.CreateInput{
		Email:      mem.GetEmail(),
		Role:       roleToStr(mem.GetRole()),
		State:      "INVITED",
		CreateTime: nowStr,
		UpdateTime: nowStr,
	}
}

func memberFromModel(m *orgschema.OrganisationMembers) *orgpbv1.Member {
	return &orgpbv1.Member{
		Name:        m.Name,
		User:        userName(m.User),
		Email:       m.Email,
		DisplayName: deref(m.DisplayName),
		Role:        roleFromStr(m.Role),
		State:       memberStateFromStr(m.State),
		Inviter:     userName(m.Inviter),
		CreateTime:  strToTS(m.CreateTime),
		UpdateTime:  strToTS(m.UpdateTime),
		Etag:        deref(m.Etag),
	}
}

func roleToStr(r orgpbv1.OrganisationRole) string {
	return strings.TrimPrefix(r.String(), "ORGANISATION_ROLE_")
}

func roleFromStr(s string) orgpbv1.OrganisationRole {
	return orgpbv1.OrganisationRole(orgpbv1.OrganisationRole_value["ORGANISATION_ROLE_"+s])
}

func memberStateFromStr(s *string) orgpbv1.MemberState {
	if s == nil || *s == "" {
		return orgpbv1.MemberState_MEMBER_STATE_UNSPECIFIED
	}
	return orgpbv1.MemberState(orgpbv1.MemberState_value["MEMBER_STATE_"+*s])
}
